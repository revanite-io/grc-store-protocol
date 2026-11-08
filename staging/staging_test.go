// SPDX-License-Identifier: Apache-2.0

package staging

import "testing"

// TestRepository is the contract table: the hub, grcli and pvtr are each
// expected to agree with every row.
func TestRepository(t *testing.T) {
	cases := []struct {
		served, want string
		ok           bool
	}{
		{"acme/baseline", "staging/acme/baseline", true},
		{"acme/team/baseline", "staging/acme/team/baseline", true},
		{"acme/plugins/scanner", "staging/acme/plugins/scanner", true},
		{"acme/baseline/staging", "staging/acme/baseline/staging", true},
		{"acme/staging/baseline", "staging/acme/staging/baseline", true},
		{"acme/staging", "staging/acme/staging", true},
		{"staging/acme/baseline", "", false},
		{"staging/acme", "", false},
		{"acme", "", false},
		{"acme/", "", false},
		{"/baseline", "", false},
		{"acme//baseline", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, ok := Repository(c.served)
		if got != c.want || ok != c.ok {
			t.Errorf("Repository(%q) = %q, %v; want %q, %v", c.served, got, ok, c.want, c.ok)
		}
		if ok {
			if !IsStaging(got) {
				t.Errorf("IsStaging(%q) = false for Repository's own output", got)
			}
			if back, ok := Served(got); !ok || back != c.served {
				t.Errorf("Served(%q) = %q, %v; want %q, true", got, back, ok, c.served)
			}
		}
	}
}

func TestIsStaging(t *testing.T) {
	cases := map[string]bool{
		"staging/acme/baseline":  true,
		"staging/acme/a/b":       true,
		"staging/acme/staging":   true,
		"staging":                false,
		"staging/":               false,
		"staging/acme":           false,
		"staging/acme/":          false,
		"staging//baseline":      false,
		"staging/acme//b":        false,
		"staging/staging/acme/b": false,
		"Staging/acme/baseline":  false,
		"acme/staging/baseline":  false,
		"acme/baseline/staging":  false,
		"acme/baseline":          false,
		"":                       false,
	}
	for repo, want := range cases {
		if got := IsStaging(repo); got != want {
			t.Errorf("IsStaging(%q) = %v, want %v", repo, got, want)
		}
		if _, ok := Served(repo); ok != want {
			t.Errorf("Served(%q) ok = %v, want %v", repo, ok, want)
		}
	}
}
