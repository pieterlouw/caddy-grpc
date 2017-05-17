package grpcwebproxyserver

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/mholt/caddy"
	"github.com/mholt/caddy/caddyfile"
	"github.com/mholt/caddy/caddytls"
)

const serverType = "grpcwebproxy"

// directives for the grpcwebproxy server type
var directives = []string{"tls"}

func init() {
	caddy.RegisterServerType(serverType, caddy.ServerType{
		Directives: func() []string { return directives },
		DefaultInput: func() caddy.Input {
			return caddy.CaddyfileInput{
				ServerTypeName: serverType,
			}
		},
		NewContext: newContext,
	})

	caddy.RegisterParsingCallback(serverType, "tls", activateTLS)
	caddytls.RegisterConfigGetter(serverType, func(c *caddy.Controller) *caddytls.Config { return GetConfig(c).TLS })
}

func newContext() caddy.Context {
	return &grpcwebproxyContext{keysToConfigs: make(map[string]*Config)}
}

type grpcwebproxyContext struct {
	// keysToConfigs maps an address at the top of a
	// server block (a "key") to its Config. Not all
	// Configs will be represented here, only ones
	// that appeared in the Caddyfile.
	keysToConfigs map[string]*Config

	// configs is the master list of all site configs.
	configs []*Config
}

func (g *grpcwebproxyContext) saveConfig(key string, cfg *Config) {
	g.configs = append(g.configs, cfg)
	g.keysToConfigs[key] = cfg
}

// InspectServerBlocks make sure that everything checks out before
// executing directives and otherwise prepares the directives to
// be parsed and executed.
func (g *grpcwebproxyContext) InspectServerBlocks(sourceFile string, serverBlocks []caddyfile.ServerBlock) ([]caddyfile.ServerBlock, error) {
	// For each address in each server block, make a new config
	for _, sb := range serverBlocks {
		for _, key := range sb.Keys {
			key = strings.ToLower(key)
			if _, dup := g.keysToConfigs[key]; dup {
				return serverBlocks, fmt.Errorf("duplicate site address: %s", key)
			}

			addr, err := standardizeAddress(key)
			if err != nil {
				return serverBlocks, err
			}

			fmt.Printf("key: %s addr: %s\n", key, addr)

			// Fill in address components from command line so that middleware
			// have access to the correct information during setup
			/*if addr.Host == "" && Host != DefaultHost {
				addr.Host = Host
			}
			if addr.Port == "" && Port != DefaultPort {
				addr.Port = Port
			}*/

			// If default HTTP or HTTPS ports have been customized,
			// make sure the ACME challenge ports match
			/*var altHTTPPort, altTLSSNIPort string
			if HTTPPort != DefaultHTTPPort {
				altHTTPPort = HTTPPort
			}
			if HTTPSPort != DefaultHTTPSPort {
				altTLSSNIPort = HTTPSPort
			}*/

			// Save the config to our master list, and key it for lookups
			cfg := &Config{
				Addr: addr,
				TLS: &caddytls.Config{
					Hostname: addr.Host,
				},
			}
			g.saveConfig(key, cfg)

		}
	}

	return serverBlocks, nil
}

// MakeServers uses the newly-created configs to create and return a list of server instances.
func (g *grpcwebproxyContext) MakeServers() ([]caddy.Server, error) {
	//  create servers based on config type
	var servers []caddy.Server
	for _, cfg := range g.configs {
		s, err := NewServer(":8443", "localhost:9090", cfg) // hard coded for now, will read form Caddyfile
		if err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}

	return servers, nil
}

// GetConfig gets the Config that corresponds to c.
// If none exist (should only happen in tests), then a
// new, empty one will be created.
func GetConfig(c *caddy.Controller) *Config {
	ctx := c.Context().(*grpcwebproxyContext)
	key := strings.ToLower(c.Key)

	//only check for config if the value is proxy or echo
	//we need to do this because we specify the ports in the server block
	//and those values need to be ignored as they are also sent from caddy main process.
	if cfg, ok := ctx.keysToConfigs[key]; ok {
		return cfg
	}

	// we should only get here if value of key in server block
	// is not echo or proxy i.e port number :12017
	// we can't return a nil because caddytls.RegisterConfigGetter will panic
	// so we return a default (blank) config value
	return &Config{TLS: new(caddytls.Config)}
}

// Address represents a site address. It contains
// the original input value, and the component
// parts of an address. The component parts may be
// updated to the correct values as setup proceeds,
// but the original value should never be changed.
type Address struct {
	Original, Scheme, Host, Port, Path string
}

// String returns a human-friendly print of the address.
func (a Address) String() string {
	if a.Host == "" && a.Port == "" {
		return ""
	}
	scheme := a.Scheme
	if scheme == "" {
		if a.Port == "8443" {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	s := scheme
	if s != "" {
		s += "://"
	}
	s += a.Host
	if a.Port != "" { /*&&
		((scheme == "https" && a.Port != DefaultHTTPSPort) ||
			(scheme == "http" && a.Port != DefaultHTTPPort))*/
		s += ":" + a.Port
	}
	if a.Path != "" {
		s += a.Path
	}
	return s
}

// VHost returns a sensible concatenation of Host:Port/Path from a.
// It's basically the a.Original but without the scheme.
func (a Address) VHost() string {
	if idx := strings.Index(a.Original, "://"); idx > -1 {
		return a.Original[idx+3:]
	}
	return a.Original
}

// standardizeAddress parses an address string into a structured format with separate
// scheme, host, port, and path portions, as well as the original input string.
func standardizeAddress(str string) (Address, error) {
	input := str

	// Split input into components (prepend with // to assert host by default)
	if !strings.Contains(str, "//") && !strings.HasPrefix(str, "/") {
		str = "//" + str
	}
	u, err := url.Parse(str)
	if err != nil {
		return Address{}, err
	}

	// separate host and port
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		host, port, err = net.SplitHostPort(u.Host + ":")
		if err != nil {
			host = u.Host
		}
	}

	// see if we can set port based off scheme
	/*	if port == "" {
		if u.Scheme == "http" {
			port = HTTPPort
		} else if u.Scheme == "https" {
			port = HTTPSPort
		}
	}*/

	// repeated or conflicting scheme is confusing, so error
	if u.Scheme != "" && (port == "http" || port == "https") {
		return Address{}, fmt.Errorf("[%s] scheme specified twice in address", input)
	}

	// error if scheme and port combination violate convention
	//if (u.Scheme == "http" && port == HTTPSPort) || (u.Scheme == "https" && port == HTTPPort) {
	//	return Address{}, fmt.Errorf("[%s] scheme and port violate convention", input)
	//}

	// standardize http and https ports to their respective port numbers
	/*if port == "http" {
		u.Scheme = "http"
		port = HTTPPort
	} else if port == "https" {
		u.Scheme = "https"
		port = HTTPSPort
	}*/

	return Address{Original: input, Scheme: u.Scheme, Host: host, Port: port, Path: u.Path}, err
}
