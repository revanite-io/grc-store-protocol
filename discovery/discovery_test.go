// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"encoding/json"
	"testing"
)

// A false policy field must stay on the wire, so clients can tell it from an
// older hub that does not serve the field.
func TestPolicyFieldsAlwaysPresent(t *testing.T) {
	b, err := json.Marshal(Document{})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"private_only", "namespace_prefix_required", "anonymous_reads", "namespace_self_service", "access_requests"} {
		if v, ok := m[k]; !ok || v != false {
			t.Errorf("%s = %v, %v; want false, present", k, v, ok)
		}
	}
}
