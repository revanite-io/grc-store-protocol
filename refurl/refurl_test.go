// SPDX-License-Identifier: Apache-2.0

package refurl

import "testing"

// TestParse is the contract table. The hub's ResolveReferenceURL wrapper, grcli
// refs.Recognize, and the frontend's resolveHubRefURL (src/lib/hub.ts) are
// each expected to agree with every row here.
func TestParse(t *testing.T) {
	cases := []struct {
		name string
		url  string
		ok   bool
		want Ref
	}{
		{"hub api with version", "https://hub.grc.store/v1/catalogs/acme/base/versions/2.0", true, Ref{"acme", "base", "2.0"}},
		{"hub api no version", "https://hub.grc.store/v1/catalogs/acme/base", true, Ref{"acme", "base", ""}},
		{"hub api sub-resource", "https://hub.grc.store/v1/catalogs/acme/base/referenced-by", true, Ref{"acme", "base", ""}},
		{"self-hosted api host-agnostic", "https://compliance.acme.internal/v1/catalogs/acme/base", true, Ref{"acme", "base", ""}},
		{"ui form on grc.store", "https://grc.store/acme/widget", true, Ref{"acme", "widget", ""}},
		{"ui form on a grc.store subdomain with port", "https://HUB.grc.store:443/acme/widget", true, Ref{"acme", "widget", ""}},
		{"ui form with version suffix is not canonical (ruled 2026-09-18)", "https://grc.store/acme/widget/versions/1.0", false, Ref{}},
		{"legacy search form", "https://grc.store/search/finos-aigf/finos-air", true, Ref{"finos-aigf", "finos-air", ""}},
		{"legacy search form with version", "https://grc.store/search/finos-aigf/finos-air/versions/0.2.0", true, Ref{"finos-aigf", "finos-air", "0.2.0"}},
		{"legacy search with one segment", "https://grc.store/search/only", false, Ref{}},
		{"legacy search off grc.store", "https://example.org/search/finos-aigf/finos-air", false, Ref{}},
		{"slugifies segments", "https://hub.grc.store/v1/catalogs/ACME/Base_Guidance", true, Ref{"acme", "base-guidance", ""}},
		{"percent-encoded and mixed case", "https://grc.store/FINOS%20CCC/CCC.ObjStor.CP", true, Ref{"finos-ccc", "ccc.objstor.cp", ""}},
		{"segments that slugify to nothing", "https://grc.store/---/baseline", false, Ref{}},
		{"github blob external", "https://github.com/complytime/complytime-policies/blob/main/x.yaml", false, Ref{}},
		{"two-seg non grc.store", "https://github.com/owner/repo", false, Ref{}},
		{"pdf external", "https://nvlpubs.nist.gov/nistpubs/CSWP/NIST.CSWP.29.pdf", false, Ref{}},
		{"one segment on grc.store", "https://grc.store/only-one-segment", false, Ref{}},
		{"not a url", "not a url", false, Ref{}},
		{"empty", "", false, Ref{}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, ok := Parse(c.url)
			if ok != c.ok {
				t.Fatalf("ok = %v, want %v (got %+v)", ok, c.ok, got)
			}
			if got != c.want {
				t.Errorf("got %+v, want %+v", got, c.want)
			}
		})
	}
}

func TestIsGrcStoreHost(t *testing.T) {
	for host, want := range map[string]bool{
		"grc.store": true, "GRC.STORE": true, "hub.grc.store": true, "hub.preview.grc.store:8443": true,
		"grc.store.evil.example": false, "notgrc.store": false, "github.com": false, "": false,
	} {
		if got := IsGrcStoreHost(host); got != want {
			t.Errorf("IsGrcStoreHost(%q) = %v, want %v", host, got, want)
		}
	}
}
