// SPDX-License-Identifier: Apache-2.0

// Package discovery is the GET /.well-known/ext.grc-store document — the
// coordinates a client reads to talk to a hub (registry, OIDC, CI audience)
// instead of hard-coding them.
package discovery

// Document is the discovery response. Clients should prefer these advertised
// values over deriving hosts from hub_url (ADR-0026/0028/0032).
type Document struct {
	RegistryURL     string `json:"registry_url"`
	HubURL          string `json:"hub_url"`
	UIURL           string `json:"ui_url"`
	APIVersion      string `json:"api_version"`
	OIDCIssuer      string `json:"oidc_issuer,omitempty"`
	OIDCCLIClientID string `json:"oidc_cli_client_id,omitempty"`
	// CIAudience is the audience a publisher must request on its GitHub Actions
	// OIDC token for trusted publishing; omitted when CI publishing is
	// unconfigured.
	CIAudience string `json:"ci_audience,omitempty"`
}
