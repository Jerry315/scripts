// adapted from https://github.com/uber-go/zap/blob/master/zapcore/json_encoder.go
// and https://github.com/uber-go/zap/blob/master/zapcore/console_encoder.go

package stdlog

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"math"
	"os"
	"sync"
	"time"
	"unicode/utf8"

	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const _hex = "0123456789abcdef"

var bufferPool = buffer.NewPool()

var _consolePool = sync.Pool{New: func() interface{} {
	return &consoleEncoder{}
}}

func getConsoleEncoder() *consoleEncoder {
	return _consolePool.Get().(*consoleEncoder)
}

func putConsoleEncoder(enc *consoleEncoder) {
	enc.EncoderConfig = nil
	enc.buf = nil
	_consolePool.Put(enc)
}

type consoleEncoder struct {
	*zapcore.EncoderConfig
	buf      *buffer.Buffer
	hostname string
	detail   bool
}

// NewConsoleEncoder creates a key=value encoder
func NewConsoleEncoder(cfg zapcore.EncoderConfig, detail bool) zapcore.Encoder {
	hostname, err := os.Hostname()
	if err != nil {
	}
	return &consoleEncoder{
		EncoderConfig: &cfg,
		buf:           bufferPool.Get(),
		hostname:      hostname,
		detail:        detail,
	}
}

func (enc *consoleEncoder) AddArray(key string, arr zapcore.ArrayMarshaler) error {
	enc.addKey(key)
	return enc.AppendArray(arr)
}

func (enc *consoleEncoder) AddObject(key string, obj zapcore.ObjectMarshaler) error {
	enc.addKey(key)
	return enc.AppendObject(obj)
}

func (enc *consoleEncoder) AddBinary(key string, val []byte) {
	enc.AddString(key, base64.StdEncoding.EncodeToString(val))
}

func (enc *consoleEncoder) AddByteString(key string, val []byte) {
	enc.addKey(key)
	enc.AppendByteString(val)
}

func (enc *consoleEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.AppendBool(val)
}

func (enc *consoleEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.AppendComplex128(val)
}

func (enc *consoleEncoder) AddDuration(key string, val time.Duration) {
	enc.addKey(key)
	enc.AppendDuration(val)
}

func (enc *consoleEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.AppendFloat64(val)
}

func (enc *consoleEncoder) AddInt64(key string, val int64) {
	enc.addKey(key)
	enc.AppendInt64(val)
}

func (enc *consoleEncoder) AddReflected(key string, obj interface{}) error {
	marshaled, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	enc.addKey(key)
	_, err = enc.buf.Write(marshaled)
	return err
}

func (enc *consoleEncoder) OpenNamespace(key string) {
}

func (enc *consoleEncoder) AddString(key, val string) {
	enc.addKey(key)
	enc.AppendString(val)
}

func (enc *consoleEncoder) AddTime(key string, val time.Time) {
	enc.addKey(key)
	enc.AppendTime(val)
}

func (enc *consoleEncoder) AddUint64(key string, val uint64) {
	enc.addKey(key)
	enc.AppendUint64(val)
}

func (enc *consoleEncoder) AppendArray(arr zapcore.ArrayMarshaler) error {
	return arr.MarshalLogArray(enc)
}

func (enc *consoleEncoder) AppendObject(obj zapcore.ObjectMarshaler) error {
	return obj.MarshalLogObject(enc)
}

func (enc *consoleEncoder) AppendBool(val bool) {
	enc.buf.AppendBool(val)
}

func (enc *consoleEncoder) AppendByteString(val []byte) {
	enc.safeAddByteString(val)
}

func (enc *consoleEncoder) AppendComplex128(val complex128) {
	// Cast to a platform-independent, fixed-size type.
	r, i := float64(real(val)), float64(imag(val))
	enc.buf.AppendByte('"')
	// Because we're always in a quoted string, we can use strconv without
	// special-casing NaN and +/-Inf.
	enc.buf.AppendFloat(r, 64)
	enc.buf.AppendByte('+')
	enc.buf.AppendFloat(i, 64)
	enc.buf.AppendByte('i')
	enc.buf.AppendByte('"')
}

func (enc *consoleEncoder) AppendDuration(val time.Duration) {
	cur := enc.buf.Len()
	enc.EncodeDuration(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeDuration is a no-op. Fall back to nanoseconds to keep
		// JSON valid.
		enc.AppendInt64(int64(val))
	}
}

func (enc *consoleEncoder) AppendInt64(val int64) {
	enc.buf.AppendInt(val)
}

func (enc *consoleEncoder) AppendReflected(val interface{}) error {
	marshaled, err := json.Marshal(val)
	if err != nil {
		return err
	}
	_, err = enc.buf.Write(marshaled)
	return err
}

