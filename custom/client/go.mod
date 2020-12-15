module github.com/gdsoumya/opentelemetry-example/custom/client

go 1.14

replace github.com/gdsoumya/opentelemetry-example/custom/simple-intrumentation => "../simple-intrumentation"
require (
	go.opentelemetry.io/otel v0.15.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.15.0
	go.opentelemetry.io/otel/sdk v0.15.0
	github.com/gdsoumya/opentelemetry-example/custom/simple-intrumentation v0.0.1
)
