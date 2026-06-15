// SPDX-License-Identifier: Apache-2.0

// Package syncapi is the request and response shapes for the hub sync endpoints:
// POST /v1/bundles/sync and POST /v1/plugins/{ns}/{id}/sync. (Named syncapi, not
// sync, so it does not shadow the standard library's sync package in consumers.)
package syncapi

// Request tells the hub which already-pushed artifact to fetch and index.
// Repository is a path within the registry (no host or scheme), e.g.
// "<ns>/plugins/<id>" for a plugin. Everything authoritative (identity,
// evaluates, entrypoint) comes from the signed artifact, never from this body.
type Request struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

// Response is the hub's reply to a catalog/bundle sync (POST /v1/bundles/sync).
// Repository and Tag echo the request; the rest summarize what was indexed.
// (The plugin sync path returns a different, plugin-specific shape that consumers
// read opaquely, so it is not modeled here.)
type Response struct {
	Repository    string   `json:"repository"`
	Tag           string   `json:"tag"`
	ManifestEtag  string   `json:"manifest_etag"`
	ArtifactCount int      `json:"artifact_count"`
	NewCount      int      `json:"new_count"`
	Types         []string `json:"types"`
}
