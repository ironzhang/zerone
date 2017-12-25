package rpc

import "context"

type keyTraceID struct{}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, keyTraceID{}, traceID)
}

func ParseTraceID(ctx context.Context) (string, bool) {
	value := ctx.Value(keyTraceID{})
	if traceID, ok := value.(string); ok {
		return traceID, true
	}
	return "", false
}

type keyVerbose struct{}

func WithVerbose(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, keyVerbose{}, verbose)
}

func ParseVerbose(ctx context.Context) (bool, bool) {
	value := ctx.Value(keyVerbose{})
	if verbose, ok := value.(bool); ok {
		return verbose, true
	}
	return false, false
}
