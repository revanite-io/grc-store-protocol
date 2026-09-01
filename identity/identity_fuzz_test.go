// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"strings"
	"testing"
)

// FuzzParseKeyless: the hub pins what CanonicalKeylessIdentity produces and
// grcli reads it back with ParseKeyless, so the pair must round-trip for
// every issuer/SAN a certificate can carry. The documented exception is an
// issuer containing "#" (split happens at the first one); that case is
// skipped, never asserted. Arbitrary strings fed to ParseKeyless must not
// panic. Crashers under testdata/fuzz/FuzzParseKeyless/ are regression
// seeds — commit them.
func FuzzParseKeyless(f *testing.F) {
	f.Add("https://token.actions.githubusercontent.com", "https://github.com/o/r/.github/workflows/publish.yml@refs/heads/main")
	f.Add("https://token.actions.githubusercontent.com", "https://github.com/o/r/.github/workflows/publish.yml@refs/tags/v1@refs/heads/x")
	f.Add("", "")
	f.Add("iss#with-hash", "san")
	f.Add("iss", "#san")
	f.Add("keyless:", "keyless:")
	f.Fuzz(func(t *testing.T, issuer, san string) {
		canonical := CanonicalKeylessIdentity(issuer, san)
		gotIssuer, gotPath, err := ParseKeyless(canonical)
		if err != nil {
			t.Fatalf("ParseKeyless(CanonicalKeylessIdentity(%q, %q) = %q): %v", issuer, san, canonical, err)
		}
		if !strings.Contains(issuer, "#") {
			if gotIssuer != issuer || gotPath != StripWorkflowRef(san) {
				t.Fatalf("round trip lost data: (%q, %q) -> %q -> (%q, %q)", issuer, san, canonical, gotIssuer, gotPath)
			}
		}
		ParseKeyless(issuer) // must not panic on arbitrary input
		ParseKeyless(san)
	})
}
