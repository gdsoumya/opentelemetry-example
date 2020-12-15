package main

import (
	"bytes"
	"context"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var tr = otel.Tracer("main")

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
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return flush
}

func main() {
	flush := initTracer()
	defer flush()

	ctx, span := tr.Start(context.Background(), "client-main")
	defer span.End()
	payload:= `{"query":"mutation{createTodo(input:{text:\"hello\",userId:\"hui\"}){id text done user{id name}}}"}`

	data, err := sendMutation(ctx, "http://localhost:8080/query",[]byte(payload))
	if err!=nil{
		log.Print("ERROR ",err)
	}
	log.Print(data)
}

func sendMutation(ctx context.Context, server string, payload []byte) (string, error) {
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	ctx = baggage.ContextWithValues(ctx, label.String("ProjectID", "1234"))
	ctx, span := tr.Start(ctx, "say hello", trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
	defer span.End()

	req, err := http.NewRequestWithContext(ctx,"POST", server, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	return string(body), nil
}