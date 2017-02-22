package slog

import (
	"bytes"
	"math"
	"strconv"
	"unicode/utf8"
)

const (
	_hex             = "0123456789abcdef"
	digits           = "0123456789abcdefghijklmnopqrstuvwxyz"
	initialFloatSize = 24
)

var shifts = [len(digits) + 1]uint{
	1 << 1: 1,
	1 << 2: 2,
	1 << 3: 3,
	1 << 4: 4,
	1 << 5: 5,
}

func appendKeyValue(b *bytes.Buffer, k, v []byte) {
	b.WriteByte('"')
	b.Write(k)
	b.WriteByte('"')
	b.WriteByte(':')
	b.WriteByte('"')
	b.Write(v)
	b.WriteByte('"')
	b.WriteByte(',')
	b.WriteByte(' ')
}

func appendString(b *bytes.Buffer, key, val string) {
	b.WriteByte('"')
	safeAppendString(b, key)
	b.WriteByte('"')
	b.WriteByte(':')
	b.WriteByte('"')
	safeAppendString(b, val)
	b.WriteByte('"')
	b.WriteByte(',')
	b.WriteByte(' ')
}

func appendBool(b *bytes.Buffer, key string, val bool) {
	b.WriteByte('"')
	safeAppendString(b, key)
	b.WriteByte('"')
	b.WriteByte(':')
	b.WriteByte('"')
	if val {
		b.WriteString("true")
	} else {
		b.WriteString("false")
	}
	b.WriteByte('"')
	b.WriteByte(',')
	b.WriteByte(' ')
}

func appendInt(b *bytes.Buffer, key string, val int) {
	appendInt64(b, key, int64(val))
}

func appendInt64(b *bytes.Buffer, key string, val int64) {
	b.WriteByte('"')
	safeAppendString(b, key)
	b.WriteByte('"')
	b.WriteByte(':')
	formatBits(b, uint64(val), 10, val < 0)
	b.WriteByte(',')
	b.WriteByte(' ')
}

func appendUint(b *bytes.Buffer, key string, val uint) {
	appendUint64(b, key, uint64(val))
}

func appendUint64(b *bytes.Buffer, key string, val uint64) {
	b.WriteByte('"')
	safeAppendString(b, key)
	b.WriteByte('"')
	b.WriteByte(':')
	formatBits(b, uint64(val), 10, val < 0)
	b.WriteByte(',')
	b.WriteByte(' ')
}

func appendUintptr(b *bytes.Buffer, key string, val uintptr) {
	appendUint64(b, key, uint64(val))
}

func appendFloat64(b *bytes.Buffer, key string, val float64) {
	b.WriteByte('"')
	safeAppendString(b, key)
	b.WriteByte('"')
	b.WriteByte(':')

	switch {
	case math.IsNaN(val):
		b.WriteString("NaN")
	case math.IsInf(val, 1):
		b.WriteString("+Inf")
	case math.IsInf(val, -1):
		b.WriteString("-Inf")
	default:
		b.Write(strconv.AppendFloat(make([]byte, 0, initialFloatSize), val, 'f', -1, 64))
	}
	b.WriteByte(',')
	b.WriteByte(' ')
}

// From `uber-go/zap`.
func safeAppendString(buf *bytes.Buffer, s string) {
	for i := 0; i < len(s); {
		if b := s[i]; b < utf8.RuneSelf {
			i++
			if 0x20 <= b && b != '\\' && b != '"' {
				buf.WriteByte(b)
				continue
			}
			switch b {
			case '\\', '"':
				buf.WriteByte('\\')
				buf.WriteByte('b')
			case '\n':
				buf.WriteByte('\\')
				buf.WriteByte('n')
			case '\r':
				buf.WriteByte('\\')
				buf.WriteByte('r')
			case '\t':
				buf.WriteByte('\\')
				buf.WriteByte('t')
			default:
				// Encode bytes < 0x20, except for the escape sequences above.
				buf.WriteString(`\u00`)
				buf.WriteByte(_hex[b>>4])
				buf.WriteByte(_hex[b&0xF])
			}
			continue
		}
		c, size := utf8.DecodeRuneInString(s[i:])
		if c == utf8.RuneError && size == 1 {
			buf.WriteString(`\ufffd`)
			i++
			continue
		}
		buf.WriteString(s[i : i+size])
		i += size
	}
}

// From go source code (https://golang.org/src/strconv/itoa.go?s=2995:3022#L60).
func formatBits(buf *bytes.Buffer, u uint64, base int, neg bool) {
	if base < 2 || base > len(digits) {
		panic("strconv: illegal AppendInt/FormatInt base")
	}
	// 2 <= base && base <= len(digits)

	var a [64 + 1]byte // +1 for sign of 64bit value in base 2
	i := len(a)

	if neg {
		u = -u
	}

	// convert bits
	if base == 10 {
		// common case: use constants for / because
		// the compiler can optimize it into a multiply+shift
		if ^uintptr(0)>>32 == 0 {
			for u > uint64(^uintptr(0)) {
				q := u / 1e9
				us := uintptr(u - q*1e9) // us % 1e9 fits into a uintptr
				for j := 9; j > 0; j-- {
					i--
					qs := us / 10
					a[i] = byte(us - qs*10 + '0')
					us = qs
				}
				u = q
			}
		}

		// u guaranteed to fit into a uintptr
		us := uintptr(u)
		for us >= 10 {
			i--
			q := us / 10
			a[i] = byte(us - q*10 + '0')
			us = q
		}
		// u < 10
		i--
		a[i] = byte(us + '0')

	} else if s := shifts[base]; s > 0 {
		// base is power of 2: use shifts and masks instead of / and %
		b := uint64(base)
		m := uintptr(b) - 1 // == 1<<s - 1
		for u >= b {
			i--
			a[i] = digits[uintptr(u)&m]
			u >>= s
		}
		// u < base
		i--
		a[i] = digits[uintptr(u)]

	} else {
		// general case
		b := uint64(base)
		for u >= b {
			i--
			q := u / b
			a[i] = digits[uintptr(u-q*b)]
			u = q
		}
		// u < base
		i--
		a[i] = digits[uintptr(u)]
	}

	// add sign, if any
	if neg {
		i--
		a[i] = '-'
	}

	buf.Write(a[i:])
}