func (enc *consoleEncoder) AppendString(val string) {
	enc.safeAddString(val)
}

func (enc *consoleEncoder) AppendTime(val time.Time) {
	cur := enc.buf.Len()
	enc.EncodeTime(val, enc)
	if cur == enc.buf.Len() {
		// User-supplied EncodeTime is a no-op. Fall back to nanos since epoch to keep
		// output JSON valid.
		enc.AppendInt64(val.UnixNano())
	}
}

func (enc *consoleEncoder) AppendUint64(val uint64) {
	enc.buf.AppendUint(val)
}

func (enc *consoleEncoder) AddComplex64(k string, v complex64) { enc.AddComplex128(k, complex128(v)) }
func (enc *consoleEncoder) AddFloat32(k string, v float32)     { enc.AddFloat64(k, float64(v)) }
func (enc *consoleEncoder) AddInt(k string, v int)             { enc.AddInt64(k, int64(v)) }
func (enc *consoleEncoder) AddInt32(k string, v int32)         { enc.AddInt64(k, int64(v)) }
func (enc *consoleEncoder) AddInt16(k string, v int16)         { enc.AddInt64(k, int64(v)) }
func (enc *consoleEncoder) AddInt8(k string, v int8)           { enc.AddInt64(k, int64(v)) }
func (enc *consoleEncoder) AddUint(k string, v uint)           { enc.AddUint64(k, uint64(v)) }
func (enc *consoleEncoder) AddUint32(k string, v uint32)       { enc.AddUint64(k, uint64(v)) }
func (enc *consoleEncoder) AddUint16(k string, v uint16)       { enc.AddUint64(k, uint64(v)) }
func (enc *consoleEncoder) AddUint8(k string, v uint8)         { enc.AddUint64(k, uint64(v)) }
func (enc *consoleEncoder) AddUintptr(k string, v uintptr)     { enc.AddUint64(k, uint64(v)) }
func (enc *consoleEncoder) AppendComplex64(v complex64)        { enc.AppendComplex128(complex128(v)) }
func (enc *consoleEncoder) AppendFloat64(v float64)            { enc.appendFloat(v, 64) }
func (enc *consoleEncoder) AppendFloat32(v float32)            { enc.appendFloat(float64(v), 32) }
func (enc *consoleEncoder) AppendInt(v int)                    { enc.AppendInt64(int64(v)) }
func (enc *consoleEncoder) AppendInt32(v int32)                { enc.AppendInt64(int64(v)) }
func (enc *consoleEncoder) AppendInt16(v int16)                { enc.AppendInt64(int64(v)) }
func (enc *consoleEncoder) AppendInt8(v int8)                  { enc.AppendInt64(int64(v)) }
func (enc *consoleEncoder) AppendUint(v uint)                  { enc.AppendUint64(uint64(v)) }
func (enc *consoleEncoder) AppendUint32(v uint32)              { enc.AppendUint64(uint64(v)) }
func (enc *consoleEncoder) AppendUint16(v uint16)              { enc.AppendUint64(uint64(v)) }
func (enc *consoleEncoder) AppendUint8(v uint8)                { enc.AppendUint64(uint64(v)) }
func (enc *consoleEncoder) AppendUintptr(v uintptr)            { enc.AppendUint64(uint64(v)) }

func (enc *consoleEncoder) Clone() zapcore.Encoder {
	clone := enc.clone()
	_, err := clone.buf.Write(enc.buf.Bytes())
	if err != nil {
		log.Printf("consoleEncoder: Clone failed %v\n", err)
	}
	return clone
}

func (enc *consoleEncoder) clone() *consoleEncoder {
	clone := getConsoleEncoder()
	clone.EncoderConfig = enc.EncoderConfig
	clone.buf = bufferPool.Get()
	clone.hostname = enc.hostname
	clone.detail = enc.detail
	return clone
}

