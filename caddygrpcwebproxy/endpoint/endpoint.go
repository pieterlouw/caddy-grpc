package endpoint

import (
	"github.com/mholt/caddy"
	"github.com/pieterlouw/caddy-grpcwebproxy/caddygrpcwebproxy/grpcwebproxyserver"
)

func init() {
	caddy.RegisterPlugin("endpoint", caddy.Plugin{
		ServerType: "grpcwebproxy",
		Action:     setupEndpoint,
	})
}

func setupEndpoint(c *caddy.Controller) error {
	config := grpcwebproxyserver.GetConfig(c)

	// Ignore call to setupHost if the key is not echo or proxy
	/*if c.Key != "echo" && c.Key != "proxy" {
		return nil
	}
	*/
	for c.Next() {
		if !c.NextArg() {
			return c.ArgErr()
		}
		/*config.Hostname = c.Val()

		if config.TLS == nil {
			config.TLS = &caddytls.Config{Hostname: c.Val()}
		} else {
			config.TLS.Hostname = c.Val()
		}
		*/
		if c.NextArg() {
			// only one argument allowed
			return c.ArgErr()
		}
	}

	return nil
}
