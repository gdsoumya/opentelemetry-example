package main

import (
	"context"
	"encoding/json"
	"github.com/gdsoumya/opentelemetry-example/custom/simple-intrumentation/store"
	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"log"
	"net/http"
)

type SockData struct{
	TraceData store.TraceData `json:"trace-data"`
	OtherData string `json:"other-data"`
}

var upgrader = websocket.Upgrader{}
var tracer = otel.Tracer("custom-sock-server")

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline.
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "custom-sock-server",
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

	http.HandleFunc("/echo", listen)
	log.Fatal(http.ListenAndServe(":8000", nil))
}

func listen(w http.ResponseWriter, r *http.Request) {
	ctx, err := extractTraceContext(r)
	c, err := upgrader.Upgrade(w, r, nil)

	if err!=nil{
		log.Print("error getting ctx")
	}
	ctx, span := tracer.Start(ctx, "server-span")
	defer span.End()

	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		sockData := SockData{}
		err = json.Unmarshal(message,&sockData)
		ctx1:=context.Background()
		ctx1=sockData.TraceData.Extract(ctx1)
		ctx1, span1 := tracer.Start(ctx1, "received data from client")
		if err != nil {
			log.Println("err:", err)
			span1.End()
			continue
		}
		span1.SetAttributes(label.String("message", sockData.OtherData))
		log.Printf("recv: %s", message)
		span1.End()
		respond(ctx1,c,sockData.OtherData+"-Hi Client")
	}
}

func respond(ctx context.Context, c *websocket.Conn, data string){
	ctx, span := tracer.Start(ctx, "respond to client")
	defer span.End()
	span.SetAttributes(label.String("message",data) )

	traceData := store.TraceData{}
	traceData.Inject(ctx)
	sockData := SockData{TraceData: traceData, OtherData: data}
	d, err := json.Marshal(sockData)
	if err != nil {
		log.Print(err)
	}
	err = c.WriteMessage(websocket.TextMessage, d)
	if err != nil {
		log.Print(err)
	}
}

func extractTraceContext(req *http.Request) (context.Context, error){
	log.Print(req.Header)
	traceData := store.TraceData{}
	ctx := context.Background()

	err:= json.Unmarshal([]byte(req.Header.Get("trace")), &traceData)
	log.Print(traceData)
	if err!=nil{
		return ctx, err
	}
	return traceData.Extract(ctx), nil
}