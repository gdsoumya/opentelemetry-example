package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	"log"

	"github.com/gdsoumya/opentelemetry-example/graphql-server/graph/generated"
	"github.com/gdsoumya/opentelemetry-example/graphql-server/graph/model"
)

func (r *mutationResolver) CreateTodo(cxt context.Context, input model.NewTodo) (*model.Todo, error) {
	span := trace.SpanFromContext(cxt)
	span.SetAttributes(semconv.HTTPRouteKey.String("hello"))

	var tracer = otel.Tracer("server")
	cxt, span = tracer.Start(cxt, "server-span")
	defer span.End()

	projectID := baggage.Value(cxt, "ProjectID")
	log.Print(projectID.AsString())
	span.SetAttributes(label.KeyValue{Key: "ProjectID", Value: projectID})
	span.RecordError(errors.New("Error Test"))
	span.SetStatus(codes.Ok, "normal error")   // removes error status
	span.RecordError(errors.New("Error Test")) // new error adds error status
	span.AddEvent("writing response", trace.WithAttributes(label.String("content", "Hello World")))

	user := model.User{
		ID:   "12",
		Name: "Hello",
	}
	todo := model.Todo{
		ID:   "1",
		Text: "Hello",
		Done: false,
		User: &user,
	}
	return &todo, nil
}

func (r *queryResolver) Todos(ctx context.Context) ([]*model.Todo, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
