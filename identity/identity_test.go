// SPDX-License-Identifier: Apache-2.0

package identity

import "testing"

// TestCanonicalKeylessIdentity_FieldValidated pins the exact string produced on
// the real preview publish that the hub and pvtr independently agreed on
// (sandbox/sandbox-plugin@0.26.1-rc, 2026-06). If this golden value ever
// changes, it is a BREAKING change to the contract — not a test to "update."
func TestCanonicalKeylessIdentity_FieldValidated(t *testing.T) {
	const issuer = "https://token.actions.githubusercontent.com"
	// The Fulcio SAN carries the per-run ref; canonicalization strips it.
	san := "https://github.com/eddie-knight/sandbox-plugin/.github/workflows/grcstore-publish.yml" +
		"@refs/tags/v0.26.1-rc"
	const want = "keyless:https://token.actions.githubusercontent.com" +
		"#https://github.com/eddie-knight/sandbox-plugin/.github/workflows/grcstore-publish.yml"

	if got := CanonicalKeylessIdentity(issuer, san); got != want {
		t.Fatalf("canonical identity drift:\n got=%q\nwant=%q", got, want)
	}
}

func TestCanonicalKeylessIdentity_RefIsStrippedSoReleaseN1Matches(t *testing.T) {
	const issuer = "https://token.actions.githubusercontent.com"
	const workflow = "https://github.com/acme/p/.github/workflows/release.yml"

	// Two different releases of the same workflow → identical canonical identity.
	v1 := CanonicalKeylessIdentity(issuer, workflow+"@refs/tags/v1.0.0")
	v2 := CanonicalKeylessIdentity(issuer, workflow+"@refs/tags/v2.0.0")
	if v1 != v2 {
		t.Fatalf("two releases of one workflow must canonicalize equally; v1=%q v2=%q", v1, v2)
	}
	if want := "keyless:" + issuer + "#" + workflow; v1 != want {
		t.Fatalf("got %q, want %q", v1, want)
	}
}

func TestStripWorkflowRef(t *testing.T) {
	cases := []struct{ name, in, want string }{
		{"tag ref", "https://github.com/o/r/.github/workflows/w.yml@refs/tags/v1", "https://github.com/o/r/.github/workflows/w.yml"},
		{"branch ref", "https://github.com/o/r/.github/workflows/w.yml@refs/heads/main", "https://github.com/o/r/.github/workflows/w.yml"},
		{"no ref passes through", "no-ref-san-passes-through", "no-ref-san-passes-through"},
		{"empty", "", ""},
		// Degenerate / malformed inputs — pinned so a future "more correct" change
		// is caught as the breaking change it would be:
		{"first @refs/ wins (only one strip)", "w.yml@refs/tags/v1@refs/tags/v2", "w.yml"},
		{"ref with no workflow prefix -> empty", "@refs/tags/v1", ""},
		{"bare @refs/ -> empty", "@refs/", ""},
	}
	for _, c := range cases {
		if got := StripWorkflowRef(c.in); got != c.want {
			t.Errorf("%s: StripWorkflowRef(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

// TestCanonicalKeylessIdentity_DegenerateInputs documents (does not endorse) the
// output for malformed inputs, so the behavior is a pinned contract, not an
// accident a maintainer might silently "fix".
func TestCanonicalKeylessIdentity_DegenerateInputs(t *testing.T) {
	cases := []struct{ name, issuer, san, want string }{
		{"empty issuer", "", "https://x/.github/workflows/w.yml@refs/heads/main", "keyless:#https://x/.github/workflows/w.yml"},
		{"empty san", "https://iss", "", "keyless:https://iss#"},
		{"both empty", "", "", "keyless:#"},
	}
	for _, c := range cases {
		if got := CanonicalKeylessIdentity(c.issuer, c.san); got != c.want {
			t.Errorf("%s: got %q, want %q", c.name, got, c.want)
		}
	}
}
