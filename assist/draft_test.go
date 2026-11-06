// SPDX-License-Identifier: Apache-2.0

package assist

import (
	"encoding/json"
	"testing"

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
			t.Errorf("%s: upstream %q, want %q", at, DraftUpstream(at), up)
		}
		if up != "" && !ValidDraftType(up) {
			t.Errorf("%s: upstream %s is not itself draftable", at, up)
		}
	}
	for _, bad := range []string{draft.ArtifactTypeEvaluationLog, "", "controlcatalog"} {
		if ValidDraftType(bad) || DraftUpstream(bad) != "" {
			t.Errorf("ValidDraftType(%q) = true", bad)
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
	b, _ = json.Marshal(DraftJob{
		State: DraftStatePartial, Attempt: 3, Result: json.RawMessage(`{}`),
		Removed: []DraftRemoved{{Path: "guidelines.3", ID: "G-4", Errors: []string{"e"}}},
		Usage:   &DraftUsage{InputTokens: 1, OutputTokens: 2},
	})
	if got := string(b); got != `{"state":"partial","attempt":3,"result":{},"removed":[{"path":"guidelines.3","id":"G-4","errors":["e"]}],"usage":{"input_tokens":1,"output_tokens":2}}` {
		t.Errorf("partial job = %s", got)
	}
}
