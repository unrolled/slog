package slog

import (
	"log"
	"os"
	"sync"

	"github.com/unrolled/slog/syslog"
)

// LockedSyslogWriteSyncer logs to syslog as well as stdout.
type LockedSyslogWriteSyncer struct {
	sync.Mutex
	w  *syslog.Writer
	ws WriteSyncer
}

// NewDatadogLockedSyslogWriteSyncer creates a new write syncer for datadog syslog.
func NewDatadogLockedSyslogWriteSyncer(network, address, tag, key string) WriteSyncer {
	w, err := syslog.Dial(network, address, syslog.LOG_INFO, tag)
	if err != nil {
		log.Fatalf("Failed to dial syslog: %s", err.Error())
	}

	syslog.SetDataDogKey(key)
	return &LockedSyslogWriteSyncer{w: w, ws: os.Stdout}
}

// NewLockedSyslogWriteSyncer creates a new write syncer for syslog.
func NewLockedSyslogWriteSyncer(network, address, tag string) WriteSyncer {
	w, err := syslog.Dial(network, address, syslog.LOG_DEBUG|syslog.LOG_LOCAL7, tag)
	if err != nil {
		log.Fatalf("Failed to dial syslog: %s", err.Error())
	}

	return &LockedSyslogWriteSyncer{w: w, ws: os.Stdout}
}

func (l *LockedSyslogWriteSyncer) Write(bs []byte) (int, error) {
	l.Lock()
	n, err := l.w.Write(bs)
	l.ws.Write(bs)
	l.Unlock()
	return n, err
}

func (l *LockedSyslogWriteSyncer) Sync() error {
	l.Lock()
	err := l.ws.Sync()
	l.Unlock()
	return err
}
