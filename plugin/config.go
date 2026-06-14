// SPDX-License-Identifier: Apache-2.0

// Package plugin is the plugin config-blob schema (mediatype.PluginConfig) —
// the signed descriptor a bare binary can't self-carry (ADR-0034 decision 2).
// It is the producer↔hub shared type: the publisher (pvtr) writes and signs it;
// the hub reads it as the authoritative source on sync.
package plugin

// Config is one plugin config blob, one per child image manifest. Platform is
// per-child; Plugin, Version, Entrypoint, Protocol, and Evaluates describe the
// plugin as a whole and must agree across every child.
type Config struct {
	// Plugin is the <namespace>/<plugin_id> REGISTRY COORDINATE the index is
	// published under — NOT a GitHub owner/repo. A mismatch with the coordinate
	// is rejected (apierror.PluginCoordinateMismatch).
	Plugin string `json:"plugin"`
	// Version must equal the index tag (else apierror.TagVersionMismatch).
	Version    string     `json:"version"`
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
