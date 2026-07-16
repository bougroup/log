package log

// Outcome describes the result of an audited action. Audit trails care most
// about the attempts that did not succeed — a denied or failed action is
// often the whole reason the trail exists.
type Outcome string

const (
	Success Outcome = "success"
	Denied  Outcome = "denied"
	Failed  Outcome = "failed"
)

// audit builds a single audit event: who did what, to which object, from
// where, and with what outcome. The event time is stamped at Emit. Start one
// with Audit and finish with Emit.
type audit struct {
	who     subject
	action  string
	what    object
	source  string
	outcome Outcome
	reason  string
}

// subject and object are the opaque values produced by the Who and What
// constructors. They are unexported so the mask Kind can only be set through a
// constructor — a caller cannot hand-build one and forget to classify a
// sensitive identifier as PII.
type subject struct {
	typ  string
	id   string
	kind IDKind
}

type object struct {
	typ  string
	id   string
	kind IDKind
}

// Who identifies the actor responsible for an action by an opaque, unmasked
// identifier (Plain). For a masked identifier use a WhoWith* helper, or This to
// pass a kind held in a variable.
func Who(typ, id string) subject {
	return subject{typ: typ, id: id, kind: Plain}
}

// This identifies the actor with an explicit mask kind — the subject-side twin
// of That, for when the kind is computed at runtime. Its fixed arity means
// passing more than one kind is a compile error.
func This(typ, id string, kind IDKind) subject {
	return subject{typ: typ, id: id, kind: kind}
}

// WhoWithEmail identifies an actor by an email address, masked as an email.
func WhoWithEmail(typ, id string) subject { return This(typ, id, Email) }

// WhoWithPhone identifies an actor by a phone number, revealing only its last digits.
func WhoWithPhone(typ, id string) subject { return This(typ, id, Phone) }

// WhoWithGovID identifies an actor by a government or national ID, revealing
// only its last characters.
func WhoWithGovID(typ, id string) subject { return This(typ, id, GovID) }

// What identifies the thing an action was performed on by an opaque, unmasked
// identifier (Plain). For a masked identifier use a WhatWith* helper, or That to
// pass a kind held in a variable.
func What(typ, id string) object {
	return object{typ: typ, id: id, kind: Plain}
}

// That identifies the thing acted on with an explicit mask kind — use it when
// the kind is computed at runtime. Its fixed arity means passing more than one
// kind is a compile error.
func That(typ, id string, kind IDKind) object {
	return object{typ: typ, id: id, kind: kind}
}

// WhatWithEmail identifies an object by an email address, masked as an email.
func WhatWithEmail(typ, id string) object { return That(typ, id, Email) }

// WhatWithPhone identifies an object by a phone number, revealing only its last digits.
func WhatWithPhone(typ, id string) object { return That(typ, id, Phone) }

// WhatWithGovID identifies an object by a government or national ID, revealing
// only its last characters.
func WhatWithGovID(typ, id string) object { return That(typ, id, GovID) }

// Audit starts an audit event. who is the actor responsible, did is the verb
// performed, and what is the thing acted on — together they are the irreducible
// core of an audit record, so they are required up front. Build who and what
// with the Who and What constructors (or their With* variants). Optional context
// is added through the chained methods before calling Emit.
func Audit(who subject, did string, what object) *audit {
	return &audit{who: who, action: did, what: what, outcome: Success}
}

// From records where the action originated — an IP, host, or request source.
func (a *audit) From(source string) *audit {
	a.source = source
	return a
}

// Outcome records the result of the action. Defaults to Success.
func (a *audit) Outcome(o Outcome) *audit {
	a.outcome = o
	return a
}

// Because records the reason for the action, most usefully the reason for a
// denial or failure.
func (a *audit) Because(reason string) *audit {
	a.reason = reason
	return a
}

// Emit writes the audit event. Audit events are always written, regardless of
// the configured log level, and are rendered with the active output Format
// (Logfmt by default, JSON via SetFormat).
func (a *audit) Emit() {
	fields := []Field{
		{"time", now()},
		{"kind", "audit"},
		{"subject", a.who.typ + "/" + maskID(a.who.kind, a.who.id)},
		{"action", a.action},
		{"object", a.what.typ + "/" + maskID(a.what.kind, a.what.id)},
	}
	if a.source != "" {
		fields = append(fields, Field{"source", a.source})
	}
	fields = append(fields, Field{"outcome", string(a.outcome)})
	if a.reason != "" {
		fields = append(fields, Field{"reason", a.reason})
	}
	emit(fields)
}
