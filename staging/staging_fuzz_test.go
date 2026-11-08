// SPDX-License-Identifier: Apache-2.0

package staging

import "testing"

// FuzzStaging: the hub grants push from IsStaging on untrusted registry-token
// scopes, so the three functions must agree on every input: IsStaging is
// Served's ok, and Repository and Served are exact inverses wherever either
// succeeds. Crashers under testdata/fuzz/FuzzStaging/ are regression seeds.
func FuzzStaging(f *testing.F) {
	for _, s := range []string{
		"staging/acme/b", "staging/acme/plugins/x", "staging/staging/acme/b",
		"acme/staging/b", "staging/acme", "staging//b", "staging/acme/",
		"Staging/acme/b", "acme/b", "", "staging", "/",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		served, ok := Served(s)
		if IsStaging(s) != ok {
			t.Fatalf("IsStaging(%q) = %v, Served ok = %v", s, !ok, ok)
		}
		if ok {
			if back, rok := Repository(served); !rok || back != s {
				t.Fatalf("Repository(Served(%q) = %q) = %q, %v", s, served, back, rok)
			}
		}
		if repo, rok := Repository(s); rok {
			if back, sok := Served(repo); !sok || back != s {
				t.Fatalf("Served(Repository(%q) = %q) = %q, %v", s, repo, back, sok)
			}
		}
	})
}
