// SPDX-License-Identifier: Apache-2.0

// Package staging is the staging-repository rule. Hub ADR-0121 owns it (it
// replaced the `<ns>/staging/<rest>` form of hub ADR-0103 point 6).
// Publishers always push a bundle to the staging form of the served
// repository it is published to; the hub checks it there and copies an
// accepted bundle into the served repository. The hub, grcli and pvtr compute
// the name with these functions, so they agree on where a bundle goes.
//
// The staging form puts "staging" first so that zot routes every staging
// repository to its own sub-store, which collects its own garbage. "staging"
// is therefore a reserved namespace slug on the hub.
package staging

import "strings"

// Segment is the reserved first path segment of a staging repository.
const Segment = "staging"

// Repository returns the staging repository for served repository
// "<ns>/<rest>": "staging/<ns>/<rest>". ok is false when served has fewer
// than two segments, has an empty segment, or has "staging" as its
// namespace (which also covers a served that is itself a staging
// repository).
func Repository(served string) (string, bool) {
	if !valid(served) {
		return "", false
	}
	return Segment + "/" + served, true
}

// IsStaging reports whether repo is a staging repository: "staging" as its
// first segment, followed by a non-empty namespace and a non-empty rest.
// A repository with "staging" in any other position is an ordinary served
// repository.
func IsStaging(repo string) bool {
	_, ok := Served(repo)
	return ok
}

// Served is the inverse of Repository: "staging/<ns>/<rest>" gives
// "<ns>/<rest>". ok is false when repo is not a staging repository.
func Served(repo string) (string, bool) {
	served, found := strings.CutPrefix(repo, Segment+"/")
	if !found || !valid(served) {
		return "", false
	}
	return served, true
}

// valid reports whether served is "<ns>/<rest>" with no empty segment and a
// namespace other than Segment.
func valid(served string) bool {
	segs := strings.Split(served, "/")
	if len(segs) < 2 || segs[0] == Segment {
		return false
	}
	for _, s := range segs {
		if s == "" {
			return false
		}
	}
	return true
}
