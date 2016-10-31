package slog

import (
	"bytes"
)

// bufPool represents a reusable buffer pool.
var bufPool *bufferPool

// bufferPool implements a pool of bytes.Buffers in the form of a bounded channel.
// Pulled from the github.com/oxtoacart/bpool package (Apache licensed).
type bufferPool struct {
	c chan *bytes.Buffer
}

// newBufferPool creates a new bufferPool bounded to the given size.
func newBufferPool(size int) (bp *bufferPool) {
	return &bufferPool{
		c: make(chan *bytes.Buffer, size),
	}
}

// get gets a Buffer from the bufferPool, or creates a new one if none are
// available in the pool.
func (bp *bufferPool) get() (b *bytes.Buffer) {
	select {
	case b = <-bp.c:
	// reuse existing buffer
	default:
		// create new buffer
		b = bytes.NewBuffer([]byte{})
	}
	return
}

// put returns the given Buffer to the bufferPool.
func (bp *bufferPool) put(b *bytes.Buffer) {
	b.Reset()
	select {
	case bp.c <- b:
	default: // Discard the buffer if the pool is full.
	}
}

// Initialize buffer pool.
func init() {
	bufPool = newBufferPool(64)
}
