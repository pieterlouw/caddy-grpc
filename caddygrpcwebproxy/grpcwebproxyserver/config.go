package grpcwebproxyserver

import "github.com/mholt/caddy/caddytls"

// Config contains configuration details about a net server type
type Config struct {
	// The address of the site
	Addr Address

	// The hostname to bind listener to;
	// defaults to Addr.Host
	ListenHost string

	// TLS configuration
	TLS *caddytls.Config
}

// TLSConfig returns s.TLS.
func (c Config) TLSConfig() *caddytls.Config {
	return c.TLS
}

// Host returns s.Addr.Host.
func (c Config) Host() string {
	return c.Addr.Host
}

// Port returns s.Addr.Port.
func (c Config) Port() string {
	return c.Addr.Port
}
