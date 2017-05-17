# Caddy grpcwebproxy #

*Note: This server type plugin is  currently still a work in progress (WIP).*

grpcwebproxy server type for [Caddy Server](https://github.com/mholt/caddy)

The server type is meant to serve the same purpose as [grpcwebproxy](https://github.com/improbable-eng/grpc-web/tree/master/go/grpcwebproxy) by Improbable, but as a Caddy server type instead of a standalone Go application.
`grpcwebproxy` makes it possible for gRPC services to be consumed from browsers using the [gRPC-Web protocol](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md)


## Downloading

The plugin will be available to download as a server type from [Caddy's Download page](https://caddyserver.com/download).
Click on `Add plugins` option and scroll down to `Server Types` section where you can tick the `grpcwebproxy` box.

To verify the plugin is part of your downloaded instance of Caddy, run Caddy with the `-plugins` command line flag:
`caddy -plugins`

`grpcwebproxy` should be listed under `Server types` along with `http` and any other server types also included.

*Note, because the plugin is still in development it's not yet available on the Caddy build server*

## Running

To run the plugin, start Caddy with the `-type` command line flag:

`caddy -type=grpcwebproxy`  

## Roadmap/TODO 

Have similar features to standalone grpcwebproxy (structured logging, monitoring, endoint debug info for requests and connections , TLS serving, secure (plaintext) and TLS gRPC backend connectivity) leveraging `Caddyfile` and Caddy's builtin TLS.

## Proposed Caddyfile 

```
{
grpcweb.example.com 
endpoint localhost:9090
}
```

The first line `grpcweb.example.com` is the hostname/address of the site to serve.
The second line is a directive called `endpoint` where the backend gRPC service endpoint address can be specified.

This server type leverage the [tls directive](https://caddyserver.com/docs/tls) from the Caddy server and can be added to the server blocks as needed. 

## References ##

[Writing a Plugin: Server Type](https://github.com/mholt/caddy/wiki/Writing-a-Plugin:-Server-Type)
[Caddyfile](https://caddyserver.com/tutorial/caddyfile)
[grpcwebproxy](https://github.com/improbable-eng/grpc-web/tree/master/go/grpcwebproxy)
[gRPC-Web protocol](https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-WEB.md)
[gRPC-Web: Moving past REST+JSON towards type-safe Web APIs](https://spatialos.improbable.io/games/grpc-web-moving-past-restjson-towards-type-safe-web-apis)

