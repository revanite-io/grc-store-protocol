// SPDX-License-Identifier: Apache-2.0

// Package sync is the request body for the hub sync endpoints:
// POST /v1/bundles/sync and POST /v1/plugins/{ns}/{id}/sync.
package sync

// Request tells the hub which already-pushed artifact to fetch and index.
// Repository is a path within the registry (no host or scheme), e.g.
// "<ns>/plugins/<id>" for a plugin. Everything authoritative (identity,
// evaluates, entrypoint) comes from the signed artifact, never from this body.
type Request struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}
