package slog

import (
	"io"
	"io/ioutil"
	"sync"
)

// A WriteSyncer is an io.Writer that can also flush any buffered data. Note
// that *os.File (and thus, os.Stderr and os.Stdout) implement WriteSyncer.
type WriteSyncer interface {
	io.Writer
	Sync() error
}

type lockedWriteSyncer struct {
	sync.Mutex
	ws WriteSyncer
}

func NewLockedWriteSyncer(ws WriteSyncer) WriteSyncer {
	return &lockedWriteSyncer{ws: ws}
}

func (s *lockedWriteSyncer) Write(bs []byte) (int, error) {
	s.Lock()
	n, err := s.ws.Write(bs)
	s.Unlock()
	return n, err
}

func (s *lockedWriteSyncer) Sync() error {
	s.Lock()
	err := s.ws.Sync()
	s.Unlock()
	return err
}

// DiscardWrapper is used for testing.
var DiscardWrapper = &noSyncWrapper{ioutil.Discard}

type noSyncWrapper struct {
	io.Writer
}

func (n noSyncWrapper) Sync() error {
	return nil
}
