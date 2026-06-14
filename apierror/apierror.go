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

	// Shared transport / drift codes.
	CoordinateMismatch = "coordinate_mismatch" // 400 — request body repository != URL coordinate
	Forbidden          = "forbidden"           // 403 — caller lacks ownership / write authority
	RegistryDiverged   = "registry_diverged"   // 502 — registry digest drifted after ingest (read path)
)
