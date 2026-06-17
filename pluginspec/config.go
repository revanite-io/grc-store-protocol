// SPDX-License-Identifier: Apache-2.0

// Package pluginspec is the plugin config-blob schema (mediatype.PluginConfig) —
// the signed descriptor a bare binary can't self-carry (ADR-0034 decision 2).
// It is the producer↔hub shared type: the publisher (pvtr) writes and signs it;
// the hub reads it as the authoritative source on sync. (Named pluginspec, not
// plugin, so it does not shadow the standard library's plugin package.)
package pluginspec

// Config is one plugin config blob, one per child image manifest. Platform is
// per-child; Plugin, Version, Entrypoint, Protocol, and Evaluates describe the
// plugin as a whole and must agree across every child.
type Config struct {
	// Plugin is the <namespace>/<plugin_id> REGISTRY COORDINATE the index is
	// published under — NOT a GitHub owner/repo. A mismatch with the coordinate
	// is rejected (apierror.PluginCoordinateMismatch).
	Plugin string `json:"plugin"`
	// Version must equal the index tag (else apierror.TagVersionMismatch).
	Version string `json:"version"`
	// License is the publication license as an SPDX expression (ADR-0037),
	// required and enforced by the hub (else apierror.LicenseRequired). It lives
	// here on the SIGNED config — NOT in an OCI annotation — because ADR-0034
	// decision 6 makes the signed config the sole source of truth for plugin
	// metadata (catalog bundles carry license in the annotation instead; only the
	// plugin path is signed-config-bound). Like Plugin/Version/etc. it describes
	// the plugin as a whole and must agree across every child. Validate with
	// grc-store-protocol/spdx; use a LicenseRef-… token for custom licenses.
	License    string     `json:"license"`
	Platform   Platform   `json:"platform"`
	Entrypoint string     `json:"entrypoint"`
	Protocol   string     `json:"protocol"`
	Evaluates  []Evaluate `json:"evaluates"` // non-empty; each RequirementIDs non-empty (else apierror.MalformedIndex)
}

// Platform is a child's OS/arch. darwin/universal is encoded as two children
// (darwin/amd64 + darwin/arm64) over one binary blob — there is no "universal"
// arch value.
type Platform struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
}

// Evaluate is one entry in a config's evaluates list: a Layer 2 ControlCatalog
// coordinate at a specific version plus the assessment-requirement IDs the
// plugin implements.
type Evaluate struct {
	Catalog        string   `json:"catalog"` // <namespace>/<catalog_id>
	CatalogVersion string   `json:"catalog_version"`
	RequirementIDs []string `json:"requirement_ids"`
}
