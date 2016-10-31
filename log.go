package slog

import (
	"io"
	"os"
	"sync"
	"time"
)

var (
	// Writer is the writer interface which the logs will be written too.
	Writer io.Writer = os.Stdout

	// TimeStampKey is the json key for the timestamp output.
	TimeStampKey = "ts"
)

var (
	mu sync.Mutex

	levelB = []byte("level")
	msgB   = []byte("msg")

	debugB = []byte("debug")
	infoB  = []byte("info")
	warnB  = []byte("warn")
	errorB = []byte("error")
	panicB = []byte("panic")
	fatalB = []byte("fatal")
)

func logMessage(l, msg []byte, fields []Field) {
	bp := bufPool.get()

	bp.WriteByte('{')

	appendKeyValue(bp, levelB, l)
	Time(TimeStampKey, time.Now()).append(bp)
	appendKeyValue(bp, msgB, msg)

	for _, f := range fields {
		f.append(bp)
	}

	bp.Truncate(bp.Len() - 1)
	bp.WriteByte('}')
	bp.WriteByte('\n')

	mu.Lock()
	bp.WriteTo(Writer)
	mu.Unlock()

	bufPool.put(bp)
}

func Debug(message string, fields ...Field) {
	logMessage(debugB, []byte(message), fields)
}

func Info(message string, fields ...Field) {
	logMessage(infoB, []byte(message), fields)
}

func Warn(message string, fields ...Field) {
	logMessage(warnB, []byte(message), fields)
}

func Error(message string, fields ...Field) {
	logMessage(errorB, []byte(message), fields)
}

func Panic(message string, fields ...Field) {
	logMessage(panicB, []byte(message), fields)
	panic(message)
}

func Fatal(message string, fields ...Field) {
	logMessage(fatalB, []byte(message), fields)
	os.Exit(1)
}
