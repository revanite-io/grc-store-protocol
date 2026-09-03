// SPDX-License-Identifier: Apache-2.0

// Package mediatype holds the OCI media types of the grc.store wire contract.
package mediatype

const (
	// PluginConfig is the per-child config blob: the descriptor a bare binary
	// can't self-carry (plugin coordinate, version, platform, entrypoint,
	// protocol, evaluates).
	PluginConfig = "application/vnd.grc-store.plugin.config.v1+json"

	// PluginBinary is the single layer carrying a plugin binary per child.
	PluginBinary = "application/vnd.grc-store.plugin.binary.v1"

	// SigstoreBundle is the Sigstore v0.3 bundle media type: the media type of
	// the LAYER carrying the bundle JSON inside a signature referrer, and ONE
	// of the two referrer artifactTypes a signature can be attached under (see
	// the RULE below).
	SigstoreBundle = "application/vnd.dev.sigstore.bundle.v0.3+json"

	// CosignSignReferrer is the OTHER referrer artifactType a signature can be
	// attached under (see the RULE below).
	CosignSignReferrer = "https://sigstore.dev/cosign/sign/v1"

	// RULE — accept BOTH referrer artifactTypes when discovering a signature.
	// The stamp is a function of the SIGNER's tooling, not of the artifact
	// kind (plugin vs catalog):
	//
	//	cosign 2.6.x `sign --new-bundle-format` → CosignSignReferrer
	//	cosign 3.x   `sign` (bundle by default) → SigstoreBundle
	//	pvtr's plugin packer                     → SigstoreBundle
	//
	// The v0.3 bundle LAYER inside is identical either way (layer media type
	// SigstoreBundle). Filtering referrers on a single artifactType silently
	// classifies the other cohort's signed artifacts as unsigned — under
	// ADR-0045 that rejects their publishes at ingest, and on the client it
	// reports signed artifacts as unverifiable. Field-confirmed against a
	// live zot with both cosign lines, 2026-07-07 (hub fix: ociref; grcli
	// fix: f8f1eb8).
	//
	// (This supersedes the earlier "do not cross these" rule, which mapped
	// artifactTypes to artifact KINDS. Its premise — that catalogs are always
	// signed by cosign 2.x — broke when cosign 3.x made the bundle format the
	// default and changed the stamp.)

	// ProvenanceBundle is the referrer artifactType (and layer media type) of a
	// SIGNED PROVENANCE attestation: a Sigstore v0.3 bundle whose DSSE in-toto
	// statement carries a SLSA v1 predicate over the same subject digest as the
	// signature. It is deliberately distinct from the two signature types above
	// so signature discovery, which filters on those and takes the first match,
	// can never pick up the provenance referrer as the signature.
	ProvenanceBundle = "application/vnd.grc-store.provenance.bundle.v0.3+json"
)