func (enc *consoleEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	final := enc.clone()

	if final.TimeKey != "" {
		final.AppendTime(ent.Time)
		final.addElementSeparator()
	}
	if final.LevelKey != "" {
		cur := final.buf.Len()
		final.EncodeLevel(ent.Level, final)
		if cur == final.buf.Len() {
			final.AppendString(ent.Level.String())
		}
		final.addElementSeparator()
	}
	if enc.hostname != "" {
		final.buf.AppendByte('[')
		final.buf.AppendString(enc.hostname)
		final.buf.AppendByte(']')
		final.addElementSeparator()
	}
	if ent.LoggerName != "" && final.NameKey != "" {
		cur := final.buf.Len()
		nameEncoder := final.EncodeName

		// if no name encoder provided, fall back to FullNameEncoder for backwards
		// compatibility
		if nameEncoder == nil {
			nameEncoder = zapcore.FullNameEncoder
		}

		nameEncoder(ent.LoggerName, final)
		if cur == final.buf.Len() {
			// User-supplied EncodeName was a no-op. Fall back to strings to
			// keep output valid.
			final.AppendString(ent.LoggerName)
		}
		final.addElementSeparator()
	}
	if ent.Caller.Defined && final.CallerKey != "" {
		cur := final.buf.Len()
		final.EncodeCaller(ent.Caller, final)
		if cur == final.buf.Len() {
			// User-supplied EncodeCaller was a no-op. Fall back to strings to
			// keep JSON valid.
			final.AppendString(ent.Caller.String())
		}
		final.addElementSeparator()
	}
	if final.MessageKey != "" {
		final.buf.AppendByte('"')
		final.AppendString(ent.Message)
		final.buf.AppendByte('"')
		final.addElementSeparator()
	}
	if enc.buf.Len() > 0 {
		_, err := final.buf.Write(enc.buf.Bytes())
		if err != nil {
			log.Printf("consoleEncoder: EncodeEntry write enc buf failed %v\n", err)
		}
	}

	final.buf.AppendByte('-')
	final.addElementSeparator()
	addFields(final, final, fields)
	final.addElementSeparator()
	if ent.Stack != "" && final.StacktraceKey != "" {
		final.AddString(final.StacktraceKey, ent.Stack)
		final.addElementSeparator()
	}
	if final.LineEnding != "" {
		final.buf.AppendString(final.LineEnding)
	} else {
		final.buf.AppendString(zapcore.DefaultLineEnding)
	}

	ret := final.buf
	putConsoleEncoder(final)
	return ret, nil
}

func (enc *consoleEncoder) addKey(key string) {
	enc.buf.AppendString(key)
	enc.buf.AppendByte(':')
}

func (enc *consoleEncoder) addElementSeparator() {
	enc.buf.AppendByte(' ')
}

func (enc *consoleEncoder) appendFloat(val float64, bitSize int) {
	switch {
	case math.IsNaN(val):
		enc.buf.AppendString(`"NaN"`)
	case math.IsInf(val, 1):
		enc.buf.AppendString(`"+Inf"`)
	case math.IsInf(val, -1):
		enc.buf.AppendString(`"-Inf"`)
	default:
		enc.buf.AppendFloat(val, bitSize)
	}
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (enc *consoleEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.AppendString(s[i : i+size])
		i += size
	}
}

// safeAddByteString is no-alloc equivalent of safeAddString(string(s)) for s []byte.
func (enc *consoleEncoder) safeAddByteString(s []byte) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRune(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		_, err := enc.buf.Write(s[i : i+size])
		if err != nil {
			log.Printf("consoleEncoder: safeAddByteString write buf failed %v\n", err)
		}
		i += size
	}
}

// tryAddRuneSelf appends b if it is valid UTF-8 character represented in a single byte.
func (enc *consoleEncoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		enc.buf.AppendByte(b)
		return true
	}
	switch b {
	case '\\', '"':
		if !enc.detail {
			enc.buf.AppendByte('\\')
			enc.buf.AppendByte(b)
		} else {
			enc.buf.AppendByte(b)
		}
	case '\n':
		if !enc.detail {
			enc.buf.AppendByte('\\')
			enc.buf.AppendByte('n')
		} else {
			enc.buf.AppendByte('\n')
		}
	case '\r':
		if !enc.detail {
			enc.buf.AppendByte('\\')
			enc.buf.AppendByte('r')
		} else {
			enc.buf.AppendByte('\r')
		}
	case '\t':
		if !enc.detail {
			enc.buf.AppendByte('\\')
			enc.buf.AppendByte('t')
		} else {
			enc.buf.AppendByte('\t')
		}
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		enc.buf.AppendString(`\u00`)
		enc.buf.AppendByte(_hex[b>>4])
		enc.buf.AppendByte(_hex[b&0xF])
	}
	return true
}

func (enc *consoleEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}

func addFields(consoleEnc *consoleEncoder, enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	if !consoleEnc.detail {
		consoleEnc.buf.AppendByte('[')
	}
	lastIdx := len(fields) - 1
	for i := range fields {
		fields[i].AddTo(enc)
		if i != lastIdx {
			consoleEnc.buf.AppendByte(',')
			consoleEnc.buf.AppendByte(' ')
		}
	}
	if !consoleEnc.detail {
		consoleEnc.buf.AppendByte(']')
	}
}
