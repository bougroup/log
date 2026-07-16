package log

import (
	"strings"
	"sync"
)

// IDKind classifies the identifier carried in a Who or What so an audit
// record knows how to mask it. The zero value, Plain, is never masked — use it
// for opaque system identifiers that are safe to store as-is (e.g. "ten_xxx").
type IDKind int

const (
	Plain IDKind = iota // stored as-is, never masked
	Email               // local part masked; domain kept unless KeepDomain(false)
	Phone               // all but the last RevealLast digits masked
	GovID               // passport / national ID etc.; all but the last RevealLast masked
)

// maskOptions is the global masking policy applied to audit identifiers. The
// masking is deterministic — equal inputs map to equal outputs under a fixed
// policy — so masked values still group across audit records.
type maskOptions struct {
	char       rune
	revealLast int
	keepDomain bool
}

// maskMu guards auditMask so SetAuditMaskOptions can be called safely while
// other goroutines are writing audit records.
var maskMu sync.RWMutex

// auditMask holds the active policy. It carries sensible defaults so masking
// works without any configuration; SetAuditMaskOptions overrides it.
var auditMask = maskOptions{
	char:       '*',
	revealLast: 4,
	keepDomain: true,
}

// MaskOption configures the global audit masking policy.
type MaskOption func(*maskOptions)

// MaskChar sets the glyph used to hide masked characters. Default '*'.
func MaskChar(r rune) MaskOption {
	return func(o *maskOptions) { o.char = r }
}

// RevealLast sets how many trailing characters stay visible for Phone and GovID
// identifiers. Default 4.
func RevealLast(n int) MaskOption {
	return func(o *maskOptions) { o.revealLast = n }
}

// KeepDomain controls whether the domain of an Email identifier stays visible
// for triage. Default true.
func KeepDomain(keep bool) MaskOption {
	return func(o *maskOptions) { o.keepDomain = keep }
}

// SetAuditMaskOptions configures, globally, how audit identifiers are masked.
// Call it once at startup before any audit records are written; if it is never
// called, sensible defaults apply. It is not safe to call concurrently with
// Emit.
func SetAuditMaskOptions(opts ...MaskOption) {
	maskMu.Lock()
	defer maskMu.Unlock()
	for _, opt := range opts {
		opt(&auditMask)
	}
}

// maskID masks value according to its kind under the active global policy. It
// snapshots the policy under a read lock so the masking runs against a single
// consistent value even if SetAuditMaskOptions is called concurrently.
func maskID(kind IDKind, value string) string {
	maskMu.RLock()
	o := auditMask
	maskMu.RUnlock()
	switch kind {
	case Email:
		return o.email(value)
	case Phone, GovID:
		return o.trailing(value)
	default:
		return value
	}
}

// email masks the local part of an address, keeping the domain visible unless
// the policy says otherwise: "david@acme.com" -> "d***d@acme.com".
func (o maskOptions) email(email string) string {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return o.middle(email, 1, 1) // not an address; mask as an opaque value
	}
	domain := email[at+1:]
	if !o.keepDomain {
		domain = o.middle(domain, 1, 1)
	}
	return o.middle(email[:at], 1, 1) + "@" + domain
}

// trailing reveals only the last revealLast runes: "+2348012345678" -> "***5678".
func (o maskOptions) trailing(s string) string {
	return o.middle(s, 0, o.revealLast)
}

// middle reveals the first head and last tail runes of s and replaces the
// interior with a fixed three-glyph mask, so the output neither exposes the
// interior nor leaks the exact length. If s is too short to reveal head+tail
// runes without exposing the whole value, it is masked entirely.
func (o maskOptions) middle(s string, head, tail int) string {
	r := []rune(s)
	fill := strings.Repeat(string(o.char), 3)
	if len(r) <= head+tail {
		return fill
	}
	return string(r[:head]) + fill + string(r[len(r)-tail:])
}
