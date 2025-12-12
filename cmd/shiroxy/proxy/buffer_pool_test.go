package proxy

import (
	"testing"
)

func TestSyncBufferPool_GetPut(t *testing.T) {
	p := NewSyncBufferPool(1024)
	buf := p.Get()
	if cap(buf) < 1024 {
		t.Fatalf("buffer too small: %d", cap(buf))
	}
	// write something and put back
	buf = buf[:512]
	p.Put(buf)
	// Get again and ensure capacity is at least the configured size
	b2 := p.Get()
	if cap(b2) < 1024 {
		t.Fatalf("buffer too small on second get: %d", cap(b2))
	}
}
