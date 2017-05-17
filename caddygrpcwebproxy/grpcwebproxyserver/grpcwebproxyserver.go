package grpcwebproxyserver

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/mholt/caddy/caddytls"
	"github.com/mwitkow/grpc-proxy/proxy"
)

// Server is an implementation of the
// caddy.Server interface type
type Server struct {
	LocalAddr   string
	BackendAddr string
	listener    net.Listener
	listenerMu  sync.Mutex
	config      *Config
}

// NewServer returns a new grpcwebproxy server
func NewServer(l string, b string, c *Config) (*Server, error) {
	return &Server{
		LocalAddr:   l,
		BackendAddr: b,
		config:      c,
	}, nil
}

// Listen starts listening by creating a new listener
// and returning it. It does not start accepting
// connections.
func (s *Server) Listen() (net.Listener, error) {
	var listener net.Listener

	tlsConfig, err := caddytls.MakeTLSConfig([]*caddytls.Config{s.config.TLS})
	if err != nil {
		return nil, err
	}

	inner, err := net.Listen("tcp", fmt.Sprintf("%s", s.LocalAddr))
	if err != nil {
		return nil, err
	}

	//simulate buildListenerOrFail from https://github.com/improbable-eng/grpc-web/blob/master/go/grpcwebproxy/main.go
	/*	inner =  conntrack.NewListener(listener,
		conntrack.TrackWithName("http"), //should be a config setting
		conntrack.TrackWithTcpKeepAlive(20*time.Second),  //should be a config setting
		conntrack.TrackWithTracing(),
	*/

	if tlsConfig != nil {
		listener = tls.NewListener(inner, tlsConfig)
	} else {
		listener = inner
	}

	return listener, nil
}

// ListenPacket is a no-op for this server type
func (s *Server) ListenPacket() (net.PacketConn, error) {
	return nil, nil

}

// Serve starts serving using the provided listener.
// Serve blocks indefinitely, or in other
// words, until the server is stopped.
func (s *Server) Serve(ln net.Listener) error {

	s.listenerMu.Lock()
	s.listener = ln
	s.listenerMu.Unlock()

	// build Grpc Proxy Server
	//grpc.EnableTracing = true //should be a config setting
	//grpc_logrus.ReplaceGrpcLogger(logger)  //should be a config setting

	// gRPC proxy logic.
	//backendConn := dialBackendOrFail()

	//dial Backend
	opt := []grpc.DialOption{}
	opt = append(opt, grpc.WithCodec(proxy.Codec()))
	//if *flagBackendIsInsecure { // should be a config setting from endpoint directive
	opt = append(opt, grpc.WithInsecure())
	//} else {
	//	opt = append(opt, grpc.WithTransportCredentials(credentials.NewTLS(buildBackendTlsOrFail())))
	//

	backendConn, err := grpc.Dial(s.BackendAddr, opt...)
	if err != nil {
		return err
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

	// Debug server.
	servingServer := http.Server{
		//WriteTimeout: *flagHttpMaxWriteTimeout, //should be a config setting
		//ReadTimeout:  *flagHttpMaxReadTimeout, //should be a config setting
		Handler: http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			wrappedGrpc.ServeHTTP(resp, req)
		}),
	}

	return servingServer.Serve(s.listener)
}

// ServePacket is a no-op for this server type
func (s *Server) ServePacket(con net.PacketConn) error {

	return nil

}

// Stop stops s gracefully and closes its listener.
func (s *Server) Stop() error {

	err := s.listener.Close()
	if err != nil {
		return err
	}

	return nil
}

// OnStartupComplete lists the sites served by this server
// and any relevant information
func (s *Server) OnStartupComplete() {
	//if !caddy.Quiet {
	//	fmt.Println("[INFO] Proxying from ", s.LocalTCPAddr, " -> ", s.DestTCPAddr)
	//}
}
