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

	// SigstoreBundle is the Sigstore v0.3 bundle media type.
	//
	// On the PLUGIN path it is also the referrer ARTIFACT TYPE: a plugin
	// index's signature is the OCI referrer of the index whose
	// artifactType == SigstoreBundle. The producer (pvtr) packs it that way and
	// the hub verifier discovers it by filtering referrers on this value; the
	// two agree (field-confirmed 2026-06).
	//
	// CAVEAT — catalogs differ: grcli signs catalogs with
	// `cosign sign --new-bundle-format`, which emits the same v0.3 bundle BLOB
	// but attaches the referrer with artifactType
	// "https://sigstore.dev/cosign/sign/v1" (verified via `cosign tree`), NOT
	// SigstoreBundle. So "same format" does not imply "same discovery": a future
	// hub catalog-verifier cannot reuse the plugin referrer filter unchanged.
	SigstoreBundle = "application/vnd.dev.sigstore.bundle.v0.3+json"

	// CosignSignReferrer is the referrer artifactType cosign stamps on a
	// `--new-bundle-format` signature (the catalog path), documented here so a
	// future catalog verifier filters on the right value rather than rediscover
	// the divergence above.
	//
	// RULE — do not cross these: discover a PLUGIN signature by filtering
	// referrers on SigstoreBundle; discover a cosign-signed CATALOG signature by
	// filtering on CosignSignReferrer. Using the wrong filter finds zero
	// referrers and treats a signed artifact as unsigned.
	CosignSignReferrer = "https://sigstore.dev/cosign/sign/v1"
)
