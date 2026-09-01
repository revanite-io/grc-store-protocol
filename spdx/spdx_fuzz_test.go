// SPDX-License-Identifier: Apache-2.0

package spdx

import "testing"

// FuzzParse: the grammar is fed publisher-controlled strings at both the hub
// (lenient path) and grcli (strict path). Invariants: never panic; a parsed
// expression's String() re-parses to the same String() (the canonical form
// is a fixed point). Crashers saved under testdata/fuzz/FuzzParse/ are regression
// seeds — commit them.
func FuzzParse(f *testing.F) {
	for _, s := range []string{
		"MIT", "MIT OR Apache-2.0", "mit or apache-2.0",
		"(MIT AND BSD-3-Clause) OR GPL-2.0-only WITH Classpath-exception-2.0",
		"LicenseRef-Revanite-Proprietary", "DocumentRef-x:LicenseRef-y",
		"MIT AND AND", "", "(", ")", "MIT+", "MIT WITH", "WITH MIT", "((MIT))", "MIT OR", " MIT ",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, expr string) {
		e, err := Parse(expr)
		if err != nil {
			return
		}
		s := e.String()
		e2, err := Parse(s)
		if err != nil {
			t.Fatalf("Parse(%q) ok but its String() %q does not re-parse: %v", expr, s, err)
		}
		if got := e2.String(); got != s {
			t.Fatalf("String() is not a fixed point: %q -> %q -> %q", expr, s, got)
		}
	})
}
