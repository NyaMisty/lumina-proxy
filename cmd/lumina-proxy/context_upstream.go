package main

import (
	"context"
)

type upstreamContextKeyType struct{}

// A context key for upstream connection. The associated value will be of type
// *lumina.ClientSession.
var upstreamContextKey = upstreamContextKeyType{}

// Return *lumina.ClientSession extracted from a given context.Context.
func getUpstream(ctx context.Context) *ClientSessionHub {
	return ctx.Value(upstreamContextKey).(*ClientSessionHub)
}

// Create a lumina.ClientSession instance as upstream for each incoming connection.
func setUpstream(ctx context.Context, session *ClientSessionHub) context.Context {
	return context.WithValue(ctx, upstreamContextKey, session)
}
