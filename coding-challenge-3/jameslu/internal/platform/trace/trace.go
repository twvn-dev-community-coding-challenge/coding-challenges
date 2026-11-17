// Package trace carries the causation id (the request or event that caused
// the current work) through a context, so every event and log line can point
// back to its cause.
package trace

import "context"

type causationKey struct{}

func WithCausation(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, causationKey{}, id)
}

func CausationID(ctx context.Context) string {
	id, _ := ctx.Value(causationKey{}).(string)
	return id
}
