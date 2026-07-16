package log_test

import "github.com/bougroup/log"

// ExampleAudit shows a denied action. The subject is identified by an email,
// which Emit masks according to the global policy; the object is an opaque
// system ID (Kind defaults to Plain) and is stored as-is.
func ExampleAudit() {
	log.Audit(
		log.WhoWithEmail("admin", "david@acme.com"),
		"delete",
		log.That("user", "jane@acme.com", log.Email),
	).
		From("10.0.0.4").
		Outcome(log.Denied).
		Because("insufficient role").
		Emit()
}

// ExampleAudit_minimal shows the required core. Even a login has an object —
// the app the user signed in to.
func ExampleAudit_minimal() {
	log.Audit(
		log.WhoWithEmail("user", "oke@bougroup.com"),
		"login",
		log.What("app", "v2.3.1"),
	).Emit()
}

// ExampleSetFormat switches output to JSON for ingestion by JSON-aware log
// tooling such as Dozzle or a Loki pipeline; the default is Logfmt.
func ExampleSetFormat() {
	log.SetFormat(log.JSON)
}

// ExampleSetAuditMaskOptions configures the masking policy once at startup. It
// then applies to every audit record; if never called, sensible defaults apply.
func ExampleSetAuditMaskOptions() {
	log.SetAuditMaskOptions(
		log.MaskChar('•'),
		log.RevealLast(2),
		log.KeepDomain(false),
	)
}
