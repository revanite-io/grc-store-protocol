// SPDX-License-Identifier: Apache-2.0

package assist

import (
	"encoding/json"

	"github.com/revanite-io/grc-store-protocol/draft"
)

// The AI draft route (hub ADR-0110, ADR-0111, ADR-0112): POST /v1/ai/draft
// writes one whole Gemara document from questionnaire answers as a job the
// client polls, and its outcome waits on the hub until the owner accepts or
// discards it.

// draftUpstream is the draft table: the six types the route writes and the
// one upstream document type each takes, "" for answers only (hub ADR-0110
// point 1).
var draftUpstream = map[string]string{
	draft.ArtifactTypeGuidanceCatalog:   "",
	draft.ArtifactTypeCapabilityCatalog: draft.ArtifactTypeGuidanceCatalog,
	draft.ArtifactTypeThreatCatalog:     draft.ArtifactTypeCapabilityCatalog,
	draft.ArtifactTypeRiskCatalog:       "",
	draft.ArtifactTypeControlCatalog:    draft.ArtifactTypeThreatCatalog,
	draft.ArtifactTypePolicy:            draft.ArtifactTypeRiskCatalog,
}

var draftTypes = func() []string {
	var out []string
	for _, t := range draft.ArtifactTypes() {
		if _, ok := draftUpstream[t]; ok {
			out = append(out, t)
		}
	}
	return out
}()

// ValidDraftType reports whether POST /v1/ai/draft writes artifactType.
func ValidDraftType(artifactType string) bool { _, ok := draftUpstream[artifactType]; return ok }

// DraftTypes lists the types POST /v1/ai/draft writes, in a stable order.
func DraftTypes() []string { return append([]string(nil), draftTypes...) }

// DraftUpstream is the type of the upstream document a draft of
// artifactType must send, or "" when it takes answers only (and for a type
// the route does not write).
func DraftUpstream(artifactType string) string { return draftUpstream[artifactType] }

// DraftRequest is the body of POST /v1/ai/draft. Upstream and Previous are
// Gemara documents as JSON. Upstream is required exactly when
// DraftUpstream(ArtifactType) is not "". References is for a Policy only.
// The answers are capped by limits.MaxAIDraftAnswersBytes and the body by
// limits.MaxAIDraftRequestBytes. Nothing in it is stored or logged by the
// hub, except the coordinates it writes into the result.
type DraftRequest struct {
	Namespace    string           `json:"namespace"`
	ArtifactType string           `json:"artifact_type"`
	Target       DraftTarget      `json:"target"`
	Answers      []DraftAnswer    `json:"answers"`
	Upstream     json.RawMessage  `json:"upstream,omitempty"`
	References   []DraftReference `json:"references,omitempty"`
	Previous     json.RawMessage  `json:"previous,omitempty"`
}

// DraftTarget is the coordinate the document will be published at; the hub
// writes it into metadata.id and metadata.version.
type DraftTarget struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

// DraftAnswer is one questionnaire answer, with its question's text.
type DraftAnswer struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// DraftReference is another catalog a Policy must reference: its intended
// publish coordinate and title. All five fields are required.
type DraftReference struct {
	ArtifactType string `json:"artifact_type"`
	Namespace    string `json:"namespace"`
	ID           string `json:"id"`
	Version      string `json:"version"`
	Title        string `json:"title"`
}

// DraftStarted is the 202 body of POST /v1/ai/draft. JobID is a UUID.
type DraftStarted struct {
	JobID string `json:"job_id"`
}

// Job states. Running until the job ends in one of the three outcomes.
const (
	DraftStateRunning   = "running"
	DraftStateSucceeded = "succeeded"
	DraftStatePartial   = "partial"
	DraftStateFailed    = "failed"
)

// Job stages, set while the state is running.
const (
	DraftStageGenerating = "generating" // a fresh generation call
	DraftStageRepairing  = "repairing"  // a repair call
	DraftStageValidating = "validating" // the hub validates or prunes
)

// DraftJob is the body of GET /v1/ai/draft/{job_id}. Attempt is the number
// of provider calls started, 0 to 3. Result is set on succeeded and partial,
// Removed on partial, Error on failed; Usage once the job has ended.
type DraftJob struct {
	State   string          `json:"state"`
	Stage   string          `json:"stage,omitempty"`
	Attempt int             `json:"attempt"`
	Result  json.RawMessage `json:"result,omitempty"`
	Removed []DraftRemoved  `json:"removed,omitempty"`
	Error   *DraftError     `json:"error,omitempty"`
	Usage   *DraftUsage     `json:"usage,omitempty"`
}

// DraftRemoved is one thing a partial result cut out. Path is its location
// in the model's output as returned (for example "guidelines.3"), ID its id
// when that is a string, and Errors the validator lines that pointed into
// it; a dropped reference has the removed id it named as its only error.
type DraftRemoved struct {
	Path   string   `json:"path"`
	ID     string   `json:"id,omitempty"`
	Errors []string `json:"errors"`
}

// DraftError is a failed job's error. Code is an apierror code.
type DraftError struct {
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

// DraftUsage is the provider's token counts summed over every call in the job.
type DraftUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// DraftListItem is one element of GET /v1/ai/draft, a JSON array of the
// caller's unresolved jobs, newest first. It carries no result. CreatedAt is
// RFC 3339.
type DraftListItem struct {
	JobID        string `json:"job_id"`
	Namespace    string `json:"namespace"`
	ArtifactType string `json:"artifact_type"`
	State        string `json:"state"`
	CreatedAt    string `json:"created_at"`
}
