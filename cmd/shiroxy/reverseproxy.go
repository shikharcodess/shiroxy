package shiroxy

import (
	"context"
	"net/http"
	"time"
)

// A BufferPool is an interface for getting and returning temporary
// byte slices for use by io.CopyBuffer.
type BufferPool interface {
	Get() []byte
	Put([]byte)
}

type ReverseProxy struct {
	Context       context.Context
	Transport     http.RoundTripper
	Director      func(*http.Request)
	FlushInterval time.Duration
	// Todo: Implement Logger
	// BufferPool optionally specifies a buffer pool to
	// get byte slices for use by io.CopyBuffer when
	// copying HTTP response bodies.
	BufferPool BufferPool
}
