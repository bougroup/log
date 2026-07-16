package log

import "testing"

var sampleFields = []Field{
	{"kind", "audit"},
	{"action", "delete"},
	{"reason", "insufficient role"},
}

func TestLogfmt(t *testing.T) {
	got := Logfmt(sampleFields)
	want := `kind=audit action=delete reason="insufficient role"`
	if got != want {
		t.Errorf("Logfmt = %q, want %q", got, want)
	}
}

func TestJSON(t *testing.T) {
	got := JSON(sampleFields)
	want := `{"kind":"audit","action":"delete","reason":"insufficient role"}`
	if got != want {
		t.Errorf("JSON = %q, want %q", got, want)
	}
}

func TestJSONEscapes(t *testing.T) {
	got := JSON([]Field{{"msg", `he said "hi"` + "\t"}})
	want := `{"msg":"he said \"hi\"\t"}`
	if got != want {
		t.Errorf("JSON = %q, want %q", got, want)
	}
}
