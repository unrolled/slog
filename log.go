package slog

import (
	"os"
	"sync"
	"time"
)

var (
	// Writer is the writer interface which the logs will be written too.
	Writer = NewLockedWriteSyncer(os.Stdout)

	// TimeStampKey is the json key for the timestamp output.
	TimeStampKey = "ts"

	// TimeFormat will set the `slog.Time` output format if supplied. Defaults to `time.Unix()`.
	TimeFormat = ""

	// SeverityKey is the json key for the initial log type (info, warn, error, etc etc).
	SeverityKey = []byte("level")

	// TitleKey is the json key for the name of the log message.
	TitleKey = []byte("msg")

	// EnableDebug will print debug logs if true.
	EnableDebug = false

	// RequestToken is the token generator for the request middleware.
	RequestToken Token = &genericToken{}
)

var (
	mu           sync.Mutex
	globalFields = []Field{}

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

	appendKeyValue(bp, SeverityKey, l)
	appendKeyValue(bp, TitleKey, msg)

	// Start with the passed in fields.
	for _, f := range fields {
		f.appendField(bp)
	}

	// Add in the global fields last.
	for _, gf := range globalFields {
		gf.appendField(bp)
	}

	// Add the time at the end... most log services pick this up automatically anyway.
	Time(TimeStampKey, time.Now()).appendField(bp)

	bp.Truncate(bp.Len() - 2) // comma and space
	bp.WriteByte('}')
	bp.WriteByte('\n')

	mu.Lock()
	bp.WriteTo(Writer)
	mu.Unlock()

	bufPool.put(bp)
}

type LogFunc func(message string, fields ...Field)

func Debug(message string, fields ...Field) {
	if EnableDebug {
		logMessage(debugB, []byte(message), fields)
	}
}

func Info(message string, fields ...Field) {
	logMessage(infoB, []byte(message), fields)
}

func Warning(message string, fields ...Field) {
	logMessage(warnB, []byte(message), fields)
}

func Error(message string, fields ...Field) {
	logMessage(errorB, []byte(message), fields)
}

func Panic(message string, fields ...Field) {
	logMessage(panicB, []byte(message), fields)
	Writer.Sync()
	panic(message)
}

func Fatal(message string, fields ...Field) {
	logMessage(fatalB, []byte(message), fields)
	Writer.Sync()
	os.Exit(1)
}

// AddGlobalFields allows you to set fields that will automatically be appended to all messages.
func AddGlobalFields(fields ...Field) {
	mu.Lock()
	defer mu.Unlock()

	globalFields = append(globalFields, fields...)
}
