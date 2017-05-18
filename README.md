# Caddy grpcwebproxy #

*Note: This plugin is currently still a work in progress (WIP).*

grpcwebproxy plugin for [Caddy Server](https://github.com/mholt/caddy)

The plugin is meant to serve the same purpose as [grpcwebproxy](https://github.com/improbable-eng/grpc-web/tree/master/go/grpcwebproxy) by Improbable, but as a Caddy server type instead of a standalone Go application.
`grpcwebproxy` makes it possible for gRPC services to be consumed from browsers using the [gRPC-Web protocol](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md)


## Downloading

The plugin will be available to download as a plugin from [Caddy's Download page](https://caddyserver.com/download).
Click on `Add plugins` option and scroll down to the section where you can tick the `http.grpcwebproxy` box.

To verify the plugin is part of your downloaded instance of Caddy, run Caddy with the `-plugins` command line flag:
`caddy -plugins`

`http.grpcwebproxy` should be listed under `Other plugins` along with `http` and any other plugins also included.

*Note, because the plugin is still in development it's not yet available on the Caddy download page* 

## Roadmap/TODO 

Have similar features to standalone grpcwebproxy (structured logging, monitoring, endoint debug info for requests and connections ,  secure (plaintext) and TLS gRPC backend connectivity)

## Proposed Caddyfile 

```
example.com 
grpcwebproxy localhost:9090
```

The first line `example.com` is the hostname/address of the site to serve.
The second line is a directive called `grpcwebproxy` where the backend gRPC service endpoint address (i.e `localhost:9090` as in the example) can be specified.


## References ##

[Extending Caddy](https://github.com/mholt/caddy/wiki/Extending-Caddy)

[Writing a Plugin: Directives](https://github.com/mholt/caddy/wiki/Writing-a-Plugin:-Directives)

[Caddyfile](https://caddyserver.com/tutorial/caddyfile)

[grpcwebproxy](https://github.com/improbable-eng/grpc-web/tree/master/go/grpcwebproxy)

[gRPC-Web protocol](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md)

[gRPC-Web: Moving past REST+JSON towards type-safe Web APIs](https://spatialos.improbable.io/games/grpc-web-moving-past-restjson-towards-type-safe-web-apis)


