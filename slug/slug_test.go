// SPDX-License-Identifier: Apache-2.0

package slug

import (
	"strings"
	"testing"
	"time"
)

// TestSlugify pins the hub's rule (internal/store.Slugify) verbatim. A verdict
// change here moves live coordinates — it is a breaking change, not a test to
// "update".
func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"finos", "finos"},
		{"FINOS CCC", "finos-ccc"},
		{"ccc.iam.cn", "ccc.iam.cn"},           // dots preserved
		{"my_plugin", "my-plugin"},             // '_' is NOT preserved
		{"web_api_ccc.iam", "web-api-ccc.iam"}, // results stream id
		{"--a--b--", "a-b"},                    // runs collapse, ends trimmed
		{"  spaced  out  ", "spaced-out"},
		{"https://github.com/acme/scanner", "https-github.com-acme-scanner"},
		{"Ünïcode", "n-code"}, // non-ASCII letters are not [a-z]
		{"...", "..."},        // dots alone survive
		{"-", ""},
		{"_", ""},
		{"a/b", "a-b"},
	}
	for _, c := range cases {
		if got := Slugify(c.in); got != c.want {
			t.Errorf("Slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestIsSlug(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"finos", true},
		{"ccc.iam.cn", true},
		{"web-api-ccc.iam", true},
		{"Finos", false},
		{"my_plugin", false},
		{"-a", false},
		{"a--b", false},
		{"a/b", false},
	}
	for _, c := range cases {
		if got := IsSlug(c.in); got != c.want {
			t.Errorf("IsSlug(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestEvaluationLogRepository(t *testing.T) {
	cases := []struct{ ns, target, catalog, want string }{
		{"acme", "web-api", "ccc.iam", "acme/web-api-ccc.iam"},
		{"Acme", "Web API", "CCC_IAM", "acme/web-api-ccc-iam"},
		{"acme", "a", "b", "acme/a-b"},
	}
	for _, c := range cases {
		if got := EvaluationLogRepository(c.ns, c.target, c.catalog); got != c.want {
			t.Errorf("EvaluationLogRepository(%q,%q,%q) = %q, want %q", c.ns, c.target, c.catalog, got, c.want)
		}
	}
}

func TestEvaluationLogVersion(t *testing.T) {
	est := time.FixedZone("EST", -5*3600)
	at := time.Date(2026, 9, 4, 5, 15, 0, 999, est) // 10:15:00Z
	if got, want := EvaluationLogVersion("1.4.2", at), "1.4.2-20260904T101500Z"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if got, want := EvaluationLogVersion("2.0.0-rc1", at), "2.0.0-rc1-20260904T101500Z"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// TestIsHubPluginCoordinate is the precise "looks like a hub coordinate" rule.
// Every row is a policy verdict.
func TestIsHubPluginCoordinate(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"privateer/github", true},
		{"local/my-plugin", true}, // IS a coordinate: fails closed when unpublished
		{"acme/ccc.scanner", true},
		{"acme-scanner", false}, // one segment
		{"", false},
		{"/", false},
		{"acme/", false},
		{"/scanner", false},
		{"a/b/c", false},                           // three segments
		{"Acme/scanner", false},                    // uppercase
		{"acme/my_plugin", false},                  // '_' is not slug
		{"https://github.com/acme/scanner", false}, // URL
		{"acme scanner/x", false},
	}
	for _, c := range cases {
		if got := IsHubPluginCoordinate(c.in); got != c.want {
			t.Errorf("IsHubPluginCoordinate(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// FuzzSlugify: Slugify must never panic, and its output must satisfy the
// invariants every consumer relies on — idempotent (so re-slugifying a
// coordinate is a no-op, which ADR-0050's normalization depends on), drawn
// from [a-z0-9.-], no leading/trailing '-', no "--", and IsSlug of a non-empty
// result. Crashers under testdata/fuzz/FuzzSlugify/ are regression seeds —
// commit them.
func FuzzSlugify(f *testing.F) {
	for _, s := range []string{"", "FINOS CCC", "ccc.iam.cn", "my_plugin", "--a--", "Ünïcode", "\xff\xfe", "a/b", "."} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		out := Slugify(s)
		if again := Slugify(out); again != out {
			t.Fatalf("not idempotent: Slugify(%q)=%q, Slugify(that)=%q", s, out, again)
		}
		if strings.Trim(out, "abcdefghijklmnopqrstuvwxyz0123456789.-") != "" {
			t.Fatalf("Slugify(%q)=%q has a character outside [a-z0-9.-]", s, out)
		}
		if strings.HasPrefix(out, "-") || strings.HasSuffix(out, "-") || strings.Contains(out, "--") {
			t.Fatalf("Slugify(%q)=%q has a leading/trailing/doubled hyphen", s, out)
		}
		if IsSlug(out) != (out != "") {
			t.Fatalf("IsSlug(Slugify(%q)=%q) = %v", s, out, IsSlug(out))
		}
		IsHubPluginCoordinate(s) // must not panic on arbitrary input
	})
}
