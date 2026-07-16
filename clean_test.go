package log

import (
	"io"
	golog "log"
	"sync"
	"testing"
)

func TestMaskID_Defaults(t *testing.T) {
	// Guard the defaults other tests/examples rely on.
	defer resetAuditMask()
	resetAuditMask()

	cases := []struct {
		name string
		kind IDKind
		in   string
		want string
	}{
		{"plain is untouched", Plain, "ten_xxx", "ten_xxx"},
		{"email keeps domain", Email, "david@acme.com", "d***d@acme.com"},
		{"short email local masked whole", Email, "al@acme.com", "***@acme.com"},
		{"non-address masked opaquely", Email, "notanemail", "n***l"},
		{"phone reveals last 4", Phone, "+2348012345678", "***5678"},
		{"govid reveals last 4", GovID, "A01234567", "***4567"},
		{"too short collapses", GovID, "12", "***"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := maskID(c.kind, c.in); got != c.want {
				t.Errorf("maskID(%v, %q) = %q, want %q", c.kind, c.in, got, c.want)
			}
		})
	}
}

func TestMaskID_Options(t *testing.T) {
	defer resetAuditMask()

	SetAuditMaskOptions(MaskChar('•'), RevealLast(2), KeepDomain(false))

	if got, want := maskID(Email, "david@acme.com"), "d•••d@a•••m"; got != want {
		t.Errorf("email with masked domain = %q, want %q", got, want)
	}
	if got, want := maskID(Phone, "+2348012345678"), "•••78"; got != want {
		t.Errorf("phone with RevealLast(2) = %q, want %q", got, want)
	}
}

// resetAuditMask restores the package defaults so global policy set by one test
// does not leak into another.
func resetAuditMask() {
	maskMu.Lock()
	defer maskMu.Unlock()
	auditMask = maskOptions{char: '*', revealLast: 4, keepDomain: true}
}

// TestConcurrentConfig exercises the logger, mask, and format configs under
// concurrent readers and writers; run with -race to catch unguarded access.
// Output is routed to a discard logger to keep the test quiet.
func TestConcurrentConfig(t *testing.T) {
	defer resetAuditMask()
	SetLogger(golog.New(io.Discard, "", 0))
	defer SetLogger(nil)
	defer SetFormat(Logfmt)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				SetLogLevel(LevelInfo)
				SetAuditMaskOptions(MaskChar('#'), RevealLast(2))
				SetFormat(JSON)
			} else {
				_ = getVerbosity()
				_ = maskID(Email, "david@acme.com")
				emit([]Field{{"k", "v"}})
			}
		}(i)
	}
	wg.Wait()
}
