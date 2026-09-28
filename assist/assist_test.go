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
	if !ValidField(FieldDescription) || !ValidField(FieldControlObjective) || ValidField("title") {
		t.Error("ValidField: description and control.objective only")
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
}
