# Open-Telemetry Go Examples

This repo contains examples for integrating and using open-telemetry with jaeger collector and golang.

## Context-Propagation Using Plugins

Open-telemetry provides various context-propagation plugins that can integrate easily with pre-existing go packages like `otelhttp` that plugs into golang's native `net/http` package

All examples inside `plugin` directory use either just open-telemetry package or pre-existing plugin integrations for gql and normal client/server applications.

## Custom Context-Propagation

It is also possible to create custom context propagation mechanism similar to the existing plugins, these usually need to inject and extract the context for the trace into the transport protocol like http or web-socket. 

The package `simple-instrumentation` implements a custom context propagator that can be used as shown in the various examples in the `custom` directory to perform context propagation across process boundaries without using pre-existing plugins.