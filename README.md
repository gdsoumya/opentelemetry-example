# Open-Telemetry Go Examples

This repo contains examples for integrating and using open-telemetry with jaeger collector and golang.

## Setting up Jaeger:

1. Local Dev with Docker container
```
$ docker run -d --name jaeger \
  -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
  -p 5775:5775/udp \
  -p 6831:6831/udp \
  -p 6832:6832/udp \
  -p 5778:5778 \
  -p 16686:16686 \
  -p 14268:14268 \
  -p 14250:14250 \
  -p 9411:9411 \
  jaegertracing/all-in-one:1.21
```
- To access jaeger UI connect to `http://localhost:16686`. 
- The collector endpoint is `http://localhost:14268/api/traces`

## Run the examples

All the examples are paired in a `client` - `server` fashion for eg. we a normal client/server a gql client/server and a websock client/server. To the run the client and server you can either build each one and run them using `go build` or you can directly run the client/server pair using `go run ..` 

## Context-Propagation Using Plugins

Open-telemetry provides various context-propagation plugins that can integrate easily with pre-existing go packages like `otelhttp` that plugs into golang's native `net/http` package

All examples inside [plugin](https://github.com/gdsoumya/opentelemetry-example/tree/master/plugin) directory use either just open-telemetry package or pre-existing plugin integrations for gql and normal client/server applications.

## Custom Context-Propagation

It is also possible to create custom context propagation mechanism similar to the existing plugins, these usually need to inject and extract the context for the trace into the transport protocol like http or web-socket. 

The package [simple-instrumentation](https://github.com/gdsoumya/opentelemetry-example/tree/master/custom/simple-intrumentation) implements a custom context propagator that can be used as shown in the various examples in the [custom](https://github.com/gdsoumya/opentelemetry-example/tree/master/custom) directory to perform context propagation across process boundaries without using pre-existing plugins.
