// SPDX-License-Identifier: Apache-2.0

package assist

import (
	"encoding/json"
	"testing"
)

func TestValidators(t *testing.T) {
	if !ValidProvider(ProviderAnthropic) || ValidProvider("openai") || ValidProvider("") {
		t.Error("ValidProvider: only anthropic is accepted today")
	}
	if !ValidField(FieldDescription) || !ValidField(FieldControlObjective) || !ValidField(FieldAssessmentRequirement) || ValidField("title") {
		t.Error("ValidField: description, control.objective and assessment-requirement.text only")
	}
	if !ValidMode(ModeGenerate) || !ValidMode(ModePolish) || ValidMode("rewrite") {
		t.Error("ValidMode: generate and polish only")
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
	if !ValidReview(ReviewDuplicates) || !ValidReview(ReviewCoverage) || ValidReview("polish") || ValidReview("") {
		t.Error("ValidReview: duplicates and coverage only")
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
