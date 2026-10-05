// SPDX-License-Identifier: Apache-2.0

package assist

import (
	"github.com/revanite-io/grc-store-protocol/draft"

	"encoding/json"
	"testing"
)

func TestValidators(t *testing.T) {
	if !ValidProvider(ProviderAnthropic) || ValidProvider("openai") || ValidProvider("") {
		t.Error("ValidProvider: only anthropic is accepted today")
	}
	if !ValidArtifactType(ArtifactTypeControlCatalog) || !ValidArtifactType(ArtifactTypeThreatCatalog) || !ValidArtifactType(ArtifactTypeGuidanceCatalog) || ValidArtifactType("Policy") || ValidArtifactType("") {
		t.Error("ValidArtifactType: ControlCatalog, ThreatCatalog and GuidanceCatalog only")
	}
	gc := ArtifactTypeGuidanceCatalog
	if !ValidField(gc, FieldDescription) || !ValidField(gc, FieldGuidelineObjective) || !ValidField(gc, FieldStatementText) || !ValidField(gc, FieldStatementTitle) || ValidField(gc, FieldControlObjective) || ValidField(ArtifactTypeControlCatalog, FieldStatementText) {
		t.Error("ValidField(GuidanceCatalog): description, guideline.objective, statement.text and statement.title only")
	}
	cc, tc := ArtifactTypeControlCatalog, ArtifactTypeThreatCatalog
	if !ValidField(cc, FieldDescription) || !ValidField(cc, FieldControlObjective) || !ValidField(cc, FieldAssessmentRequirement) || ValidField(cc, "title") || ValidField(cc, FieldThreatDescription) {
		t.Error("ValidField(ControlCatalog): description, control.objective and assessment-requirement.text only")
	}
	if !ValidField(tc, FieldDescription) || !ValidField(tc, FieldThreatDescription) || ValidField(tc, FieldControlObjective) || ValidField("Policy", FieldDescription) {
		t.Error("ValidField(ThreatCatalog): description and threat.description only")
	}
	if !ValidMode(ModeGenerate) || !ValidMode(ModePolish) || ValidMode("rewrite") {
		t.Error("ValidMode: generate and polish only")
	}
	// The listings agree with the validators, so a message built from them
	// cannot drift; they hand out copies.
	types := ArtifactTypes()
	if len(types) != 3 || types[0] != ArtifactTypeControlCatalog || types[2] != ArtifactTypeGuidanceCatalog {
		t.Errorf("ArtifactTypes = %v", types)
	}
	for _, at := range types {
		if !ValidArtifactType(at) {
			t.Errorf("ArtifactTypes lists %s but ValidArtifactType refuses it", at)
		}
		for _, f := range Fields(at) {
			if !ValidField(at, f) {
				t.Errorf("Fields(%s) lists %s but ValidField refuses it", at, f)
			}
		}
		for _, k := range Reviews(at) {
			if !ValidReview(at, k) {
				t.Errorf("Reviews(%s) lists %s but ValidReview refuses it", at, k)
			}
		}
	}
	if Fields("Policy") != nil || Reviews("") != nil {
		t.Error("Fields/Reviews of an unknown type must be nil")
	}
	types[0] = "x"
	if ArtifactTypes()[0] != ArtifactTypeControlCatalog {
		t.Error("ArtifactTypes must return a copy")
	}
}

