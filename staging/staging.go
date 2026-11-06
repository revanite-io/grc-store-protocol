// SPDX-License-Identifier: Apache-2.0

// Package staging is the staging-repository rule of hub ADR-0103 point 6.
// Publishers push a bundle to the staging form of the served repository it is
// published to; the hub checks it there and copies an accepted bundle into the
// served repository. The hub, grcli and pvtr compute the name with this one
// function, so they agree on where a bundle goes.
//
// A client pushes to staging only when the hub's discovery document sets
// staging_push (discovery.Document.StagingPush); otherwise it pushes to the
// served repository, as hubs without staging expect.
package staging

import "strings"

// Segment is the reserved second path segment of a staging repository.
const Segment = "staging"

// Repository returns the staging repository for served repository
// "<ns>/<rest>": "<ns>/staging/<rest>". ok is false when served has fewer
// than two segments, has an empty segment, or is itself a staging repository.
func Repository(served string) (string, bool) {
	ns, rest, found := strings.Cut(served, "/")
	if !found || ns == "" || rest == "" || IsStaging(served) {
		return "", false
	}
	for _, seg := range strings.Split(rest, "/") {
		if seg == "" {
			return "", false
		}
	}
	return ns + "/" + Segment + "/" + rest, true
}

// IsStaging reports whether repo is a staging repository: one whose second
// segment is exactly "staging" and that has a segment after it. A repository
// with "staging" in any other position is an ordinary served repository.
func IsStaging(repo string) bool {
	segs := strings.Split(repo, "/")
	return len(segs) >= 3 && segs[0] != "" && segs[1] == Segment
}

// Served returns the served repository a staging repository stands for:
// "<ns>/staging/<rest>" gives "<ns>/<rest>". ok is false when repo is not a
// staging repository.
func Served(repo string) (string, bool) {
	if !IsStaging(repo) {
		return "", false
	}
	ns, rest, _ := strings.Cut(repo, "/")
	return ns + "/" + strings.TrimPrefix(rest, Segment+"/"), true
}
