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
		{"acme/baseline", "acme/staging/baseline", true},
		{"acme/team/baseline", "acme/staging/team/baseline", true},
		{"acme/baseline/staging", "acme/staging/baseline/staging", true},
		{"acme/staging/baseline", "", false},
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
		}
	}
}

func TestIsStaging(t *testing.T) {
	cases := map[string]bool{
		"acme/staging/baseline": true,
		"acme/staging/a/b":      true,
		"acme/staging":          false,
		"acme/staging/":         false,
		"staging/acme/baseline": false,
		"acme/baseline/staging": false,
		"acme/plugins/staging":  false,
		"acme/Staging/baseline": false,
		"/staging/baseline":     false,
		"acme/baseline":         false,
		"":                      false,
	}
	for repo, want := range cases {
		if got := IsStaging(repo); got != want {
			t.Errorf("IsStaging(%q) = %v, want %v", repo, got, want)
		}
	}
}