// The context is opaque on the envelope and typed per field: a round trip
// through PolishRequest must hand the field's own struct back unchanged.
func TestContextRoundTrip(t *testing.T) {
	var octx ObjectiveContext
	octx.Catalog.Title = "ACME Cloud"
	octx.Control.ID = "CN01"
	octx.Requirements = []Requirement{{ID: "CN01.AR01", Text: "MUST", Applicability: []string{"tlp-green"}}}
	octx.Siblings = []ControlSummary{{ID: "CN02", Title: "B", Objective: "Ensure B."}}
	raw, err := json.Marshal(octx)
	if err != nil {
		t.Fatal(err)
	}
	env, err := json.Marshal(PolishRequest{ArtifactType: ArtifactTypeControlCatalog, Field: FieldControlObjective, Mode: ModeGenerate, Context: raw})
	if err != nil {
		t.Fatal(err)
	}
	var back PolishRequest
	if err := json.Unmarshal(env, &back); err != nil {
		t.Fatal(err)
	}
	var got ObjectiveContext
	if err := json.Unmarshal(back.Context, &got); err != nil {
		t.Fatal(err)
	}
	if got.Control.ID != "CN01" || len(got.Requirements) != 1 || got.Siblings[0].Objective != "Ensure B." {
		t.Errorf("context did not survive the envelope: %+v", got)
	}
	if back.Current != "" {
		t.Errorf("current = %q, want empty on generate", back.Current)
	}

	var rctx RequirementContext
	rctx.Control.Objective = "Ensure keys rotate."
	rctx.Requirement.Applicability = []string{"tlp-amber"}
	rctx.Siblings = []Requirement{{ID: "CN01.AR01", Text: "MUST"}}
	rctx.ApplicabilityGroups = []Group{{ID: "tlp-amber", Title: "Amber"}}
	raw, _ = json.Marshal(rctx)
	var rgot RequirementContext
	if err := json.Unmarshal(raw, &rgot); err != nil {
		t.Fatal(err)
	}
	if rgot.Control.Objective != "Ensure keys rotate." || rgot.Requirement.Applicability[0] != "tlp-amber" || rgot.Siblings[0].ID != "CN01.AR01" || rgot.ApplicabilityGroups[0].Title != "Amber" {
		t.Errorf("requirement context did not round-trip: %+v", rgot)
	}
}

func TestValidReview(t *testing.T) {
	if !ValidReview(ArtifactTypeControlCatalog, ReviewDuplicates) || !ValidReview(ArtifactTypeControlCatalog, ReviewCoverage) || ValidReview(ArtifactTypeControlCatalog, "polish") || ValidReview(ArtifactTypeControlCatalog, "") {
		t.Error("ValidReview(ControlCatalog): duplicates and coverage only")
	}
	if !ValidReview(ArtifactTypeThreatCatalog, ReviewDuplicates) || !ValidReview(ArtifactTypeThreatCatalog, ReviewEntries) || ValidReview(ArtifactTypeThreatCatalog, ReviewCoverage) || ValidReview("Policy", ReviewDuplicates) {
		t.Error("ValidReview(ThreatCatalog): duplicates and entries only")
	}
	if !ValidReview(ArtifactTypeControlCatalog, ReviewEntries) {
		t.Error("ValidReview(ControlCatalog): entries")
	}
	if !ValidReview(ArtifactTypeGuidanceCatalog, ReviewDuplicates) || !ValidReview(ArtifactTypeGuidanceCatalog, ReviewCoverage) || !ValidReview(ArtifactTypeGuidanceCatalog, ReviewEntries) || !ValidReview(ArtifactTypeGuidanceCatalog, ReviewRecommendations) || ValidReview(ArtifactTypeGuidanceCatalog, "polish") {
		t.Error("ValidReview(GuidanceCatalog): duplicates, coverage, entries and recommendations")
	}
	// Only a guidance statement nests recommendations.
	if ValidReview(ArtifactTypeControlCatalog, ReviewRecommendations) || ValidReview(ArtifactTypeThreatCatalog, ReviewRecommendations) {
		t.Error("ValidReview: recommendations is GuidanceCatalog only")
	}
	entry, _ := json.Marshal(Suggestion{Action: SuggestionAdd, ID: "CN03", Title: "Rotate", Text: "Ensure keys rotate.", Group: "Encryption"})
	if string(entry) != `{"action":"add","id":"CN03","title":"Rotate","text":"Ensure keys rotate.","group":"Encryption"}` {
		t.Errorf("entries Suggestion JSON = %s", entry)
	}
}

// A coverage context round-trips through the opaque envelope, and the
// embedded ControlSummary flattens into the control's own keys.
func TestReviewRoundTrip(t *testing.T) {
	var cctx CoverageContext
	cctx.Catalog.Title = "ACME"
	cctx.Control.ID, cctx.Control.Objective = "CN01", "Ensure encryption."
	cctx.Requirements = []Requirement{{ID: "row 3", Text: "MUST encrypt"}}
	inner, err := json.Marshal(cctx)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(ReviewRequest{ArtifactType: ArtifactTypeControlCatalog, Kind: ReviewCoverage, Context: inner})
	if err != nil {
		t.Fatal(err)
	}
	var req ReviewRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatal(err)
	}
	var back CoverageContext
	if err := json.Unmarshal(req.Context, &back); err != nil {
		t.Fatal(err)
	}
	if back.Control.Objective != "Ensure encryption." || back.Requirements[0].ID != "row 3" {
		t.Errorf("round trip lost fields: %+v", back)
	}
	flat, _ := json.Marshal(ControlRequirements{ControlSummary: ControlSummary{ID: "CN01", Title: "Encrypt"}, Requirements: []Requirement{{ID: "a", Text: "b"}}})
	if string(flat) != `{"id":"CN01","title":"Encrypt","requirements":[{"id":"a","text":"b"}]}` {
		t.Errorf("ControlRequirements JSON = %s", flat)
	}
}

