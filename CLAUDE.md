# grc-store-protocol — agent orientation

Apache-2.0 Go module that is the **shared wire contract** for grc.store — constants, types, and
pure functions agreed on by the hub, grcli, and pvtr. Module: `github.com/revanite-io/grc-store-protocol`.

**Key principle: this is a contract, not a client.** Zero network calls, auth, signing, or
registry I/O. If code needs sockets or heavy deps (oras, sigstore-go, OIDC), it does **not**
belong here. `README.md` is the full contract overview + package glossary — point there.

## Dev loop
- `go test ./...` (add `-count=1` to skip cache) · `go vet ./...` · `gofmt -l .` (format check is a CI gate)
- **Zero-dependency invariant**: `go mod tidy` must be a no-op — any added dependency is rejected in CI.
- No Makefile. CI matrix: Go 1.22 (floor) and a recent toolchain (`.github/workflows/ci.yml`).

## Packages (each defines one wire concept)
- `apierror` — JSON error envelope + stable error-code vocabulary
- `discovery` — `/.well-known/grc-store-configuration` document
- `identity` — keyless signer-identity canonicalization **(frozen — see gotcha)**
- `limits` — producer-facing size limits (e.g. `MaxPluginBlobBytes = 4 MiB`)
- `slug` — the hub coordinate slug rule (`Slugify`, frozen) + EvaluationLog results coordinate helpers
- `mediatype` — OCI media types (plugin config/binary, Sigstore bundle)
- `pluginspec` — plugin config-blob schema · `registrytoken` — `GET /v2/token` response
- `spdx` — license-expression validation + canonicalization (ADR-0036) · `syncapi` — sync request/response shapes

## Gotchas — editing here is high-blast-radius
- **Consumers: hub, grcli, pvtr.** A breaking change ripples to all three — coordinate releases.
- **`identity.CanonicalKeylessIdentity` output is a frozen contract.** Even a one-character change is a breaking major bump; it silently breaks every publisher's *second* release (identity mismatch on re-publish).
- Identity strings are **opaque** — compare only, never parse. Registered-key identities are out of scope for v1.
- Pre-v1.0.0 (v0.x). Breaking changes are signaled in GitHub release titles (`BREAKING:`), not issues — importers must watch releases.
