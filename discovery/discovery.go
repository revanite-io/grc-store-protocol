// SPDX-License-Identifier: Apache-2.0

// Package discovery is the GET /.well-known/grc-store-configuration document — the
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

	// The hub's namespace and account policy, so a client can leave out an
	// action the hub would refuse. Always present on a hub that serves them:
	// a client that finds one absent is talking to an older hub, and should
	// offer the action and handle the refusal.

	// PrivateOnly is true when the hub refuses public namespaces
	// (apierror private_only_hub).
	PrivateOnly bool `json:"private_only"`
	// NamespacePrefixRequired is true when a new namespace's slug must start
	// with its enterprise's slug.
	NamespacePrefixRequired bool `json:"namespace_prefix_required"`
	// AnonymousReads is true when a caller may read without signing in.
	AnonymousReads bool `json:"anonymous_reads"`
	// NamespaceSelfService is true when a signed-in user may create a
	// namespace without an operator.
	NamespaceSelfService bool `json:"namespace_self_service"`
	// AccessRequests is true when a person without an account may ask for one.
	AccessRequests bool `json:"access_requests"`
}
