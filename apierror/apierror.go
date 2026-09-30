// SPDX-License-Identifier: Apache-2.0

// Package apierror is the grc.store hub's JSON error envelope and its stable
// error-code vocabulary.
//
// Codes are string constants a consumer MAY branch on
// (e.g. `if env.Error == apierror.PluginUnsigned`) without coupling to HTTP
// status — status is the hub's, kept in the trailing doc comments here, not a
// branchable map. A consumer is never required to branch; opaque pass-through
// stays valid. (They are untyped string constants, not a named Code type — a
// raw string compares equal to them, by design.)
package apierror

// Envelope is the JSON body the hub returns on error: {"error","detail"}.
// (Named Envelope, not Response, to avoid colliding with registrytoken.Response
// when both are imported.)
type Envelope struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

// Stable error codes. The trailing comment is the HTTP status the hub uses.
const (
	// Plugin publish / sync path (ADR-0034, as amended).
	PluginUnsigned             = "plugin_unsigned"              // 422 — no signature bundle attached to the index
	PluginVerificationFailed   = "plugin_verification_failed"   // 422 — signature present but invalid / fail-closed trust error
	PluginSignerMismatch       = "plugin_signer_mismatch"       // 422 — signer identity disagrees with the TOFU-pinned one
	PluginVersionImmutable     = "plugin_version_immutable"     // 409 — (ns,id,version) already published with different content
	PluginCoordinateMismatch   = "plugin_coordinate_mismatch"   // 422 — signed config's plugin field disagrees with the published <ns>/<id> coordinate
	PluginSyncNotEnabled       = "plugin_sync_not_enabled"      // 501 — hub verifier not configured (operational)
	InteractivePublishDisabled = "interactive_publish_disabled" // 403 — interactive publish off; use CI trusted publishing
	TagVersionMismatch         = "tag_version_mismatch"         // 422 — config version != index tag
	MalformedIndex             = "malformed_index"              // 422 — bad child media types / missing layers / empty evaluates
	RegistryNotAllowed         = "registry_not_allowed"         // 400 — repository not the <ns>/plugins/<id> shape

	// Catalog publish / sync path (ADR-0045). The catalog analogue of the
	// plugin codes above: the hub verifies a catalog bundle's Sigstore
	// signature (discovered as an OCI referrer of the bundle) at ingest and
	// TOFU-pins the canonical signer identity per (namespace, catalog_id).
	// Same reject matrix, same 422 status, same identity scheme as plugins —
	// one trust model across both artifact kinds.
	CatalogUnsigned           = "catalog_unsigned"            // 422 — no signature referrer found on the bundle
	CatalogVerificationFailed = "catalog_verification_failed" // 422 — signature present but invalid / fail-closed trust error
	CatalogSignerMismatch     = "catalog_signer_mismatch"     // 422 — signer identity disagrees with the TOFU-pinned one

	// Publication-license requirement (ADR-0037) — shared across the catalog/bundle
	// sync path and the plugin sync path; the carrier differs (catalog bundles use
	// the org.opencontainers.image.licenses OCI annotation; plugins use the signed
	// pluginspec.Config.License field), the code does not. Rejected when a license
	// is absent or not a well-formed SPDX expression; a well-formed-but-unknown id
	// is accepted (no version-skew false-reject). To resolve: declare an SPDX
	// expression — see https://spdx.org/licenses — or a LicenseRef-… token for a
	// custom/proprietary license. 422 (the bundle parsed and is otherwise valid but
	// violates the license precondition), consistent with the policy-precondition
	// reading of the 422 codes above rather than the field-shape 400s.
	LicenseRequired = "license_required" // 422 — no valid publication license declared

	// EvaluationLog publish / sync path. Results ride the bundle
	// path (ADR-0034 d.10), so a log is ALSO subject to every bundle code above
	// — catalog_unsigned / catalog_verification_failed / license_required /
	// tag_version_mismatch / catalog_coordinate_mismatch — plus these. All are
	// 422 (the bundle parsed; a relationship or precondition is violated), the
	// same reading as tag_version_mismatch, except results_ingest_disabled
	// (501). The namespace is the publish coordinate's, authorized by
	// ownership (403 forbidden); the codes below are the cross-checks between
	// the log's own claims and the hub's verified state, then the hub ADR-0055
	// acceptance decision: a log is stored only when its Sigstore certificate
	// names the operator-configured pvtr-publish-results workflow release on a
	// GitHub-hosted runner, the bearer is that run, and signed provenance binds
	// it to a published plugin digest. There is no unverified tier.
	TargetNotFound              = "target_not_found"               // 422 — log.target names no registered, non-archived target in the publishing namespace
	TargetNotVerified           = "target_not_verified"            // 422 — the target exists but its ownership has not been verified
	TargetNotOwned              = "target_not_owned"               // 422 — log.target's namespace is not the namespace the bundle was published to
	EvaluationLogTargetMismatch = "evaluation_log_target_mismatch" // 422 — log.target is not a usable hub coordinate (id not <ns>/<slug>, version empty) or metadata.version does not start with "<target.version>-"
	EvaluatorUnpublished        = "evaluator_unpublished"          // 422 — metadata.author.id is not a hub plugin coordinate, differs from the provenance's evaluator, or no published+signed+live plugin exists at the provenance digest
	EvaluationLogTooLarge       = "evaluation_log_too_large"       // 422 — bundle exceeds limits.MaxEvaluationLogBundleBytes
	ResultsIngestDisabled       = "results_ingest_disabled"        // 501 — the hub has no results trust root (HUB_RESULTS_TRUST_ROOT) or no verifier; every results publish is refused
	ResultsSignerUntrusted      = "results_signer_untrusted"       // 422 — certificate issuer is not GitHub Actions, SAN != the configured job_workflow_ref (ref included), or the runner is not github-hosted
	ResultsCallerMismatch       = "results_caller_mismatch"        // 422 — the bearer is not a GitHub Actions OIDC token, or its repository/ref differ from the certificate's caller repository/ref
	ResultsProvenanceMissing    = "results_provenance_missing"     // 422 — no mediatype.ProvenanceBundle referrer on the bundle manifest
	ResultsProvenanceInvalid    = "results_provenance_invalid"     // 422 — provenance fails verification, was not minted by the run that signed the log, or its subject/predicate/target disagree with the log

	// Deprecated: never emitted since ADR-0055 — results streams carry no
	// TOFU signer pin; the signer is fixed by HUB_RESULTS_TRUST_ROOT. Kept so
	// v0.6.0 importers compile.
	EvaluationLogSignerMismatch = "evaluation_log_signer_mismatch"

	// Target registry: POST /v1/targets/{ns}/{id}/verify.
	TargetVerificationFailed = "target_verification_failed" // 422 — the ownership proof did not match (OIDC repository claim, DNS TXT, or well-known body disagrees with the target / challenge)

	// AI-assisted drafting (hub ADR-0059): the per-user provider credential
	// (/v1/me/ai-credential) and POST /v1/ai/polish. The hub makes the
	// provider call with the caller's own stored token; these codes tell the
	// client which side to fix. Every route is hub-admin only for now (the
	// drafting flow's gate), so 401/403 there are the usual unauthorized /
	// forbidden.
	AINotEnabled          = "ai_not_enabled"          // 501 — the hub has no credential-encryption key (AI_CREDENTIAL_KEY); nothing AI-related is available
	AICredentialMissing   = "ai_credential_missing"   // 412 — the caller has not stored a provider token; add one on the settings page
	AIProviderRejected    = "ai_provider_rejected"    // 422 — the provider refused the stored token (401/403 upstream); rotate it
	AIProviderUnavailable = "ai_provider_unavailable" // 502 — the provider errored, timed out or returned an unusable body
	AIContextTooLarge     = "ai_context_too_large"    // 413 — request exceeds limits.MaxAIPolishRequestBytes, MaxAIPolishCurrentBytes or MaxAIReviewRequestBytes
	AIRateLimited         = "ai_rate_limited"         // 429 — the caller's per-user polish budget is spent; wait a moment
	// Hub-held drafts (hub ADR-0060): /v1/namespaces/{slug}/drafts.
	DraftTooLarge = "draft_too_large" // 413 — body exceeds limits.MaxDraftBodyBytes; trim the draft
	DraftConflict = "draft_conflict"  // 409 — PUT's expected_updated_at is stale: the draft was saved elsewhere; reload before saving again

	// Enterprise accounts (hub ADR-0061): /v1/enterprises/* and the owner
	// argument on POST /v1/namespaces. Enterprise shapes stay hub-internal;
	// only the codes are shared so clients can branch on them.
	EnterpriseNotFound           = "enterprise_not_found"            // 404 — no active enterprise with that slug
	EnterpriseNotAdmin           = "enterprise_not_admin"            // 403 — caller holds no open admin/owner span for the enterprise (membership, owned namespaces, enterprise-owned create)
	EnterpriseNotOwner           = "enterprise_not_owner"            // 403 — owner-only action (billing report, granting or revoking owner) from an admin or member
	EnterpriseUserAlreadyManaged = "enterprise_user_already_managed" // 409 — the user already holds an open span with another enterprise; a move is a transfer (REV-419)
	EnterpriseOwnsNamespaces     = "enterprise_owns_namespaces"      // 409 — archive refused while the enterprise still owns namespaces; transfer them out first
	NamespacePersonalImmovable   = "namespace_personal_immovable"    // 422 — a personal namespace is always owned by its user and cannot be enterprise-owned
	EnterpriseArchived           = "enterprise_archived"             // 409 — the enterprise is archived; it confers no authority and accepts no writes

	// Shared transport / drift codes.
	CoordinateMismatch = "coordinate_mismatch" // 400 — request body repository != URL coordinate
	Forbidden          = "forbidden"           // 403 — caller lacks ownership / write authority
	RegistryDiverged   = "registry_diverged"   // 502 — registry digest drifted after ingest (read path)
)
