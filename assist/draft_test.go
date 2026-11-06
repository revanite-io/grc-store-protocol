// SPDX-License-Identifier: Apache-2.0

package assist

import (
	"encoding/json"
	"testing"

	"github.com/revanite-io/grc-store-protocol/apierror"
	"github.com/revanite-io/grc-store-protocol/draft"
)

func TestDraftUpstreamTable(t *testing.T) {
	want := map[string]string{
		"GuidanceCatalog":   "",
		"CapabilityCatalog": "GuidanceCatalog",
		"ThreatCatalog":     "CapabilityCatalog",
		"RiskCatalog":       "",
		"ControlCatalog":    "ThreatCatalog",
		"Policy":            "RiskCatalog",
	}
	types := DraftTypes()
	if len(types) != len(want) {
		t.Fatalf("DraftTypes = %v, want the six of hub ADR-0110", types)
	}
	for _, at := range types {
		up, ok := want[at]
		if !ok || !ValidDraftType(at) || DraftUpstream(at) != up {
			t.Errorf("%s: in table=%v valid=%v upstream=%q, want %q", at, ok, ValidDraftType(at), DraftUpstream(at), up)
		}
		if up != "" && !ValidDraftType(up) {
			t.Errorf("%s: upstream %s is not itself draftable", at, up)
		}
	}
	for _, bad := range []string{draft.ArtifactTypeEvaluationLog, "", "controlcatalog"} {
		if ValidDraftType(bad) || DraftUpstream(bad) != "" {
			t.Errorf("%q: valid=%v upstream=%q, want false and \"\"", bad, ValidDraftType(bad), DraftUpstream(bad))
		}
	}
	types[0] = "x"
	if DraftTypes()[0] == "x" {
		t.Error("DraftTypes must return a copy")
	}
}

func TestDraftJobWire(t *testing.T) {
	b, err := json.Marshal(DraftJob{State: DraftStateRunning, Stage: DraftStageGenerating})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"state":"running","stage":"generating","attempt":0}` {
		t.Errorf("running job = %s", got)
	}
	b, err = json.Marshal(DraftJob{
		State: DraftStatePartial, Attempt: 3, Result: json.RawMessage(`{}`),
		Removed: []DraftRemoved{{Path: "guidelines.3", ID: "G-4", Errors: []string{"e"}}},
		Usage:   &DraftUsage{InputTokens: 1, OutputTokens: 2},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"state":"partial","attempt":3,"result":{},"removed":[{"path":"guidelines.3","id":"G-4","errors":["e"]}],"usage":{"input_tokens":1,"output_tokens":2}}` {
		t.Errorf("partial job = %s", got)
	}
	b, err = json.Marshal(DraftJob{State: DraftStateFailed, Attempt: 1, Error: &DraftError{Code: apierror.AIDraftTimeout, Detail: "d"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"state":"failed","attempt":1,"error":{"code":"ai_draft_timeout","detail":"d"}}` {
		t.Errorf("failed job = %s", got)
	}
}

func TestDraftRequestAndListWire(t *testing.T) {
	b, err := json.Marshal(DraftRequest{
		Namespace: "acme", ArtifactType: draft.ArtifactTypePolicy,
		Target:     DraftTarget{ID: "p", Version: "1.0.0"},
		Answers:    []DraftAnswer{{Question: "q", Answer: "a"}},
		Upstream:   json.RawMessage(`{}`),
		References: []DraftReference{{ArtifactType: draft.ArtifactTypeControlCatalog, Namespace: "acme", ID: "c", Version: "1", Title: "C"}},
		Previous:   json.RawMessage(`{}`),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"namespace":"acme","artifact_type":"Policy","target":{"id":"p","version":"1.0.0"},"answers":[{"question":"q","answer":"a"}],"upstream":{},"references":[{"artifact_type":"ControlCatalog","namespace":"acme","id":"c","version":"1","title":"C"}],"previous":{}}` {
		t.Errorf("full request = %s", got)
	}
	b, err = json.Marshal(DraftRequest{Namespace: "acme", ArtifactType: draft.ArtifactTypeGuidanceCatalog, Answers: []DraftAnswer{}})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"namespace":"acme","artifact_type":"GuidanceCatalog","target":{"id":"","version":""},"answers":[]}` {
		t.Errorf("answers-only request = %s", got)
	}
	b, err = json.Marshal([]any{DraftStarted{JobID: "j"}, DraftListItem{JobID: "j", Namespace: "acme", ArtifactType: "RiskCatalog", State: DraftStateFailed, CreatedAt: "2026-11-06T00:00:00Z"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `[{"job_id":"j"},{"job_id":"j","namespace":"acme","artifact_type":"RiskCatalog","state":"failed","created_at":"2026-11-06T00:00:00Z"}]` {
		t.Errorf("started and list item = %s", got)
	}
}
