// SPDX-License-Identifier: Apache-2.0

package slug

import (
	"strings"
	"time"
)

// EvaluationLog results coordinate scheme.
//
// A published EvaluationLog lives at
//
//	<target-namespace>/<slug(target-id + "_" + catalog-id)>:<target-version>-<run-at>
//
// The repository's second segment is one STREAM per (target, catalog): every
// run of any evaluator against that pair publishes into it, so a multi-plugin
// run never collides and the hub's TOFU signer pin (ADR-0045) is per stream.
// The '_' is cosmetic — Slugify turns it into '-' on the wire — and the hub
// never parses the id back apart; it links a log to its target and catalog via
// the log's own `target` and mapping references.

// EvaluationLogRepository returns the OCI repository (no registry host) an
// EvaluationLog for targetID evaluated against catalogID is published to under
// the target owner's namespace. All three inputs are slugified so a producer
// holding raw ids and one holding slugs compute the same repository.
func EvaluationLogRepository(targetNamespace, targetID, catalogID string) string {
	return Slugify(targetNamespace) + "/" + Slugify(targetID+"_"+catalogID)
}

// EvaluationLogVersionTimeLayout is the run-timestamp layout embedded in an
// EvaluationLog's version/tag: basic ISO-8601, UTC, second precision, 'Z'
// suffix. No ':' — those are not OCI-tag-legal.
const EvaluationLogVersionTimeLayout = "20060102T150405Z"

// EvaluationLogVersion returns the per-run version an EvaluationLog is
// published as: metadata.version == the OCI tag ==
// "<targetVersion>-<runAt as EvaluationLogVersionTimeLayout, UTC>". Tags sort
// lexically by run time within a target version; a same-second re-run
// collides with the immutable tag (ADR-0033) and must be re-timestamped by
// the producer.
//
// The caller passes the target version verbatim; note that a target version
// carrying an "-rc" prerelease (e.g. "2.0.0-rc1") makes every log for it a
// release candidate under the hub's -rc* convention (ADR-0041).
func EvaluationLogVersion(targetVersion string, runAt time.Time) string {
	return targetVersion + "-" + runAt.UTC().Format(EvaluationLogVersionTimeLayout)
}

// IsHubPluginCoordinate reports whether an EvaluationLog's metadata.author.id
// names a hub plugin coordinate — "<namespace>/<plugin-id>" — and therefore
// MUST resolve to a published plugin at the provenance-declared digest for the
// hub to accept the log (the conditionally fail-closed producer
// gate). The rule is exact, so it cannot become an accidental bypass:
//
//	exactly two '/'-separated segments, each non-empty and IsSlug.
//
// So "local/my-plugin" IS a coordinate (and fails closed when unpublished);
// "acme-scanner", "https://…", "Acme/Scanner" (uppercase) and "a/b/c" are NOT,
// and are accepted with the producer marked unverified.
func IsHubPluginCoordinate(authorID string) bool {
	ns, id, found := strings.Cut(authorID, "/")
	return found && IsSlug(ns) && IsSlug(id) // IsSlug(id) fails on a further '/'
}
