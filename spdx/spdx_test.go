// SPDX-License-Identifier: Apache-2.0

package spdx

import (
	"errors"
	"testing"
)

// TestCanonicalize_Valid pins the canonical output for well-formed, fully-known
// expressions. These golden values are the contract grcli and the hub both
// produce; a change here is a change to what the index stores.
func TestCanonicalize_Valid(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"already canonical", "Apache-2.0", "Apache-2.0"},
		{"lowercase recased", "apache-2.0", "Apache-2.0"},
		{"uppercase recased", "MIT", "MIT"},
		{"mixed case recased", "mIt", "MIT"},
		{"surrounding whitespace", "  MIT  ", "MIT"},
		{"or-later plus preserved", "Apache-2.0+", "Apache-2.0+"},
		{"or-later recased", "gpl-3.0-or-later", "GPL-3.0-or-later"},
		{"dual OR", "MIT OR Apache-2.0", "MIT OR Apache-2.0"},
		{"dual OR ids recased (operators stay uppercase)", "mit OR apache-2.0", "MIT OR Apache-2.0"},
		{"AND", "MIT AND BSD-3-Clause", "MIT AND BSD-3-Clause"},
		{"WITH exception", "GPL-2.0-or-later WITH Classpath-exception-2.0", "GPL-2.0-or-later WITH Classpath-exception-2.0"},
		{"WITH exception ids recased", "gpl-2.0-or-later WITH classpath-exception-2.0", "GPL-2.0-or-later WITH Classpath-exception-2.0"},
		{"redundant parens around WITH-left dropped", "(MIT) WITH Classpath-exception-2.0", "MIT WITH Classpath-exception-2.0"},
		{"AND binds tighter than OR (no parens needed)", "MIT OR Apache-2.0 AND BSD-3-Clause", "MIT OR Apache-2.0 AND BSD-3-Clause"},
		{"explicit parens preserved where load-bearing", "(MIT OR Apache-2.0) AND BSD-3-Clause", "(MIT OR Apache-2.0) AND BSD-3-Clause"},
		{"redundant parens dropped", "(MIT)", "MIT"},
		{"redundant parens around AND-in-OR dropped", "MIT OR (Apache-2.0 AND BSD-3-Clause)", "MIT OR Apache-2.0 AND BSD-3-Clause"},
		{"nested", "(MIT OR Apache-2.0) AND (BSD-3-Clause OR ISC)", "(MIT OR Apache-2.0) AND (BSD-3-Clause OR ISC)"},
		{"deprecated id still valid", "GPL-2.0", "GPL-2.0"},
	}
	for _, c := range cases {
		got, err := Canonicalize(c.in)
		if err != nil {
			t.Errorf("%s: Canonicalize(%q) unexpected error: %v", c.name, c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: Canonicalize(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

// TestCanonicalize_LicenseRef covers the custom-identifier case the project
// itself relies on (LicenseRef-Revanite-Proprietary, ADR-0036): LicenseRef-/
// DocumentRef- tokens are valid grammar, exempt from the id-list check, and
// their suffix casing is PRESERVED (unlike standard ids).
func TestCanonicalize_LicenseRef(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"project's own proprietary ref", "LicenseRef-Revanite-Proprietary", "LicenseRef-Revanite-Proprietary"},
		{"ref casing preserved, not recased", "LicenseRef-My-Mixed-Case", "LicenseRef-My-Mixed-Case"},
		{"documentref", "DocumentRef-spdx-tool:LicenseRef-Foo", "DocumentRef-spdx-tool:LicenseRef-Foo"},
		{"ref OR known id", "LicenseRef-Proprietary OR MIT", "LicenseRef-Proprietary OR MIT"},
	}
	for _, c := range cases {
		got, err := Canonicalize(c.in)
		if err != nil {
			t.Errorf("%s: Canonicalize(%q) unexpected error: %v", c.name, c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: Canonicalize(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

// TestCanonicalize_Syntax pins malformed inputs to ErrSyntax. These are the
// inputs the hub maps to a NULL license (and logs); grcli rejects them too.
func TestCanonicalize_Syntax(t *testing.T) {
	bad := []struct{ name, in string }{
		{"empty", ""},
		{"whitespace only", "   "},
		{"trailing operator", "MIT OR"},
		{"leading operator", "OR MIT"},
		{"double operator", "MIT AND AND BSD-3-Clause"},
		{"adjacent ids no operator", "MIT Apache-2.0"},
		{"lowercase and is not an operator", "MIT and Apache-2.0"},
		{"unbalanced open paren", "(MIT"},
		{"unbalanced close paren", "MIT)"},
		{"empty parens", "()"},
		{"WITH on a compound group", "(MIT OR Apache-2.0) WITH Classpath-exception-2.0"},
		{"lowercase operator is not an operator", "MIT or Apache-2.0"},
		{"WITH without exception", "GPL-2.0-or-later WITH"},
		{"WITH followed by operator", "GPL-2.0-or-later WITH AND MIT"},
		{"plus on a ref", "LicenseRef-Foo+"},
		{"bare plus", "+"},
		{"illegal char", "MIT/Apache-2.0"},
		{"malformed ref", "LicenseRef-"},
		{"documentref missing colon", "DocumentRef-foo"},
	}
	for _, c := range bad {
		_, err := Canonicalize(c.in)
		if !errors.Is(err, ErrSyntax) {
			t.Errorf("%s: Canonicalize(%q) error = %v, want ErrSyntax", c.name, c.in, err)
		}
	}
}

// TestCanonicalize_UnknownID: well-formed but unrecognized ids are fatal for the
// strict path (grcli), distinguished from ErrSyntax so the hub can tell a typo
// from a not-yet-known id.
func TestCanonicalize_UnknownID(t *testing.T) {
	cases := []struct{ name, in string }{
		{"typo'd id", "Apache2"},
		{"made-up id", "Foo-1.0"},
		{"unknown exception", "MIT WITH No-Such-Exception-1.0"},
		{"one known one unknown", "MIT OR Foo-1.0"},
	}
	for _, c := range cases {
		_, err := Canonicalize(c.in)
		if !errors.Is(err, ErrUnknownID) {
			t.Errorf("%s: Canonicalize(%q) error = %v, want ErrUnknownID", c.name, c.in, err)
		}
		if errors.Is(err, ErrSyntax) {
			t.Errorf("%s: Canonicalize(%q) wrongly classified as ErrSyntax", c.name, c.in)
		}
	}
}

// TestHubLenientPath models the hub's behavior (ADR-0036 decision 4): Parse
// validates grammar only; String canonicalizes; UnknownIDs reports — but an
// unknown-but-well-formed id is NEVER a parse error (no orphaned bytes). This is
// QA review test T12.
func TestHubLenientPath(t *testing.T) {
	// Well-formed, unknown id: hub stores the structurally-canonical form as-is.
	const newerThanHub = "future-license-9.0 OR mit"
	e, err := Parse(newerThanHub)
	if err != nil {
		t.Fatalf("Parse(%q) should not error on an unknown-but-well-formed id: %v", newerThanHub, err)
	}
	if got, want := e.String(), "future-license-9.0 OR MIT"; got != want {
		// Known id (mit) is recased; unknown id is passed through verbatim.
		t.Errorf("lenient canonical = %q, want %q", got, want)
	}
	if unknown := e.UnknownIDs(); len(unknown) != 1 || unknown[0] != "future-license-9.0" {
		t.Errorf("UnknownIDs = %v, want [future-license-9.0]", unknown)
	}

	// Malformed: hub gets an error and stores NULL. This is QA review test T11.
	if _, err := Parse("MIT AND AND"); !errors.Is(err, ErrSyntax) {
		t.Errorf("malformed expression should be ErrSyntax for the hub's NULL path, got %v", err)
	}
}

// TestParse_RoundTripStable: canonical output is idempotent — re-canonicalizing
// a canonical string yields itself. Protects the "both ends agree byte-for-byte"
// contract.
func TestParse_RoundTripStable(t *testing.T) {
	inputs := []string{
		"mit",
		"(mit OR apache-2.0) AND bsd-3-clause",
		"gpl-2.0-or-later WITH classpath-exception-2.0",
		"LicenseRef-Revanite-Proprietary",
	}
	for _, in := range inputs {
		once, err := Canonicalize(in)
		if err != nil {
			t.Fatalf("Canonicalize(%q): %v", in, err)
		}
		twice, err := Canonicalize(once)
		if err != nil {
			t.Fatalf("Canonicalize(%q) [2nd]: %v", once, err)
		}
		if once != twice {
			t.Errorf("not idempotent: %q -> %q -> %q", in, once, twice)
		}
	}
}

// TestSPDXListVersionPresent guards against a missing/empty generated data file.
func TestSPDXListVersionPresent(t *testing.T) {
	if SPDXListVersion == "" {
		t.Fatal("SPDXListVersion empty — run `go generate ./spdx`")
	}
	if _, ok := spdxLicenseIDs["apache-2.0"]; !ok {
		t.Error("license table missing apache-2.0 — generated data looks wrong")
	}
	if _, ok := spdxExceptionIDs["classpath-exception-2.0"]; !ok {
		t.Error("exception table missing classpath-exception-2.0 — generated data looks wrong")
	}
}
