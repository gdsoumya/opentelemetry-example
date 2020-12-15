module github.com/gdsoumya/opentelemetry-example/graphql-server

go 1.14

require (
	github.com/99designs/gqlgen v0.13.0
	github.com/vektah/gqlparser/v2 v2.1.0
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.15.1
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
)
