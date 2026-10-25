// SPDX-License-Identifier: Apache-2.0

// Package targetgroup is the wire contract for target groups (hub ADR-0091):
// a namespace's named, ordered set of its own targets, and the live view of
// where each member's evaluations stand.
//
// Routes, all under one namespace:
//
//	GET    /v1/targets/{ns}/groups[?limit=&after=]  → ListResponse (archived groups omitted)
//	GET    /v1/targets/{ns}/groups/{slug}           → View (archived groups still read)
//	POST   /v1/targets/{ns}/groups/{slug}           UpsertRequest → Group (201 created, 200 replaced)
//	DELETE /v1/targets/{ns}/groups/{slug}           → 204 (archive; the slug stays taken)
//
// Reads follow the namespace viewer rule: a group in a namespace hidden
// from the caller reads exactly like a missing one. Writes are for human
// members of the namespace only; a CI principal gets 403
// apierror.Forbidden. A POST to an archived group's slug is 409
// apierror.TargetGroupArchived. A member that is not a target in {ns}, or
// an archived target that is not already a member, is 422
// apierror.TargetGroupMemberInvalid. The POST body is capped at
// limits.MaxTargetGroupRequestBytes (413 apierror.TargetGroupTooLarge).
// Neither members per group nor groups per namespace are capped (Q16).
// A malformed request (a slug that is not a hub slug, an empty name, a
// member listed twice) is 400 validation_failed.
//
// Every array in these shapes is always an array, never null. A nil Go
// slice marshals as null, so producers must set each one to an empty slice.
//
// A group names no catalog and has no result of its own. Each member's
// evaluations are listed as they are, one per (catalog, evaluator plugin),
// each showing its newest live run's own result.
package targetgroup

import "time"

// ReservedTargetID is the target id the group routes take: no target may
// be named "groups", so /v1/targets/{ns}/groups is never a target.
const ReservedTargetID = "groups"

// IsReservedTargetID reports whether id is a path segment under
// /v1/targets/{ns}/ that no target may take: "artifacts" and "admin" for
// the hub's other target routes, and ReservedTargetID for groups.
func IsReservedTargetID(id string) bool {
	switch id {
	case "artifacts", "admin", ReservedTargetID:
		return true
	}
	return false
}

// Group is a target group's definition.
type Group struct {
	Namespace   string `json:"namespace"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Members in the owner's order. Always an array, never null.
	Members    []Member   `json:"members"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Member is one member target of a group, by its id in the group's
// namespace. ArchivedAt is set when the target has been archived since it
// joined; it stays a member until a human removes it.
type Member struct {
	TargetID   string     `json:"target_id"`
	ArchivedAt *time.Time `json:"archived_at,omitempty"`
}

// UpsertRequest is the POST body: it creates the group or replaces its
// definition, membership included. Members are target ids in the group's
// namespace, in display order, each at most once. An empty list is a
// valid group with no members yet. A missing or null members means the
// same: the POST replaces membership, so it leaves the group with none.
type UpsertRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Members     []string `json:"members"`
}

// ListResponse is the body of GET /v1/targets/{ns}/groups, ordered by slug.
type ListResponse struct {
	Items      []Group `json:"items"`
	NextCursor *string `json:"next_cursor"`
}

// View is the body of GET /v1/targets/{ns}/groups/{slug}: the definition
// and the live view, one MemberView per member in the group's order.
type View struct {
	Group   Group        `json:"group"`
	Members []MemberView `json:"members"`
}

// MemberView is one member target and its evaluations. An archived member
// carries ArchivedAt and no evaluations. A member with no published log
// has an empty Evaluations array.
type MemberView struct {
	TargetID    string           `json:"target_id"`
	Name        string           `json:"name"`
	URI         string           `json:"uri"`
	VerifiedAt  *time.Time       `json:"verified_at"` // null = unverified
	ArchivedAt  *time.Time       `json:"archived_at,omitempty"`
	Evaluations []EvaluationView `json:"evaluations"`
}

// EvaluationView is one evaluation of a member: its logs sharing one
// catalog reference and one evaluator plugin, shown by the newest run that
// is neither yanked nor hidden by moderation, across every target version.
// Result and the counts are the log's own claims, stored verbatim.
type EvaluationView struct {
	CatalogRefID     string  `json:"catalog_ref_id"`
	CatalogNamespace *string `json:"catalog_namespace"` // null = the catalog is off-hub
	CatalogID        *string `json:"catalog_id"`        // null iff catalog_namespace is null
	// PluginDeclaredCatalogVersion is the catalog version the evaluator
	// plugin version declares it evaluates, when it declares exactly one;
	// null means the version is not recorded. The log itself records none.
	PluginDeclaredCatalogVersion *string `json:"plugin_declared_catalog_version"`

	EvaluatorPlugin EvaluatorPlugin `json:"evaluator_plugin"`

	TargetVersion string    `json:"target_version"`
	RunAt         time.Time `json:"run_at"`

	// The log artifact's coordinate; pull it for per-requirement detail.
	LogNamespace      string `json:"log_namespace"`
	LogID             string `json:"log_id"`
	LogVersion        string `json:"log_version"`
	LogManifestDigest string `json:"log_manifest_digest"`

	Result                string `json:"result"` // the log's Result wire string
	Counts                Counts `json:"counts"`
	EvaluatedRequirements int    `json:"evaluated_requirements"`
	TotalRequirements     int    `json:"total_requirements"`
}

// EvaluatorPlugin is the published plugin version the run's signed
// provenance resolved to.
type EvaluatorPlugin struct {
	Namespace string `json:"namespace"`
	PluginID  string `json:"plugin_id"`
	Version   string `json:"version"`
}

// Counts are the run's assessment results by Result value.
type Counts struct {
	Passed        int `json:"passed"`
	Failed        int `json:"failed"`
	NeedsReview   int `json:"needs_review"`
	NotRun        int `json:"not_run"`
	NotApplicable int `json:"not_applicable"`
	Unknown       int `json:"unknown"`
}
