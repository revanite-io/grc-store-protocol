// SPDX-License-Identifier: Apache-2.0

// Package assist is the wire contract for AI-assisted drafting on the hub
// (hub ADR-0059): the per-user provider credential at /v1/me/ai-credential
// and the polish call at POST /v1/ai/polish. The hub holds the credential and
// makes every provider call itself; a client only ever sends the token in,
// never reads it back, and sends draft context to be generated from or
// polished. Prompt templates are the hub's, not the client's.
package assist

import "encoding/json"

// ProviderAnthropic is the one provider accepted today. The provider field
// exists from day one so a second one is additive.
const ProviderAnthropic = "anthropic"

// ValidProvider reports whether p names a provider the contract knows.
func ValidProvider(p string) bool { return p == ProviderAnthropic }

// CredentialRequest is the body of PUT /v1/me/ai-credential. The token is
// write-only: no response ever carries it.
type CredentialRequest struct {
	Provider string `json:"provider"`
	Token    string `json:"token"`
}

// CredentialStatus is the body of GET /v1/me/ai-credential (404 when none is
// set). Set is always true on a 200; it is here so a client can treat the
// 404 and the 200 uniformly. UpdatedAt is RFC 3339.
type CredentialStatus struct {
	Provider  string `json:"provider"`
	Set       bool   `json:"set"`
	UpdatedAt string `json:"updated_at"`
}

// Artifact types, fields and modes a polish request may name. The hub owns
// one prompt template per (artifact type, field, mode).
const (
	ArtifactTypeControlCatalog = "ControlCatalog"

	FieldDescription      = "description"       // ControlCatalog metadata.description
	FieldControlObjective = "control.objective" // one control's objective

	ModeGenerate = "generate" // Current is empty: write the field from Context
	ModePolish   = "polish"   // Current has text: improve it, keeping its meaning
)

// ValidField reports whether field is one the hub has a template for.
func ValidField(field string) bool {
	return field == FieldDescription || field == FieldControlObjective
}

// ValidMode reports whether mode is generate or polish.
func ValidMode(mode string) bool { return mode == ModeGenerate || mode == ModePolish }

// PolishRequest is the body of POST /v1/ai/polish. Context's shape depends
// on Field: DescriptionContext for FieldDescription, ObjectiveContext for
// FieldControlObjective. Nothing in it is persisted or logged by the hub.
type PolishRequest struct {
	ArtifactType string          `json:"artifact_type"`
	Field        string          `json:"field"`
	Mode         string          `json:"mode"`
	Current      string          `json:"current"`
	Context      json.RawMessage `json:"context"`
}

// PolishResponse carries the generated or polished text. Note is a short
// status for the author, present only when the hub has something to say
// about how the text was produced (written from the title alone, an
// objective that claims more than its requirements, and so on).
type PolishResponse struct {
	Text string `json:"text"`
	Note string `json:"note,omitempty"`
}

// ControlSummary is a control as the description template and the
// objective template's sibling list see it. Objective may be empty in a
// DescriptionContext; a sibling in an ObjectiveContext always has one.
type ControlSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Objective string `json:"objective,omitempty"`
}

// Group is a control group as the description template sees it.
type Group struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

// DescriptionContext is PolishRequest.Context for FieldDescription: the
// catalog title, its groups and its controls at objective altitude.
// Requirements are deliberately not sent for this field.
type DescriptionContext struct {
	Title    string           `json:"title"`
	Groups   []Group          `json:"groups,omitempty"`
	Controls []ControlSummary `json:"controls,omitempty"`
}

// Requirement is one assessment requirement of the control being written.
type Requirement struct {
	ID            string   `json:"id"`
	Text          string   `json:"text"`
	Applicability []string `json:"applicability,omitempty"`
}

// ObjectiveContext is PolishRequest.Context for FieldControlObjective. The
// control's own requirements are the content source; Siblings are every
// other control with a non-blank objective and are the style pattern, not
// content. The control being written never appears in Siblings.
type ObjectiveContext struct {
	Catalog struct {
		Title       string `json:"title"`
		Description string `json:"description,omitempty"`
	} `json:"catalog"`
	Control struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Group string `json:"group,omitempty"`
	} `json:"control"`
	Requirements []Requirement    `json:"requirements,omitempty"`
	Siblings     []ControlSummary `json:"siblings,omitempty"`
}
