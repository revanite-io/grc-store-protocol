// SPDX-License-Identifier: Apache-2.0

// Package assist is the wire contract for AI-assisted drafting on the hub
// (hub ADR-0059): the per-user provider credential at /v1/me/ai-credential,
// the polish call at POST /v1/ai/polish and the review call at POST
// /v1/ai/review, for the control, threat and guidance catalog editors. The hub
// holds the credential and makes every provider call itself; a client only
// ever sends the token in, never reads it back, and sends draft context to
// be generated from or polished. Prompt templates are the hub's, not the
// client's.
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
	ArtifactTypeControlCatalog  = "ControlCatalog"
	ArtifactTypeThreatCatalog   = "ThreatCatalog"
	ArtifactTypeGuidanceCatalog = "GuidanceCatalog"

	FieldDescription           = "description"                 // the catalog's metadata.description, every type
	FieldControlObjective      = "control.objective"           // ControlCatalog: one control's objective
	FieldAssessmentRequirement = "assessment-requirement.text" // ControlCatalog: one assessment requirement's text
	FieldThreatDescription     = "threat.description"          // ThreatCatalog: one threat's description
	FieldGuidelineObjective    = "guideline.objective"         // GuidanceCatalog: one guideline's objective
	FieldStatementText         = "statement.text"              // GuidanceCatalog: one statement's text

	ModeGenerate = "generate" // Current is empty: write the field from Context
	ModePolish   = "polish"   // Current has text: improve it, keeping its meaning
)

// fields is the template table: which fields each artifact type has.
var fields = map[string][]string{
	ArtifactTypeControlCatalog:  {FieldDescription, FieldControlObjective, FieldAssessmentRequirement},
	ArtifactTypeThreatCatalog:   {FieldDescription, FieldThreatDescription},
	ArtifactTypeGuidanceCatalog: {FieldDescription, FieldGuidelineObjective, FieldStatementText},
}

// reviews is the review table: which review kinds each artifact type has.
// A threat has no objective/requirements pair, so no coverage review; a
// guideline's objective/statements pair gets one like a control's.
var reviews = map[string][]string{
	ArtifactTypeControlCatalog:  {ReviewDuplicates, ReviewCoverage, ReviewEntries},
	ArtifactTypeThreatCatalog:   {ReviewDuplicates, ReviewEntries},
	ArtifactTypeGuidanceCatalog: {ReviewDuplicates, ReviewCoverage, ReviewEntries},
}

// artifactTypes is the order the hub lists the types in, for messages.
var artifactTypes = []string{ArtifactTypeControlCatalog, ArtifactTypeThreatCatalog, ArtifactTypeGuidanceCatalog}

// ValidArtifactType reports whether artifactType is one the hub drafts.
func ValidArtifactType(artifactType string) bool { _, ok := fields[artifactType]; return ok }

// ArtifactTypes lists the types the hub drafts, in a stable order, so an
// error message can name them without keeping its own copy of the table.
func ArtifactTypes() []string { return append([]string(nil), artifactTypes...) }

// Fields lists the polish fields artifactType has, in template order; nil
// for a type the hub does not draft.
func Fields(artifactType string) []string { return append([]string(nil), fields[artifactType]...) }

// Reviews lists the review kinds artifactType has; nil for an unknown type.
func Reviews(artifactType string) []string { return append([]string(nil), reviews[artifactType]...) }

// ValidField reports whether field is one the hub has a template for on
// artifactType.
func ValidField(artifactType, field string) bool { return has(fields[artifactType], field) }

func has(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// ValidMode reports whether mode is generate or polish.
func ValidMode(mode string) bool { return mode == ModeGenerate || mode == ModePolish }

// PolishRequest is the body of POST /v1/ai/polish. Context's shape depends
// on Field: DescriptionContext for FieldDescription, ObjectiveContext for
// FieldControlObjective, RequirementContext for FieldAssessmentRequirement,
// ThreatDescriptionContext for FieldThreatDescription,
// GuidelineObjectiveContext for FieldGuidelineObjective, StatementContext
// for FieldStatementText. Nothing in it is persisted or logged by the hub.
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

// Group is a catalog group as the description templates see it.
type Group struct {
	ID    string `json:"id"`
	Title string `json:"title,omitempty"`
}

// ThreatSummary is a threat as the threat catalog templates see it: in a
// DescriptionContext, a DuplicatesContext (where ID may be a client-side
// label such as "row 7"), and as a sibling in a ThreatDescriptionContext.
type ThreatSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Group       string `json:"group,omitempty"`
}

// GuidelineSummary is a guideline as the guidance catalog templates see it:
// in a DescriptionContext and an EntriesContext, as a sibling in a
// GuidelineObjectiveContext, and (embedded in GuidelineStatements) in a
// DuplicatesContext, where ID may be a client-side label such as "row 7".
type GuidelineSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Objective string `json:"objective,omitempty"`
	Group     string `json:"group,omitempty"`
}

