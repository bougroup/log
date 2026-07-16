package log

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Field is one key/value pair in a log record. Values are rendered as strings —
// levels, audit identifiers, and timestamps are all string-valued.
type Field struct {
	Key   string
	Value string
}

// Format renders an ordered set of fields into a single log line (no trailing
// newline). Logfmt and JSON are provided; supply your own to SetFormat for a
// custom rendering.
type Format func(fields []Field) string

// formatMu guards the active output format.
var formatMu sync.RWMutex
var format Format = Logfmt

// SetFormat selects the output format for all log output — audit records and
// level logs alike — so the stream stays uniform for parsers like Loki or
// Dozzle. Defaults to Logfmt. It is not safe to call concurrently with logging.
func SetFormat(f Format) {
	formatMu.Lock()
	defer formatMu.Unlock()
	format = f
}

// emit renders fields with the active format and writes the line to the sink.
func emit(fields []Field) {
	formatMu.RLock()
	f := format
	formatMu.RUnlock()
	printLine(f(fields))
}

// Logfmt renders fields as space-separated key=value pairs, quoting a value
// that contains a space or quote so the pair survives parsing. Loki parses this
// natively via the logfmt pipeline stage.
func Logfmt(fields []Field) string {
	parts := make([]string, len(fields))
	for i, f := range fields {
		if strings.ContainsAny(f.Value, ` "`) {
			parts[i] = fmt.Sprintf("%s=%q", f.Key, f.Value)
		} else {
			parts[i] = f.Key + "=" + f.Value
		}
	}
	return strings.Join(parts, " ")
}

// JSON renders fields as a JSON object, preserving field order. Keys and values
// are JSON-escaped, so the line is safe for JSON-aware tooling (Dozzle, ELK,
// Vector, Datadog) to ingest.
func JSON(fields []Field) string {
	var b strings.Builder
	b.WriteByte('{')
	for i, f := range fields {
		if i > 0 {
			b.WriteByte(',')
		}
		key, _ := json.Marshal(f.Key)
		val, _ := json.Marshal(f.Value)
		b.Write(key)
		b.WriteByte(':')
		b.Write(val)
	}
	b.WriteByte('}')
	return b.String()
}

// now returns the current time formatted for a log record.
func now() string {
	return time.Now().Format(time.RFC3339)
}
