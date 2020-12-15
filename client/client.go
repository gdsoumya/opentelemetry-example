package main

import (
	"context"
	"go.opentelemetry.io/otel/baggage"
	"io/ioutil"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "client",
			Tags: []label.KeyValue{
				label.String("exporter", "jaeger"),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		log.Fatal(err)
	}
	return flush
}

func main() {
	flush := initTracer()
	defer flush()
	var tr = otel.Tracer("main")
	ctx, span := tr.Start(context.Background(), "client-main")
	defer span.End()

	makeRequest(ctx)
}

func makeRequest(ctx context.Context) {
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	ctx = baggage.ContextWithValues(ctx, label.String("ProjectID", "1234"))
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:9000/hello", nil)
	if err != nil {
		panic(err)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	log.Print(string(body))
}
