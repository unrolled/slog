package slog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

type fieldType int

const (
	_ fieldType = iota
	boolType
	floatType
	intType
	int64Type
	uintType
	uint64Type
	uintptrType
	stringType
	jsonStringType
	errorType
	skipType
	rawType
)

type Field struct {
	key       string
	fieldType fieldType
	ival      int64
	str       string
	raw       []byte
	obj       interface{}
}

func Skip() Field {
	return Field{fieldType: skipType}
}

func Bool(key string, val bool) Field {
	var ival int64
	if val {
		ival = 1
	}

	return Field{key: key, fieldType: boolType, ival: ival}
}

func Float64(key string, val float64) Field {
	return Field{key: key, fieldType: floatType, ival: int64(math.Float64bits(val))}
}

func Int(key string, val int) Field {
	return Field{key: key, fieldType: intType, ival: int64(val)}
}

func Int64(key string, val int64) Field {
	return Field{key: key, fieldType: int64Type, ival: val}
}

func Uint(key string, val uint) Field {
	return Field{key: key, fieldType: uintType, ival: int64(val)}
}

func Uint64(key string, val uint64) Field {
	return Field{key: key, fieldType: uint64Type, ival: int64(val)}
}

func Uintptr(key string, val uintptr) Field {
	return Field{key: key, fieldType: uintptrType, ival: int64(val)}
}

func String(key string, val string) Field {
	return Field{key: key, fieldType: stringType, str: val}
}

func NullableString(key string, val string) Field {
	if val == "" {
		return Skip()
	}
	return Field{key: key, fieldType: stringType, str: val}
}

func JsonString(key string, val string) Field {
	return Field{key: key, fieldType: jsonStringType, str: val}
}

func RawJSON(key string, val []byte) Field {
	return Field{key: key, fieldType: rawType, raw: val}
}

func Jsonify(key string, val interface{}) Field {
	if val == nil {
		return Skip()
	}

	var final string
	result, err := json.Marshal(val)
	if err != nil {
		final = fmt.Sprintf("%#v || err(%s)", val, err.Error())
	} else {
		final = string(result)
	}
	return Field{key: key, fieldType: jsonStringType, str: final}
}

func Err(err error) Field {
	if err == nil {
		return Skip()
	}
	return Field{key: "err", fieldType: errorType, obj: err}
}

func Time(key string, val time.Time) Field {
	if len(TimeFormat) > 0 {
		return String(key, val.Format(TimeFormat))
	}

	return Int64(key, val.Unix())
}

func Duration(key string, val time.Duration) Field {
	return Int64(key, int64(val))
}

func Request(r *http.Request) Field {
	if token := r.Header.Get(RequestHeaderKey); token != "" {
		return String(RequestFieldKey, token)
	}

	return Skip()
}

func Raw(key string, val interface{}) Field {
	if out, err := json.Marshal(val); err == nil {
		return String(key, string(out))
	}

	return String(key, fmt.Sprintf("%#v", val))
}

func (f Field) appendField(b *bytes.Buffer) {
	switch f.fieldType {
	case boolType:
		appendBool(b, f.key, f.ival == 1)
	case floatType:
		appendFloat64(b, f.key, math.Float64frombits(uint64(f.ival)))
	case intType:
		appendInt(b, f.key, int(f.ival))
	case int64Type:
		appendInt64(b, f.key, f.ival)
	case uintType:
		appendUint(b, f.key, uint(f.ival))
	case uint64Type:
		appendUint64(b, f.key, uint64(f.ival))
	case uintptrType:
		appendUintptr(b, f.key, uintptr(f.ival))
	case stringType:
		appendString(b, f.key, f.str)
	case jsonStringType:
		appendJsonString(b, f.key, f.str)
	case errorType:
		appendString(b, f.key, f.obj.(error).Error())
	case skipType:
		break
	case rawType:
		appendRaw(b, f.key, f.raw)
	default:
		panic(fmt.Sprintf("unknown field type found: %v", f))
	}
}
