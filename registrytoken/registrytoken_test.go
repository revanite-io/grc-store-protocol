// SPDX-License-Identifier: Apache-2.0

package registrytoken

import "testing"

func TestBearerToken(t *testing.T) {
	cases := []struct {
		name string
		r    Response
		want string
	}{
		{"token preferred", Response{Token: "tk", AccessToken: "ac"}, "tk"},
		{"access_token fallback", Response{AccessToken: "ac"}, "ac"},
		{"token only", Response{Token: "tk"}, "tk"},
		{"neither", Response{}, ""},
	}
	for _, c := range cases {
		if got := c.r.BearerToken(); got != c.want {
			t.Errorf("%s: BearerToken() = %q, want %q", c.name, got, c.want)
		}
	}
}
