package store

import (
	"context"
	"go.opentelemetry.io/otel"
)

type TraceData map[string]string

var propagator = otel.GetTextMapPropagator()

func (t TraceData) Get(key string) string{
	if val, ok := t[key];ok{
		return val
	}
	return ""
}

func (t TraceData) Set(key, value string){
	t[key] = value
}

func (t TraceData) Inject(ctx context.Context){
	propagator.Inject(ctx, t)
}

func (t TraceData) Extract(ctx context.Context) context.Context{
	return propagator.Extract(ctx, t)
}