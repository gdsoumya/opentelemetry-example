package main

import (
	"errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
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
			ServiceName: "server",
			Tags: []label.KeyValue{
				label.String("exporter", "jaeger"),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	if err != nil {
		log.Fatal(err)
	}
	return flush
}

func main() {
	flush := initTracer()
	defer flush()

	wrappedHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "/hello")
	http.Handle("/hello", wrappedHandler)
	http.ListenAndServe(":9000", nil)
}

func helloHandler(w http.ResponseWriter, req *http.Request) {
	time.Sleep(100 * time.Millisecond)
	cxt := req.Context()
	span := trace.SpanFromContext(cxt)
	span.SetAttributes(semconv.HTTPRouteKey.String("hello"))

	var tracer = otel.Tracer("server")
	cxt, span = tracer.Start(cxt, "server-span")
	defer span.End()
	time.Sleep(10 * time.Millisecond)
	projectID := baggage.Value(cxt, "ProjectID")
	log.Print("Propagated Baggage: ", projectID.AsString())
	span.SetAttributes(label.KeyValue{Key: "ProjectID", Value: projectID})
	span.RecordError(errors.New("Error Test"))
	span.SetStatus(codes.Ok, "normal error")   // removes error status
	span.RecordError(errors.New("Error Test")) // new error adds error status
	span.AddEvent("writing response", trace.WithAttributes(label.String("content", "Hello World")))

	time.Sleep(time.Second)

	w.Write([]byte("Hello World"))
}
