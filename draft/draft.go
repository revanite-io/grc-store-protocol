// SPDX-License-Identifier: Apache-2.0

// Package draft is the wire contract for hub-held drafts (hub ADR-0060): an
// unpublished, unsigned working copy of an artifact that the web editors save
// to the hub so it can be reopened from any browser. A draft is not an
// artifact: it is never indexed, never enters the registry and needs no
// signature. Publishing stays a separate, signed step.
//
// Routes, all namespace-scoped and open to any member of the namespace:
//
//	GET    /v1/namespaces/{slug}/drafts[?type=<artifact type>]  → ListResponse (no bodies)
//	POST   /v1/namespaces/{slug}/drafts                          SaveRequest → Draft (201)
//	GET    /v1/namespaces/{slug}/drafts/{id}                     → Draft
//	PUT    /v1/namespaces/{slug}/drafts/{id}                     SaveRequest → Draft (no body echoed)
//	DELETE /v1/namespaces/{slug}/drafts/{id}                     → 204
//
// The body is the editor's own state, opaque to the hub: any JSON object up
// to limits.MaxDraftBodyBytes. The hub stores it as jsonb and never reads
// into it — key order and whitespace come back normalised, and an escaped
// NUL is refused (422) — and the title travels beside it so a list needs no
// parsing. A PUT that names ExpectedUpdatedAt is refused with 409
// apierror.DraftConflict when the row has moved on, so two tabs cannot
// silently overwrite each other.
package draft

import (
	"bytes"
	"encoding/json"
	"time"
)

// Artifact types the hub drafts today — the ones with a web editor. The
// assist package aliases these so the AI tables and the draft store agree;
// a draftable type need not have AI assist (CapabilityCatalog has none yet).
const (
	ArtifactTypeControlCatalog    = "ControlCatalog"
	ArtifactTypeThreatCatalog     = "ThreatCatalog"
	ArtifactTypeGuidanceCatalog   = "GuidanceCatalog"
	ArtifactTypeCapabilityCatalog = "CapabilityCatalog"
)

var artifactTypes = []string{ArtifactTypeControlCatalog, ArtifactTypeThreatCatalog, ArtifactTypeGuidanceCatalog, ArtifactTypeCapabilityCatalog}

// ArtifactTypes lists the draftable types in a stable order, for messages.
func ArtifactTypes() []string { return append([]string(nil), artifactTypes...) }

// ValidArtifactType reports whether artifactType has a web editor.
func ValidArtifactType(artifactType string) bool {
	for _, t := range artifactTypes {
		if t == artifactType {
			return true
		}
	}
	return false
}

// ValidBody reports whether raw is a JSON object — the only body shape a
// draft may carry. An empty object is a valid, empty draft.
func ValidBody(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) > 0 && trimmed[0] == '{' && json.Valid(trimmed)
}

// Draft is one hub-held draft. Body is omitted from list responses; a
// single-draft response always carries it (an empty draft's body is {}).
type Draft struct {
	ID                string          `json:"id"`
	Namespace         string          `json:"namespace"`
	ArtifactType      string          `json:"artifact_type"`
	Title             string          `json:"title"`
	Body              json.RawMessage `json:"body,omitempty"`
	CreatedBy         string          `json:"created_by"`
	CreatedByUsername string          `json:"created_by_username"`
	UpdatedBy         string          `json:"updated_by"`
	UpdatedByUsername string          `json:"updated_by_username"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// SaveRequest is the body of POST (create) and PUT (replace). On create
// ArtifactType is required and fixed for the draft's life; on update it is
// ignored. Body must satisfy ValidBody. ExpectedUpdatedAt, when set on a
// PUT, must equal the draft's current UpdatedAt or the hub answers 409
// apierror.DraftConflict and writes nothing.
type SaveRequest struct {
	ArtifactType      string          `json:"artifact_type,omitempty"`
	Title             string          `json:"title"`
	Body              json.RawMessage `json:"body"`
	ExpectedUpdatedAt *time.Time      `json:"expected_updated_at,omitempty"`
}

// ListResponse is the body of GET /v1/namespaces/{slug}/drafts: newest
// updated first, bodies omitted.
type ListResponse struct {
	Items []Draft `json:"items"`
}
