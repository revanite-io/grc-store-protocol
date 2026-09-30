// SPDX-License-Identifier: Apache-2.0

// Package limits holds producer-facing size limits of the grc.store wire
// contract — limits a client must honor to avoid rejection, whether it is a
// publisher pushing a bundle or a web editor saving a draft. Server-side
// shaping limits (pagination, internal caps) and client-side pull caps stay
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
// verified target can publish at every scan, and every accepted byte is stored
// in R2 for the lifetime of an immutable version and parsed in full at ingest
// (the hub indexes coordinates and a rollup, not the body) — so the cap bounds
// both storage cost and the work an abusive or runaway pipeline can cause.
// 16 MiB is
// ~4x the largest real pvtr log seen (a full-catalog run with per-step
// messages); a producer that hits it should split by catalog, which the
// results coordinate (one stream per target × catalog) already does.
const MaxEvaluationLogBundleBytes = 16 << 20 // 16 MiB

// MaxAICredentialTokenBytes caps the provider token a user stores with
// PUT /v1/me/ai-credential. Real keys are a few hundred bytes; anything
// past this is not a key.
const MaxAICredentialTokenBytes = 1 << 10 // 1 KiB

// MaxAIPolishRequestBytes caps the whole POST /v1/ai/polish body (the
// context included) and MaxAIPolishCurrentBytes the field text being
// polished. Over either the hub answers 413 apierror.AIContextTooLarge
// before touching the provider. A control catalog's title, groups and
// control objectives fit in a few KiB; the caps leave room for large
// catalogs while bounding what one call can send upstream on the user's
// own token.
const (
	MaxAIPolishRequestBytes = 64 << 10 // 64 KiB
	MaxAIPolishCurrentBytes = 8 << 10  // 8 KiB
)

// MaxAIReviewRequestBytes caps the whole POST /v1/ai/review body. A
// duplicates review sends every requirement in the catalog, so it is
// four times the polish cap; over it the hub answers 413
// apierror.AIContextTooLarge before touching the provider.
const MaxAIReviewRequestBytes = 256 << 10 // 256 KiB

// MaxDraftBodyBytes caps the body of a hub-held draft (POST/PUT
// /v1/namespaces/{slug}/drafts, hub ADR-0060). Over it the hub answers 413
// apierror.DraftTooLarge and writes nothing. The body is the editor's own
// state for one catalog; real catalogs run to a few hundred KiB, so 2 MiB
// leaves headroom while bounding what one save can put in Postgres.
const MaxDraftBodyBytes = 2 << 20 // 2 MiB
