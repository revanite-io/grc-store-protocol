// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"encoding/json"
	"testing"
)

// Each policy field lands on its own key, and an older hub's document, which
// has none of them, decodes to nil rather than false.
func TestPolicyFields(t *testing.T) {
	f, tr := false, true
	b, err := json.Marshal(Document{PrivateOnly: &tr, NamespacePrefixRequired: &f, AnonymousReads: &tr, NamespaceSelfService: &f, AccessRequests: &tr})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	want := map[string]bool{"private_only": true, "namespace_prefix_required": false, "anonymous_reads": true, "namespace_self_service": false, "access_requests": true}
	for k, w := range want {
		if v, ok := m[k]; !ok || v != w {
			t.Errorf("%s = %v, %v; want %v, present", k, v, ok, w)
		}
	}

	var old Document
	if err := json.Unmarshal([]byte(`{"hub_url":"https://hub.example"}`), &old); err != nil {
		t.Fatal(err)
	}
	if old.PrivateOnly != nil || old.NamespacePrefixRequired != nil || old.AnonymousReads != nil || old.NamespaceSelfService != nil || old.AccessRequests != nil {
		t.Errorf("older hub's document decoded policy fields as set: %+v", old)
	}
}
