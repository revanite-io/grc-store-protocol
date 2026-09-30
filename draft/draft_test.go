// SPDX-License-Identifier: Apache-2.0

package draft

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// The assist package's tests check that its AI tables cover every type here.
func TestArtifactTypes(t *testing.T) {
	for _, tt := range ArtifactTypes() {
		if !ValidArtifactType(tt) {
			t.Errorf("%q listed but not valid", tt)
		}
	}
	if ValidArtifactType("EvaluationLog") || ValidArtifactType("") {
		t.Error("non-editor types must be invalid")
	}
}

func TestValidBody(t *testing.T) {
	cases := map[string]bool{
		"{}": true, " {\"a\":1} ": true, "[]": false, "": false, "null": false,
		"{\"a\":": false, "\"s\"": false,
	}
	for in, want := range cases {
		if got := ValidBody(json.RawMessage(in)); got != want {
			t.Errorf("ValidBody(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestDraftJSON(t *testing.T) {
	at := time.Date(2026, 9, 30, 1, 2, 3, 456789000, time.UTC)
	d := Draft{ID: "x", Namespace: "acme", ArtifactType: ArtifactTypeControlCatalog, CreatedAt: at, UpdatedAt: at}
	out, _ := json.Marshal(d)
	if strings.Contains(string(out), `"body"`) {
		t.Errorf("empty body must be omitted (lists): %s", out)
	}
	d.Body = json.RawMessage(`{"meta":{"title":"T"}}`)
	out, _ = json.Marshal(d)
	var back Draft
	if err := json.Unmarshal(out, &back); err != nil || string(back.Body) != string(d.Body) || !back.UpdatedAt.Equal(at) {
		t.Errorf("round trip lost data: %s (%v)", out, err)
	}
	var req SaveRequest
	if err := json.Unmarshal([]byte(`{"title":"T","body":{},"expected_updated_at":"2026-09-30T01:02:03.456789Z"}`), &req); err != nil || req.ExpectedUpdatedAt == nil || !req.ExpectedUpdatedAt.Equal(at) {
		t.Errorf("SaveRequest decode: %+v (%v)", req, err)
	}
}