// A threat description context round-trips, and the shared shapes carry
// the threat catalog's lists beside the control catalog's.
func TestThreatContexts(t *testing.T) {
	var tctx ThreatDescriptionContext
	tctx.Catalog.Title = "Threats"
	tctx.Threat.ID, tctx.Threat.Group = "TH01", "Data"
	tctx.Capabilities = []CapabilityMapping{{Reference: "CP", Entries: []string{"CP08", "CP09"}, Remarks: "rules"}}
	tctx.Siblings = []ThreatSummary{{ID: "TH02", Title: "B", Description: "d"}}
	raw, err := json.Marshal(PolishRequest{ArtifactType: ArtifactTypeThreatCatalog, Field: FieldThreatDescription, Mode: ModeGenerate, Context: mustRaw(t, tctx)})
	if err != nil {
		t.Fatal(err)
	}
	var back PolishRequest
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	var got ThreatDescriptionContext
	if err := json.Unmarshal(back.Context, &got); err != nil {
		t.Fatal(err)
	}
	if got.Threat.Group != "Data" || got.Capabilities[0].Entries[1] != "CP09" || got.Siblings[0].Description != "d" {
		t.Errorf("threat context did not survive the envelope: %+v", got)
	}
	dup, _ := json.Marshal(DuplicatesContext{Catalog: CatalogSummary{Title: "T"}, Threats: []ThreatSummary{{ID: "row 2", Title: "x"}}})
	if string(dup) != `{"catalog":{"title":"T"},"threats":[{"id":"row 2","title":"x"}]}` {
		t.Errorf("threat DuplicatesContext JSON = %s", dup)
	}
}

// The guidance catalog's own polish contexts round-trip, and the shared
// review shapes carry its lists beside the other two types'.
func TestGuidanceContexts(t *testing.T) {
	var octx GuidelineObjectiveContext
	octx.Catalog.Title = "Guidance"
	octx.Guideline.ID, octx.Guideline.Group = "GL01", "Setup"
	octx.Statements = []Requirement{{ID: "GL01.S1", Text: "Organizations SHOULD…"}}
	octx.Siblings = []GuidelineSummary{{ID: "GL02", Title: "B", Objective: "Ensure B."}}
	raw, err := json.Marshal(PolishRequest{ArtifactType: ArtifactTypeGuidanceCatalog, Field: FieldGuidelineObjective, Mode: ModeGenerate, Context: mustRaw(t, octx)})
	if err != nil {
		t.Fatal(err)
	}
	var back PolishRequest
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	var got GuidelineObjectiveContext
	if err := json.Unmarshal(back.Context, &got); err != nil {
		t.Fatal(err)
	}
	if got.Guideline.Group != "Setup" || got.Statements[0].ID != "GL01.S1" || got.Siblings[0].Objective != "Ensure B." {
		t.Errorf("guideline objective context did not survive the envelope: %+v", got)
	}
	var sctx StatementContext
	sctx.Guideline.ID, sctx.Guideline.Objective = "GL01", "Ensure setup."
	sctx.Statement.ID = "row 3"
	if s, _ := json.Marshal(sctx); string(s) != `{"catalog":{"title":""},"guideline":{"id":"GL01","title":"","objective":"Ensure setup."},"statement":{"id":"row 3"}}` {
		t.Errorf("StatementContext JSON = %s", s)
	}
	// For the title field the statement's text is content and siblings carry titles.
	sctx.Statement.Text = "Organizations SHOULD assess maintainers."
	sctx.Siblings = []Requirement{{ID: "GL01.S1", Title: "Reject Untrustworthy Maintainers", Text: "…"}}
	if s, _ := json.Marshal(sctx); string(s) != `{"catalog":{"title":""},"guideline":{"id":"GL01","title":"","objective":"Ensure setup."},"statement":{"id":"row 3","text":"Organizations SHOULD assess maintainers."},"siblings":[{"id":"GL01.S1","title":"Reject Untrustworthy Maintainers","text":"…"}]}` {
		t.Errorf("StatementContext title JSON = %s", s)
	}
	dup, _ := json.Marshal(DuplicatesContext{Catalog: CatalogSummary{Title: "G"}, Guidelines: []GuidelineStatements{{GuidelineSummary: GuidelineSummary{ID: "GL01", Title: "A"}, Statements: []Requirement{{ID: "row 2", Text: "x"}}}}})
	if string(dup) != `{"catalog":{"title":"G"},"guidelines":[{"id":"GL01","title":"A","statements":[{"id":"row 2","text":"x"}]}]}` {
		t.Errorf("guidance DuplicatesContext JSON = %s", dup)
	}
	cov, _ := json.Marshal(CoverageContext{Catalog: CatalogSummary{Title: "G"}, Guideline: &GuidelineSummary{ID: "GL01", Title: "A", Objective: "o"}, Statements: []Requirement{{ID: "a", Text: "b"}}})
	if string(cov) != `{"catalog":{"title":"G"},"control":{"id":"","title":"","objective":""},"guideline":{"id":"GL01","title":"A","objective":"o"},"statements":[{"id":"a","text":"b"}]}` {
		t.Errorf("guidance CoverageContext JSON = %s", cov)
	}
}

