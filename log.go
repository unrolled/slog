package slog

import (
	"os"
	"runtime"
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

const (
	SeverityDebug = "debug"
	SeverityInfo  = "info"
	SeverityWarn  = "warn"
	SeverityError = "error"
	SeverityPanic = "panic"
	SeverityFatal = "fatal"
)

var (
	mu           sync.Mutex
	globalFields = []Field{}

	debugB = []byte(SeverityDebug)
	infoB  = []byte(SeverityInfo)
	warnB  = []byte(SeverityWarn)
	errorB = []byte(SeverityError)
	panicB = []byte(SeverityPanic)
	fatalB = []byte(SeverityFatal)
	traceB = errorB
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
	_, _ = bp.WriteTo(Writer)
	mu.Unlock()

	bufPool.put(bp)
}

// LogFunc is the generic interface that the level funcs conform with.
type LogFunc func(message string, fields ...Field)

// Debug outputs a debug message. If `EnabledDebug` is false, this turns into a noop.
func Debug(message string, fields ...Field) {
	if EnableDebug {
		logMessage(debugB, []byte(message), fields)
	}
}

// Info outputs an info message.
func Info(message string, fields ...Field) {
	logMessage(infoB, []byte(message), fields)
}

// Warning outputs a warning message.
func Warning(message string, fields ...Field) {
	logMessage(warnB, []byte(message), fields)
}

// Error outputs an error message.
func Error(message string, fields ...Field) {
	logMessage(errorB, []byte(message), fields)
}

// Panic outputs a panic message and also calls `panic` with the original message.
func Panic(message string, fields ...Field) {
	logMessage(panicB, []byte(message), fields)
	_ = Writer.Sync()
	panic(message)
}

// Fatal outputs a fatal message and forces the application to exit with return code 1.
func Fatal(message string, fields ...Field) {
	logMessage(fatalB, []byte(message), fields)
	_ = Writer.Sync()
	os.Exit(1)
}

// TraceErr outputs the error with it's trace as an error log line, but also returns the original error.
func TraceErr(err error, fields ...Field) error {
	// If there is no error, do nothing!
	if err == nil {
		return err
	}

	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	traceFields := []Field{Err(err), String("file", frame.File), Int("line", frame.Line), String("func", frame.Function)}
	logMessage(traceB, []byte("trace"), append(traceFields, fields...))

	return err
}

// AddGlobalFields allows you to set fields that will automatically be appended to all messages.
func AddGlobalFields(fields ...Field) {
	mu.Lock()
	defer mu.Unlock()

	globalFields = append(globalFields, fields...)
}

// SetTraceErrSeverity allows you to change the severity type for trace errors (default is "error").
func SetTraceErrSeverity(s string) {
	traceB = []byte(s)
}
