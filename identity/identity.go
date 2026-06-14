// SPDX-License-Identifier: Apache-2.0

// Package identity canonicalizes signer identities for grc.store TOFU pinning.
//
// It is the load-bearing reason this module exists. A plugin coordinate's
// signer identity is Trust-On-First-Use pinned by the hub at first publish, and
// every later release must produce the *same* canonical string or the hub
// rejects it (422 plugin_signer_mismatch). The producer (which signs and, on
// install, re-extracts the identity) and the hub (which pins and compares) MUST
// agree byte-for-byte. Historically each side carried its own copy of this
// logic, kept in sync only by a comment; a one-character drift would silently
// break every publisher's *second* release. This package makes the agreement
// structural: one definition that the hub, grcli, and pvtr all import.
//
// STABILITY: the OUTPUT of CanonicalKeylessIdentity is frozen contract. Any
// change to what it returns for a given input is BREAKING — major version bump
// and a heads-up to every importer — even before v1.0.0.
package identity

import "strings"

// KeylessScheme prefixes a keyless (Fulcio / OIDC) signer identity. The other
// scheme named in grc.store docs — key:sha256:<fpr> for registered keys — is
// intentionally NOT implemented here: the hub does not TOFU-pin key identities
// today (keyless only). If it ever does, this package gains a matching
// CanonicalKeyIdentity so producers and the hub canonicalize it identically.
const KeylessScheme = "keyless:"

// CanonicalKeylessIdentity encodes a keyless signer as
// "keyless:<oidc-issuer>#<workflow-path>", with the SAN's per-release
// "@refs/..." suffix stripped (see StripWorkflowRef). Stripping the ref is
// load-bearing: a GitHub Actions Fulcio SAN is the per-run workflow ref
// (…/release.yml@refs/tags/v1.1.0), which changes every release; pinning the
// raw SAN would make release N+1 fail signer-identity comparison. The pinned
// identity is the workflow PATH — the correct TOFU granularity ("this plugin is
// produced by this publisher's CI workflow"), not the per-tag ref.
//
// The returned string is OPAQUE: compare it for equality, never parse it back
// apart on the "#" separator — an OIDC issuer or SAN may itself contain "#", so
// splitting is not reversible. TOFU only ever needs equality.
func CanonicalKeylessIdentity(issuer, san string) string {
	return KeylessScheme + issuer + "#" + StripWorkflowRef(san)
}

// StripWorkflowRef drops a trailing "@refs/..." git ref from a workflow SAN,
// leaving the workflow path. SANs without "@refs/" pass through unchanged.
//
// This is exported for display use (showing the workflow path). To produce an
// identity string for hub TOFU comparison, use CanonicalKeylessIdentity — do NOT
// compose one yourself from this, or it won't match what the hub stored.
func StripWorkflowRef(san string) string {
	if before, _, found := strings.Cut(san, "@refs/"); found {
		return before
	}
	return san
}
