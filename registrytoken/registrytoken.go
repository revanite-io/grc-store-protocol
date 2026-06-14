// SPDX-License-Identifier: Apache-2.0

// Package registrytoken is the GET /v2/token response shape (ADR-0031). The
// token-minting flow and JWT access-claim parsing stay in consumers — only the
// response type is contract.
package registrytoken

// Response is the hub's Distribution bearer-token response. The Distribution
// spec splits the field — some responses carry token, some access_token — so use
// BearerToken to read it rather than picking a field by hand.
type Response struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
	IssuedAt    string `json:"issued_at,omitempty"`
}

// BearerToken returns the registry bearer token, preferring Token and falling
// back to AccessToken. Empty only if the hub returned neither.
func (r Response) BearerToken() string {
	if r.Token != "" {
		return r.Token
	}
	return r.AccessToken
}
