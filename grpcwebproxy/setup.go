package grpcwebproxy

import (
	"fmt"
	"net/http"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"

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
	next httpserver.Handler
}

// ServeHTTP satisfies the httpserver.Handler interface.
func (s server) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	//dial Backend
	opt := []grpc.DialOption{}
	opt = append(opt, grpc.WithCodec(proxy.Codec()))
	//if *flagBackendIsInsecure { // should be a config setting from endpoint directive
	opt = append(opt, grpc.WithInsecure())
	//} else {
	//	opt = append(opt, grpc.WithTransportCredentials(credentials.NewTLS(buildBackendTlsOrFail())))
	//

	backendConn, err := grpc.Dial("localhost:9090", opt...)
	if err != nil {
		return s.next.ServeHTTP(w, r)
	}

	director := func(ctx context.Context, fullMethodName string) (*grpc.ClientConn, error) {
		return backendConn, nil
	}
	// Server with logging and monitoring enabled.
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

// setup configures a new Proxy middleware instance.
func setup(c *caddy.Controller) error {
	/*upstreams, err := NewStaticUpstreams(c.Dispenser, httpserver.GetConfig(c).Host())
	if err != nil {
		return err
	}
	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return Proxy{Next: next, Upstreams: upstreams}
	})

	// Register shutdown handlers.
	for _, upstream := range upstreams {
		c.OnShutdown(upstream.Stop)
	}*/

	for c.Next() {
		fmt.Println("grpcwebproxy setup:", c.Val())
	}

	httpserver.GetConfig(c).AddMiddleware(func(next httpserver.Handler) httpserver.Handler {
		return server{next: next}
	})

	return nil
}
