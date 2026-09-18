// SPDX-License-Identifier: Apache-2.0

// Package refurl is THE rule for "does this mapping-reference url name an
// artifact on a grc.store hub, and which one". The hub's reference index
// (ADR-0040; also the EvaluationLog ingest gate of ADR-0053), grcli
// unpack --with-references, and the web UI's link resolver all make this
// decision, and three private copies disagreed (REV-382), so the rule lives
// here. The frontend keeps a TypeScript mirror pinned to this package's
// table test.
//
// Shapes that resolve, with every coordinate segment slugified by the slug
// package so the result is the coordinate the hub actually indexes:
//
//   - the hub API path .../v1/catalogs/{ns}/{id}[/versions/{v}][/...] on ANY
//     host: distinctive enough not to false-match, and host-agnostic so a
//     self-hosted hub's references resolve without this package knowing
//     every deployment's domain;
//   - the UI path /{ns}/{id} on a grc.store host only (a two-segment path is
//     otherwise indistinguishable from github.com/{owner}/{repo}), with NO
//     version suffix — ruled 2026-09-18: the version of a reference belongs
//     in mapping-references[].version, and the canonical form the docs
//     teach is exactly https://grc.store/<ns>/<id>;
//   - the legacy UI path /search/{ns}/{id}[/versions/{v}] on a grc.store
//     host: the address the site handed out before the org-scoped scheme,
//     still carried by published catalogs.
//
// Which hub to FETCH from is not decided here: the hub indexes what it
// stores, and a client that pulls bytes applies its own host policy on top.
package refurl

import (
	"net/url"
	"strings"

	"github.com/revanite-io/grc-store-protocol/slug"
)

// Ref is a hub coordinate read out of a reference url. Version is set only
// when the url itself carried one and is informational: the authoritative
// version of a reference is mapping-references[].version.
type Ref struct {
	Namespace string
	CatalogID string
	Version   string
}

// Parse reads a hub coordinate out of rawURL, or reports false when the url
// does not name a hub artifact (an external standard, a PDF, a docs page).
func Parse(rawURL string) (Ref, bool) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return Ref{}, false
	}
	var segs []string
	for _, s := range strings.Split(u.Path, "/") {
		if s != "" {
			segs = append(segs, s)
		}
	}

	// API shape, any host. Trailing segments after the coordinate are
	// ignored, so a sub-resource url (…/referenced-by) still names the artifact.
	for i := 0; i+3 < len(segs); i++ {
		if segs[i] == "v1" && segs[i+1] == "catalogs" {
			version := ""
			if rest := segs[i+4:]; len(rest) >= 2 && rest[0] == "versions" {
				version = rest[1]
			}
			if r, ok := ref(segs[i+2], segs[i+3], version); ok {
				return r, true
			}
		}
	}

	if !IsGrcStoreHost(u.Host) {
		return Ref{}, false
	}
	if len(segs) > 0 && segs[0] == "search" { // legacy UI shape
		rest := segs[1:]
		switch {
		case len(rest) == 2:
			return ref(rest[0], rest[1], "")
		case len(rest) == 4 && rest[2] == "versions":
			return ref(rest[0], rest[1], rest[3])
		}
		return Ref{}, false
	}
	if len(segs) == 2 { // UI shape, exactly
		return ref(segs[0], segs[1], "")
	}
	return Ref{}, false
}

func ref(ns, id, version string) (Ref, bool) {
	r := Ref{Namespace: slug.Slugify(ns), CatalogID: slug.Slugify(id), Version: version}
	if r.Namespace == "" || r.CatalogID == "" {
		return Ref{}, false
	}
	return r, true
}

// IsGrcStoreHost reports whether host is grc.store or a subdomain of it
// (hub.grc.store, hub.preview.grc.store, ...), case- and port-insensitively.
func IsGrcStoreHost(host string) bool {
	h := strings.ToLower(host)
	if i := strings.IndexByte(h, ':'); i >= 0 {
		h = h[:i]
	}
	return h == "grc.store" || strings.HasSuffix(h, ".grc.store")
}
