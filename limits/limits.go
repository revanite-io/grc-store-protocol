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

// MaxEvaluationLogBundleBytes is the hub's ingest cap on the whole pushed
// EvaluationLog bundle (manifest + every layer the hub reads) on
// POST /v1/bundles/sync. A results publish whose bundle exceeds it is
// rejected (422 apierror.EvaluationLogTooLarge) BEFORE the body is parsed.
//
// 16 MiB, because a log is unbounded producer output that anyone owning a
// verified target can publish at every scan, and every accepted byte is
// stored twice (the OCI blob in R2, the indexed body in Postgres) for the
// lifetime of an immutable version — so the cap bounds both storage cost and
// the write-amplification an abusive or runaway pipeline can cause. 16 MiB is
// ~4x the largest real pvtr log seen (a full-catalog run with per-step
// messages); a producer that hits it should split by catalog, which the
// results coordinate (one stream per target × catalog) already does.
const MaxEvaluationLogBundleBytes = 16 << 20 // 16 MiB
