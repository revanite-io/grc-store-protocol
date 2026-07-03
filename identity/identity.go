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
// The canonical string is opaque to equality-only consumers, but it is NOT
// off-limits to parse: a consumer needing the parts back (grcli's
// verify-by-coordinate, ADR-0045 decision 8) parses it VIA ParseKeyless ONLY —
// the inverse defined here — never by splitting the string itself, so the
// producer and every parser share one definition of the format.
//
// STABILITY: the OUTPUT of CanonicalKeylessIdentity is frozen contract. Any
// change to what it returns for a given input is BREAKING — major version bump
// and a heads-up to every importer — even before v1.0.0. ParseKeyless is its
// exact inverse and moves in lockstep.
package identity

import (
	"errors"
	"strings"
)

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
// The returned string is treated as OPAQUE by consumers that only TOFU-compare
// it for equality (the common case). When a consumer genuinely needs the parts
// back — e.g. grcli composing a cosign `--certificate-oidc-issuer` /
// `--certificate-identity-regexp` invocation for verify-by-coordinate
// (ADR-0045 decision 8) — it MUST use ParseKeyless, this function's inverse,
// rather than splitting the string itself: parsing lives in the package that
// owns the format so producer and parser cannot drift.
func CanonicalKeylessIdentity(issuer, san string) string {
	return KeylessScheme + issuer + "#" + StripWorkflowRef(san)
}

// ErrUnknownScheme is returned by ParseKeyless when the input does not carry
// the keyless: scheme prefix (e.g. a key:<fpr> identity, or a non-identity
// string). Consumers branch on it to fall back to their explicit-trust path.
var ErrUnknownScheme = errors.New("identity: not a keyless: identity")

// ErrMissingSeparator is returned by ParseKeyless when a keyless: identity has
// no "#" separating issuer from workflow path — a malformed identity string.
var ErrMissingSeparator = errors.New("identity: keyless identity has no '#' separating issuer and workflow path")

// ParseKeyless is the inverse of CanonicalKeylessIdentity: it splits a
// canonical "keyless:<issuer>#<workflow-path>" string back into its issuer and
// ref-stripped workflow path. It strips the KeylessScheme prefix (ErrUnknownScheme
// if absent) and splits at the FIRST "#" (ErrMissingSeparator if none) — the
// same "#" CanonicalKeylessIdentity joined on.
//
// Round-trip: for any issuer that contains no "#" (every real OIDC issuer is a
// bare https URL, so this always holds in practice),
//
//	ParseKeyless(CanonicalKeylessIdentity(issuer, san)) == (issuer, StripWorkflowRef(san), nil)
//
// The workflowPath returned is already ref-stripped (CanonicalKeylessIdentity
// strips the "@refs/..." suffix before joining), so it is the pinnable workflow
// path, not a per-run ref. Splitting at the FIRST "#" means an issuer that
// itself contained "#" would not round-trip — the same non-reversibility caveat
// CanonicalKeylessIdentity carries; it does not arise for GitHub Actions issuers.
func ParseKeyless(canonical string) (issuer, workflowPath string, err error) {
	rest, ok := strings.CutPrefix(canonical, KeylessScheme)
	if !ok {
		return "", "", ErrUnknownScheme
	}
	issuer, workflowPath, found := strings.Cut(rest, "#")
	if !found {
		return "", "", ErrMissingSeparator
	}
	return issuer, workflowPath, nil
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
