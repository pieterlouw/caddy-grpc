package grpcwebproxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
	"github.com/mwitkow/grpc-proxy/proxy"
)

func init() {
	caddy.RegisterPlugin("grpcwebproxy", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
}

type server struct {
	backendAddr       string
	next              httpserver.Handler
	backendIsInsecure bool
	backendTLS        *tls.Config
	wrappedGrpc       *grpcweb.WrappedGrpcServer
}

// ServeHTTP satisfies the httpserver.Handler interface.
func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	//dial Backend
	opt := []grpc.DialOption{}
	opt = append(opt, grpc.WithCodec(proxy.Codec()))
	if s.backendIsInsecure {
		opt = append(opt, grpc.WithInsecure())
	} else {
		opt = append(opt, grpc.WithTransportCredentials(credentials.NewTLS(s.backendTLS)))
	}

	backendConn, err := grpc.Dial(s.backendAddr, opt...)
	if err != nil {
		return s.next.ServeHTTP(w, r)
	}

	director := func(ctx context.Context, fullMethodName string) (*grpc.ClientConn, error) {
		return backendConn, nil
	}
	grpcServer := grpc.NewServer(
		grpc.CustomCodec(proxy.Codec()), // needed for proxy to function.
		grpc.UnknownServiceHandler(proxy.TransparentHandler(director)),
		/*grpc_middleware.WithUnaryServerChain(
			grpc_logrus.UnaryServerInterceptor(logger),
			grpc_prometheus.UnaryServerInterceptor,
		),
		grpc_middleware.WithStreamServerChain(
			grpc_logrus.StreamServerInterceptor(logger),
			grpc_prometheus.StreamServerInterceptor,
		),*/ //middleware should be a config setting or 3rd party middleware plugins like for caddyhttp
	)

	// gRPC-Web compatibility layer with CORS configured to accept on every
	wrappedGrpc := grpcweb.WrapServer(grpcServer, grpcweb.WithCorsForRegisteredEndpointsOnly(false))
	wrappedGrpc.ServeHTTP(w, r)

	return 0, nil
}

func (s server) Stop() {
	// TODO(pieterlouw): Add graceful shutdown.
	// Currently grpcweb.WrappedGrpcServer don't have a Stop/GracefulStop method
}

// setup configures a new server middleware instance.
func setup(c *caddy.Controller) error {
	for c.Next() {
		var s server

		if !c.Args(&s.backendAddr) { //loads next argument into backendAddr and fail if none specified
			return c.ArgErr()
		}

		tlsConfig := &tls.Config{}
		tlsConfig.MinVersion = tls.VersionTLS12

		s.backendTLS = tlsConfig
		s.backendIsInsecure = false

		//check for more settings in Caddyfile
		for c.NextBlock() {
			switch c.Val() {
			case "backend_is_insecure":
				s.backendIsInsecure = true
			case "backend_tls_noverify":
				s.backendTLS = buildBackendTLSNoVerify()
			case "backend_tls_ca_files":
				t, err := buildBackendTLSFromCAFiles(c.RemainingArgs())
				if err != nil {
					return err
				}
				s.backendTLS = t
			default:
				return c.Errf("unknown property '%s'", c.Val())
			}
		}

		httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
			s.next = next
			return s
		})

	}

	return nil
}

func buildBackendTLSNoVerify() *tls.Config {
	tlsConfig := &tls.Config{}
	tlsConfig.MinVersion = tls.VersionTLS12
	tlsConfig.InsecureSkipVerify = true

	return tlsConfig
}

func buildBackendTLSFromCAFiles(certs []string) (*tls.Config, error) {
	tlsConfig := &tls.Config{}
	tlsConfig.MinVersion = tls.VersionTLS12

	if len(certs) > 0 {
		tlsConfig.ClientCAs = x509.NewCertPool()
		for _, path := range certs {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return nil, fmt.Errorf("failed reading backend CA file %v: %v", path, err)
			}
			if ok := tlsConfig.ClientCAs.AppendCertsFromPEM(data); !ok {
				return nil, fmt.Errorf("failed processing backend CA file %v", path)
			}
		}
	}

	return tlsConfig, nil
}
