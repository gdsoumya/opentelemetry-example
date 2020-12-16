package main

import (
	"context"
	"encoding/json"
	"github.com/gdsoumya/opentelemetry-example/custom/simple-intrumentation/store"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"log"
	"net/http"
	"net/url"
	"time"
)

type SockData struct{
	TraceData store.TraceData `json:"trace-data"`
	OtherData string `json:"other-data"`
}

var tr = otel.Tracer("custom-sock-main")

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "custom-sock-client",
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

	makeSockConnection(ctx)
}

func makeSockConnection(ctx context.Context) {
	ctx = baggage.ContextWithValues(ctx, label.String("ProjectID", "1234"))
	ctx, span := tr.Start(ctx, "say hello", trace.WithAttributes(semconv.PeerServiceKey.String("ExampleService")))
	defer span.End()

	traceData := store.TraceData{}
	traceData.Inject(ctx)
	d, err:= json.Marshal(traceData)
	if err != nil {
		log.Fatal("marshal:", err)
	}
	req := http.Header{}
	req.Set("trace",string(d))
	log.Print(req,string(d))
	u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), req)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	go listenToServer(ctx, c)

	for i:=1;i<=2;i++ {
		sendDataToServer(ctx, c, "hello server "+string(rune(i)))
		time.Sleep(5*time.Second)
	}
	//time.Sleep(100*time.Second)
}

func listenToServer(ctx context.Context, c *websocket.Conn) {
	ctx, span := tr.Start(ctx, "listen to server")
	defer span.End()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read err:", err)
			return
		}
		sockData := SockData{}
		err = json.Unmarshal(message,&sockData)
		ctx1:=context.Background()
		ctx1=sockData.TraceData.Extract(ctx1)
		ctx1, span1 := tr.Start(ctx1, "received data from server")
		if err != nil {
			log.Println("err:", err)
			span1.End()
			continue
		}
		span1.SetAttributes(label.String("message", sockData.OtherData))
		log.Printf("recv: %s", message)
		span1.End()
	}
}

func sendDataToServer(ctx context.Context, c *websocket.Conn, data string){
	ctx, span := tr.Start(ctx, "talk to server")
	defer span.End()
	span.SetAttributes(label.String("message",data) )

	traceData := store.TraceData{}
	traceData.Inject(ctx)
	sockData := SockData{TraceData: traceData, OtherData: data}
	d, err := json.Marshal(sockData)
	if err != nil {
		panic(err)
	}
	err = c.WriteMessage(websocket.TextMessage, d)
	if err != nil {
		panic(err)
	}
}

