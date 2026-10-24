// SPDX-License-Identifier: Apache-2.0

package targetgroup

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// The JSON keys are the contract the hub and the frontend share.
func TestViewJSONKeys(t *testing.T) {
	at := time.Date(2026, 10, 24, 1, 2, 3, 0, time.UTC)
	ns, id, ver := "acme", "osps", "2025.10"
	v := View{
		Group: Group{Namespace: "acme", Slug: "privateer", Name: "Privateer", Members: []Member{{TargetID: "sdk"}}, CreatedAt: at, UpdatedAt: at},
		Members: []MemberView{{TargetID: "sdk", Name: "SDK", URI: "https://github.com/acme/sdk", Evaluations: []EvaluationView{{
			CatalogRefID: "OSPS-B", CatalogNamespace: &ns, CatalogID: &id, PluginDeclaredCatalogVersion: &ver,
			EvaluatorPlugin: EvaluatorPlugin{Namespace: "acme", PluginID: "pvtr", Version: "1.0.0"},
			TargetVersion:   "1.2.0", RunAt: at, LogNamespace: "acme", LogID: "results", LogVersion: "1.2.0-20261024T010203Z",
			LogManifestDigest: "sha256:ab", Result: "Passed", Counts: Counts{Passed: 3},
		}}}},
	}
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		`"group":`, `"slug":"privateer"`, `"members":[{"target_id":"sdk"}]`, `"verified_at":null`,
		`"plugin_declared_catalog_version":"2025.10"`, `"evaluator_plugin":{"namespace":"acme","plugin_id":"pvtr","version":"1.0.0"}`,
		`"log_manifest_digest":"sha256:ab"`, `"counts":{"passed":3,`, `"evaluated_requirements":0`,
	} {
		if !strings.Contains(string(out), key) {
			t.Errorf("missing %s in %s", key, out)
		}
	}
	for _, absent := range []string{`"archived_at"`, `"description"`} {
		if strings.Contains(string(out), absent) {
			t.Errorf("%s must be omitted when empty: %s", absent, out)
		}
	}
	var back View
	if err := json.Unmarshal(out, &back); err != nil || !back.Members[0].Evaluations[0].RunAt.Equal(at) {
		t.Errorf("round trip lost data: %v", err)
	}
}

func TestUpsertRequestDecode(t *testing.T) {
	var req UpsertRequest
	if err := json.Unmarshal([]byte(`{"name":"P","members":["a","b"]}`), &req); err != nil ||
		req.Name != "P" || len(req.Members) != 2 || req.Members[1] != "b" {
		t.Errorf("decode: %+v (%v)", req, err)
	}
}
