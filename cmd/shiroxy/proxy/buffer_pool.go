package proxy

import (
	"sync"
)

// DefaultBufferSize is the default size for buffers in the pool
const DefaultBufferSize = 32 * 1024 // 32KB

// SyncBufferPool implements BufferPool using sync.Pool for efficient buffer reuse
type SyncBufferPool struct {
	pool sync.Pool
	size int
}

// // NewSyncBufferPool creates a new buffer pool with the given size
func NewSyncBufferPool(size int) *SyncBufferPool {
	if size <= 0 {
		size = DefaultBufferSize
	}

	return &SyncBufferPool{
		size: size,
		pool: sync.Pool{
			New: func() interface{} {
				buf := make([]byte, size)
				return &buf
			},
		},
	}
}

// Get returns a buffer from the pool, or creates a new one if none are available
func (p *SyncBufferPool) Get() []byte {
	// Get buffer pointer from pool and dereference it
	return *p.pool.Get().(*[]byte)
}

// Put returns a buffer to the pool
func (p *SyncBufferPool) Put(buf []byte) {
	// Don't put buffers that are too small back in the pool
	if cap(buf) < p.size {
		return
	}

	// Reset the buffer to use its full capacity but zero length
	buf = buf[:cap(buf)]

	// Return buffer pointer to the pool
	p.pool.Put(&buf)
}
