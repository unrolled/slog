package slog

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

type logRecord struct {
	Level   string                 `json:"level"`
	Message string                 `json:"msg"`
	Time    time.Time              `json:"ts"`
	Fields  map[string]interface{} `json:"fields"`
}

func BenchmarkJSON(b *testing.B) {
	Writer = ioutil.Discard
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Debug(
				"fake",
				String("str", "foo"),
				Int("int", 1),
				Int64("int64", 1),
				Float64("float64", 1.0),
				String("string1", "\n"),
				String("string2", "💩"),
				String("string3", "🤔"),
				String("string4", "🙊"),
				Bool("bool", true),
			)
		}
	})
}

func BenchmarkStandardJSON(b *testing.B) {
	record := logRecord{
		Level:   "debug",
		Message: "fake",
		Time:    time.Unix(0, 0),
		Fields: map[string]interface{}{
			"str":     "foo",
			"int":     int(1),
			"int64":   int64(1),
			"float64": float64(1.0),
			"string1": "\n",
			"string2": "💩",
			"string3": "🤔",
			"string4": "🙊",
			"bool":    true,
		},
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			json.Marshal(record)
		}
	})
}

func BenchmarkStandardLog(b *testing.B) {
	stdlog := log.New(ioutil.Discard, "", 0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			stdlog.Printf("level:debug msg:fake time:%d str:%s int:%d int64:%d float64:%f string1:%s string2:%s string3:%s string4:%s bool:%v",
				time.Now(), "foo", 1, 1, 1.0, "\n", "💩", "🤔", "🙊", true)
		}
	})
}

func BenchmarkStandardCombo(b *testing.B) {
	stdlog := log.New(ioutil.Discard, "", 0)
	record := logRecord{
		Level:   "debug",
		Message: "fake",
		Time:    time.Unix(0, 0),
		Fields: map[string]interface{}{
			"str":     "foo",
			"int":     int(1),
			"int64":   int64(1),
			"float64": float64(1.0),
			"string1": "\n",
			"string2": "💩",
			"string3": "🤔",
			"string4": "🙊",
			"bool":    true,
		},
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			out, _ := json.Marshal(record)
			stdlog.Println(string(out))
		}
	})
}
