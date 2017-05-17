package grpcwebproxyserver

import (
	// plug in the server
	_ "github.com/pieterlouw/caddy-grpcwebproxy/grpcwebproxy/grpcwebproxyserver"
	// plug in the directives
	_ "github.com/pieterlouw/caddy-grpcwebproxy/grpcwebproxy/endpoint"
)
