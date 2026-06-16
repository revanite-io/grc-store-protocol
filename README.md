# grc-store-protocol

The **wire contract** for [grc.store](https://grc.store) — the media types, error
codes, size limits, identity-canonicalization rule, and request/response shapes a
producer must speak to publish to a grc.store hub. One open, versioned definition
that the hub, [`grcli`](https://github.com/revanite-io/grcli), and external
producers (e.g. Privateer's `pvtr`) all import, so the contract can't silently
drift between independently-maintained copies.

> **Contract, not client.** This module holds **constants, types, and pure
> functions** only — no network calls, no auth flows, no signing, no registry
> client. If something here ever needs to open a socket or import `oras` /
> `sigstore-go` / an OIDC library, it does not belong here. Auth, signing, push,
> and domain assembly stay in each consumer, where they legitimately differ.

## Packages

| Package | What |
|---|---|
| `identity` | Keyless signer-identity canonicalization (`CanonicalKeylessIdentity`) — the load-bearing invariant the hub TOFU-pins and producers re-derive. **Its output is frozen contract.** |
| `mediatype` | OCI media types (plugin config/binary, Sigstore bundle) + the referrer artifactType note. |
| `apierror` | The `{error, detail}` envelope and the stable error-code vocabulary. |
| `pluginspec` | The signed plugin config-blob schema. |
| `discovery` | The `/.well-known/ext.grc-store` document. |
| `registrytoken` | The `GET /v2/token` response. |
| `syncapi` | The sync request + response shapes. |
| `limits` | Producer-facing size limits. |
| `spdx` | SPDX license-expression validation + canonicalization for the publication-license field (ADR-0036). grcli is strict (`Canonicalize`); the hub is lenient (`Parse`/`String`). |

> `pluginspec` and `syncapi` are deliberately *not* named `plugin`/`sync` — those
> would shadow the standard library's `plugin` and `sync` packages in consumers.

## Usage

```go
import (
	"github.com/revanite-io/grc-store-protocol/identity"
	"github.com/revanite-io/grc-store-protocol/apierror"
	"github.com/revanite-io/grc-store-protocol/pluginspec"
)

// Canonicalize a keyless signer identity for TOFU comparison (compare-only,
// never parse the result — see the note in identity).
id := identity.CanonicalKeylessIdentity(issuer, fulcioSAN)

// Branch on a hub error code (optional — opaque pass-through is also valid).
var env apierror.Envelope
_ = json.Unmarshal(body, &env)
if env.Error == apierror.PluginSignerMismatch { /* ... */ }

// Build the signed plugin config descriptor.
cfg := pluginspec.Config{Plugin: "<ns>/<plugin_id>", Version: tag /* == index tag */, /* ... */}
```

## Scope notes

- **Registered-key identities (`key:sha256:<fpr>`) are out of scope for v1.** The hub
  keyless-pins only; a matching `CanonicalKeyIdentity` arrives if/when that changes.
- The canonical identity string is **opaque** — compare it, never split it apart.

## Stability

Pre-`v1.0.0` (`v0.x`) while the hub, `grcli`, and `pvtr` adopt it. One rule is
stricter than semver from day one: **any change to `identity.CanonicalKeylessIdentity`'s
output is breaking** (major bump + heads-up to importers) — a one-character drift
silently breaks every publisher's *second* release.

**How breaking changes are signaled:** a GitHub **release** on this repo, with
`BREAKING:` in the title for any output/wire-shape change. External importers
should watch releases — that is the contract's announcement channel, not the issue
tracker.

## License

[Apache-2.0](LICENSE).
