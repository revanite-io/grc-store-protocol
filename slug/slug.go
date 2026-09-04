// SPDX-License-Identifier: Apache-2.0

// Package slug is THE hub coordinate slug rule — the one function that turns a
// free-form identifier (metadata.author.id, metadata.id, a target id) into the
// namespace / artifact-id segment the hub indexes it under and the UI links it
// by. It is a byte-for-byte port of the hub's internal/store.Slugify; the hub,
// grcli, and pvtr must all compute the same coordinate for the same input, and
// a second implementation of this rule has already drifted once (ADR-0050 — grcli
// kept '_' and used a different trim set, and the hub had to normalize around
// it). Import this; do not re-derive it.
//
// The rule is frozen contract: changing Slugify's output for any input moves
// every coordinate computed from that input and is a BREAKING change.
package slug

import "strings"

// Slugify produces an OCI-and-URL-safe lowercase identifier from a free-form
// input. Rules, in order:
//
//   - Lowercase the input (strings.ToLower, rune-wise).
//   - Keep [a-z0-9.]; replace every other run of runes with a single '-'.
//   - Strip leading and trailing '-'.
//   - An input with nothing to keep returns "" (callers validate non-empty).
//
// '.' is preserved because ids like "ccc.iam.cn" carry it meaningfully;
// collapsing dots would erase the namespace-in-id encoding real catalogs use.
// '_' is NOT preserved — "target_catalog" slugifies to "target-catalog".
func Slugify(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevHyphen := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '.':
			b.WriteRune(r)
			prevHyphen = false
		default:
			if !prevHyphen {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// IsSlug reports whether s is already in canonical slug form: non-empty and a
// fixed point of Slugify. A coordinate segment the hub will accept verbatim
// satisfies IsSlug; anything else is either rejected or re-slugified depending
// on the path, so producers should send slugs, not raw ids.
func IsSlug(s string) bool {
	return s != "" && Slugify(s) == s
}