// DescriptionContext is PolishRequest.Context for FieldDescription: the
// catalog title, its groups and its entries, Controls at objective
// altitude for a ControlCatalog, Threats for a ThreatCatalog, Guidelines
// at objective altitude for a GuidanceCatalog. Requirements, capability
// mappings and statements are deliberately not sent for this field.
type DescriptionContext struct {
	Title      string             `json:"title"`
	Groups     []Group            `json:"groups,omitempty"`
	Controls   []ControlSummary   `json:"controls,omitempty"`
	Threats    []ThreatSummary    `json:"threats,omitempty"`
	Guidelines []GuidelineSummary `json:"guidelines,omitempty"`
}

// CapabilityMapping is one of a threat's capability mappings as the
// threat description template sees it: the referenced catalog and the
// entry IDs in it.
type CapabilityMapping struct {
	Reference string   `json:"reference"`
	Entries   []string `json:"entries"`
	Remarks   string   `json:"remarks,omitempty"`
}

// ThreatDescriptionContext is PolishRequest.Context for
// FieldThreatDescription. The threat's capability mappings are the content
// source; Siblings, every other threat with a non-blank description, are
// the style pattern. The threat being written never appears in Siblings;
// its own text travels as PolishRequest.Current. Clients cap Siblings.
type ThreatDescriptionContext struct {
	Catalog CatalogSummary `json:"catalog"`
	Threat  struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Group string `json:"group,omitempty"`
	} `json:"threat"`
	Capabilities []CapabilityMapping `json:"capabilities,omitempty"`
	Siblings     []ThreatSummary     `json:"siblings,omitempty"`
}

// Requirement is one assessment requirement: the control's own in an
// ObjectiveContext, a sibling in a RequirementContext. The guidance catalog
// contexts carry a guideline's statements in the same shape (ID and Text;
// Applicability stays empty), so one list type serves both editors.
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

// RequirementContext is PolishRequest.Context for FieldAssessmentRequirement,
// the inverse of ObjectiveContext: the parent control's objective is the
// content source and Siblings, every other non-blank requirement (the same
// control's first, then the rest of the catalog's), are the style pattern.
// The row being written never appears in Siblings; its own text travels as
// PolishRequest.Current. ApplicabilityGroups is the catalog's authored
// applicability list so an ID like "tlp-amber" reads as a scope. Clients cap
// Siblings so a large catalog stays inside limits.MaxAIPolishRequestBytes.
type RequirementContext struct {
	Catalog struct {
		Title       string `json:"title"`
		Description string `json:"description,omitempty"`
	} `json:"catalog"`
	Control struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Objective string `json:"objective,omitempty"`
		Group     string `json:"group,omitempty"`
	} `json:"control"`
	Requirement struct {
		ID             string   `json:"id,omitempty"`
		Applicability  []string `json:"applicability,omitempty"`
		Recommendation string   `json:"recommendation,omitempty"`
	} `json:"requirement"`
	Siblings            []Requirement `json:"siblings,omitempty"`
	ApplicabilityGroups []Group       `json:"applicability_groups,omitempty"`
}

// GuidelineObjectiveContext is PolishRequest.Context for
// FieldGuidelineObjective, the GuidanceCatalog twin of ObjectiveContext.
// The guideline's own statements are the content source; Siblings, every
// other guideline with a non-blank objective, are the style pattern. The
// guideline being written never appears in Siblings.
type GuidelineObjectiveContext struct {
	Catalog   CatalogSummary `json:"catalog"`
	Guideline struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Group string `json:"group,omitempty"`
	} `json:"guideline"`
	Statements []Requirement      `json:"statements,omitempty"`
	Siblings   []GuidelineSummary `json:"siblings,omitempty"`
}

// StatementContext is PolishRequest.Context for FieldStatementText, the
// GuidanceCatalog twin of RequirementContext: the parent guideline's
// objective is the content source and Siblings, every other non-blank
// statement (the same guideline's first, then the rest of the catalog's),
// are the style pattern. The row being written never appears in Siblings;
// its own text travels as PolishRequest.Current. Clients cap Siblings.
type StatementContext struct {
	Catalog   CatalogSummary `json:"catalog"`
	Guideline struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Objective string `json:"objective,omitempty"`
		Group     string `json:"group,omitempty"`
	} `json:"guideline"`
	Statement struct {
		ID    string `json:"id,omitempty"`
		Title string `json:"title,omitempty"`
	} `json:"statement"`
	Siblings []Requirement `json:"siblings,omitempty"`
}

// Review kinds for POST /v1/ai/review (REV-415): whole-draft checks the
// polish call cannot express. Each answers structured JSON, not field
// text, and the hub owns one prompt template per kind.
const (
	ReviewDuplicates = "duplicates" // identical or functionally indistinguishable requirements anywhere in the catalog
	ReviewCoverage   = "coverage"   // the requirements one control's objective still lacks, plus rewrites of ones that fall short
	ReviewEntries    = "entries"    // the controls (or threats) the catalog's description and groups still lack
)

// ValidReview reports whether kind is a review the hub has a template for
// on artifactType.
func ValidReview(artifactType, kind string) bool { return has(reviews[artifactType], kind) }

