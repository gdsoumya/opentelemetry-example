package main

import (
	"context"
	"encoding/json"
	"github.com/gdsoumya/opentelemetry-example/custom/simple-intrumentation/store"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"log"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var tr = otel.Tracer("custom-main")
const traceHeader = "trace"

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "custom-client",
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

	makeRequest(ctx)
}

func makeRequest(ctx context.Context) {
	client := http.Client{}
	ctx = baggage.ContextWithValues(ctx, label.String("ProjectID", "1234"))

	ctx, span := tr.Start(ctx, "say hello", trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
	defer span.End()

	req, err := http.NewRequest("GET", "http://localhost:9000/hello", nil)
	if err != nil {
		panic(err)
	}
	if err = addTraceHeaders(ctx,req);err!=nil{
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

func addTraceHeaders(ctx context.Context, req *http.Request)(error){
	traceData := store.TraceData{}
	traceData.Inject(ctx)
	data, err:= json.Marshal(traceData)
	if err!=nil{
		return err
	}
	req.Header.Set(traceHeader,string(data))
	return nil
}