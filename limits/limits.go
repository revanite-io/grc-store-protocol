// SPDX-License-Identifier: Apache-2.0

// Package limits holds producer-facing size limits of the grc.store wire
// contract — only limits a publisher must honor to avoid rejection. Server-side
// shaping limits (pagination, request-body caps) and client-side pull caps stay
// in their owners; they are not contract.
package limits

// MaxPluginBlobBytes is the hub's ingest cap on each non-binary plugin blob
// (the index, child manifests, the config blobs, and the signature bundle). A
// publisher whose config/bundle exceeds this is rejected. The plugin BINARY
// layer is not bounded by this — the hub digest-pins it but does not read it.
const MaxPluginBlobBytes = 4 << 20 // 4 MiB