// ReviewRequest is the body of POST /v1/ai/review. Context's shape depends
// on Kind: DuplicatesContext for ReviewDuplicates, CoverageContext for
// ReviewCoverage, EntriesContext for ReviewEntries. Nothing in it is persisted or logged by the hub. The
// body is capped by limits.MaxAIReviewRequestBytes.
type ReviewRequest struct {
	ArtifactType string          `json:"artifact_type"`
	Kind         string          `json:"kind"`
	Context      json.RawMessage `json:"context"`
}

// CatalogSummary is the catalog as the review templates see it.
type CatalogSummary struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// ControlRequirements is one control with its requirements, as the
// duplicates template sees the whole catalog. A requirement's ID is the
// authored one, or a client-side label such as "row 7" when none is typed
// yet, so the answer can name it.
type ControlRequirements struct {
	ControlSummary
	Requirements []Requirement `json:"requirements,omitempty"`
}

// GuidelineStatements is one guideline with its statements, as the
// duplicates template sees a guidance catalog. IDs are the authored ones or
// client-side labels, as in ControlRequirements.
type GuidelineStatements struct {
	GuidelineSummary
	Statements []Requirement `json:"statements,omitempty"`
}

// DuplicatesContext is ReviewRequest.Context for ReviewDuplicates: for a
// ControlCatalog every control with its requirements, for a ThreatCatalog
// every threat, for a GuidanceCatalog every guideline with its statements,
// in catalog order.
type DuplicatesContext struct {
	Catalog    CatalogSummary        `json:"catalog"`
	Controls   []ControlRequirements `json:"controls,omitempty"`
	Threats    []ThreatSummary       `json:"threats,omitempty"`
	Guidelines []GuidelineStatements `json:"guidelines,omitempty"`
}

// CoverageContext is ReviewRequest.Context for ReviewCoverage: one control
// whose objective is the content, its existing requirements (the ones the
// answer may add to or rewrite), and Siblings, other controls'
// requirements as the style pattern, capped by the client. A
// GuidanceCatalog sends Guideline and Statements instead of Control and
// Requirements, Siblings being other guidelines' statements; Control is
// then left empty and ApplicabilityGroups unused.
type CoverageContext struct {
	Catalog CatalogSummary `json:"catalog"`
	Control struct {
		ID        string `json:"id"`
		Title     string `json:"title"`
		Objective string `json:"objective"`
		Group     string `json:"group,omitempty"`
	} `json:"control"`
	Requirements        []Requirement     `json:"requirements,omitempty"`
	Guideline           *GuidelineSummary `json:"guideline,omitempty"`
	Statements          []Requirement     `json:"statements,omitempty"`
	Siblings            []Requirement     `json:"siblings,omitempty"`
	ApplicabilityGroups []Group           `json:"applicability_groups,omitempty"`
}

// EntriesContext is ReviewRequest.Context for ReviewEntries: the catalog's
// title and description (the content), its groups, and its entries so far,
// Controls for a ControlCatalog, Threats for a ThreatCatalog or Guidelines
// for a GuidanceCatalog. The answer proposes entries (Suggestion with
// Title, Text as the objective or description, and Group) the catalog
// still lacks.
type EntriesContext struct {
	Catalog    CatalogSummary     `json:"catalog"`
	Groups     []Group            `json:"groups,omitempty"`
	Controls   []ControlSummary   `json:"controls,omitempty"`
	Threats    []ThreatSummary    `json:"threats,omitempty"`
	Guidelines []GuidelineSummary `json:"guidelines,omitempty"`
}

// ReviewResponse is the body of POST /v1/ai/review. Duplicates is set for
// ReviewDuplicates (empty means none found), Suggestions for
// ReviewCoverage and ReviewEntries (empty means nothing is missing). Note is a short
// status for the author, as on PolishResponse.
type ReviewResponse struct {
	Duplicates  []DuplicateGroup `json:"duplicates,omitempty"`
	Suggestions []Suggestion     `json:"suggestions,omitempty"`
	Note        string           `json:"note,omitempty"`
}

// DuplicateGroup is one set of requirements the review judged identical or
// functionally indistinguishable, by the IDs (or labels) the context used.
type DuplicateGroup struct {
	IDs    []string `json:"ids"`
	Reason string   `json:"reason"`
}

// Suggestion actions: add a new requirement under the control, or rewrite
// an existing one (ID names it) so the objective is covered.
const (
	SuggestionAdd     = "add"
	SuggestionRewrite = "rewrite"
)

// Suggestion is one proposed change from a coverage or entries review. The
// client shows it for acceptance; nothing is written until the author
// agrees. A coverage suggestion is a requirement (Text, Applicability) or a
// statement (Text); an entries suggestion is a control, threat or guideline
// (Title, Text as its objective or description, Group), always an add.
type Suggestion struct {
	Action string `json:"action"`
	// ID: on add, a proposed ID in the siblings' pattern (may be empty);
	// on rewrite, the existing requirement's ID or label from the context.
	ID            string   `json:"id,omitempty"`
	Title         string   `json:"title,omitempty"`
	Text          string   `json:"text"`
	Group         string   `json:"group,omitempty"`
	Applicability []string `json:"applicability,omitempty"`
	Reason        string   `json:"reason,omitempty"`
}
