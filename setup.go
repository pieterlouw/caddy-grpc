package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyhttp/httpserver"
)

func init() {
	caddy.RegisterPlugin("grpc", caddy.Plugin{
		ServerType: "http",
		Action:     setup,
	})
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