// A recommendations context round-trips through the review envelope with
// the statement, its existing recommendations and the style siblings.
func TestRecommendationsContext(t *testing.T) {
	ctx := RecommendationsContext{
		Catalog:         CatalogSummary{Title: "G"},
		Guideline:       GuidelineSummary{ID: "GL01", Title: "A", Objective: "o"},
		Statement:       Requirement{ID: "GL01.S1", Title: "Review Source", Text: "Confirm the source can be retrieved."},
		Recommendations: []string{"Check the release artifacts."},
		Siblings:        []StatementRecommendations{{Statement: Requirement{ID: "GL01.S2", Text: "x"}, Recommendations: []string{"y"}}},
	}
	raw, err := json.Marshal(ReviewRequest{ArtifactType: ArtifactTypeGuidanceCatalog, Kind: ReviewRecommendations, Context: mustRaw(t, ctx)})
	if err != nil {
		t.Fatal(err)
	}
	var back ReviewRequest
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatal(err)
	}
	if back.Kind != "recommendations" {
		t.Errorf("Kind = %q", back.Kind)
	}
	got, _ := json.Marshal(json.RawMessage(back.Context))
	want := `{"catalog":{"title":"G"},"guideline":{"id":"GL01","title":"A","objective":"o"},"statement":{"id":"GL01.S1","title":"Review Source","text":"Confirm the source can be retrieved."},"recommendations":["Check the release artifacts."],"siblings":[{"statement":{"id":"GL01.S2","text":"x"},"recommendations":["y"]}]}`
	if string(got) != want {
		t.Errorf("RecommendationsContext JSON = %s", got)
	}
	// "None" is an empty answer with a reason in Note.
	none, _ := json.Marshal(ReviewResponse{Note: "The statement is already a single check."})
	if string(none) != `{"note":"The statement is already a single check."}` {
		t.Errorf("empty recommendations ReviewResponse JSON = %s", none)
	}
}

func mustRaw(t *testing.T, v any) json.RawMessage {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

// Every assist type is draftable, the aliased constants are the same
// strings, and a draftable type without AI assist (CapabilityCatalog) is
// left out of the assist list, so the editors, the draft store and the AI
// routes never disagree on a type name.
func TestArtifactTypesCoverDraftable(t *testing.T) {
	draftable := map[string]bool{}
	for _, tt := range draft.ArtifactTypes() {
		draftable[tt] = true
	}
	for _, tt := range ArtifactTypes() {
		if !draftable[tt] {
			t.Errorf("assist type %q is not draftable", tt)
		}
	}
	if ArtifactTypeControlCatalog != draft.ArtifactTypeControlCatalog {
		t.Error("assist and draft constants differ")
	}
	if ValidArtifactType(draft.ArtifactTypeCapabilityCatalog) {
		t.Error("CapabilityCatalog has no assist table yet; it must not be valid here")
	}
	for _, tt := range ArtifactTypes() {
		if tt == draft.ArtifactTypeCapabilityCatalog {
			t.Error("assist lists CapabilityCatalog without a template table")
		}
	}
}
