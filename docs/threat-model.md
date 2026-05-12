# Threat model

## What pii-tokenizer protects against

1. **Cross-request correlation.** Same plaintext in different requests produces different tokens because each request has a unique `K_request`. An attacker who collects tokens across many requests cannot infer that two tokens decrypt to the same plaintext.
2. **Plaintext exposure on cold storage.** Plaintext is never written to disk, logs, or metrics. Only the encrypted token leaves the tokenizer.
3. **Forward secrecy on request expiry.** Once a request's TTL fires, the wrapped `K_request` is gone. Even an attacker who later steals `K_master` cannot decrypt past tokens.
4. **Token tampering.** AES-SIV-CMAC's synthetic IV doubles as an integrity check; any modification of the token (or `request_uuid` or type in AAD) breaks the SIV verification. Tamper → `AAD_MISMATCH`.
5. **Wrong-request token re-use.** A token from request A cannot be decrypted under request B's K_request — even if both used the same plaintext.

## What pii-tokenizer does NOT protect against

1. **Frequency analysis within a single request.** Same plaintext + same type + same request_uuid → same token. An attacker watching one request can count repetitions. The platform's design (per `design-doc.md` §6.5) accepts this as a deliberate trade-off for referential reasoning by the LLM.
2. **`K_master` compromise during an active request.** If `K_master` and Redis are both compromised concurrently, in-flight tokens can be decrypted. Forward secrecy applies only to expired requests.
3. **Compromise of `agent-gateway`.** The gateway is the only authorized caller; if it's compromised, it can call the tokenizer with arbitrary inputs. The tokenizer's threat model assumes the gateway is honest. See `design-doc.md` §2 trust boundaries.
4. **Side channels on the host.** Cache timing, electromagnetic emanation, etc., are out of scope — the tokenizer assumes the K8s node is not actively malicious. Production deployments should use TEEs or HSMs (`vault-transit` with HSM seal, or `aws-kms` with CloudHSM) to mitigate.
5. **DoS via flooding.** Rate-limiting is the gateway's job; the tokenizer trusts that its single caller is rate-limited upstream.

## Mesh-trust assumption

Linkerd's `AuthorizationPolicy` is the **sole** caller-identity check. The tokenizer does not re-verify mTLS identities in handler code. If the mesh's authz is compromised or misconfigured (e.g., the AuthorizationPolicy CR is deleted), the tokenizer would accept calls from any pod with valid mesh certs. Mitigations:

- The Helm chart bundles the AuthorizationPolicy; uninstalling the chart removes it. Operators should not remove it independently.
- A periodic conformance test (in `secure-agent-demo`) should verify that calls from a non-gateway SPIFFE are rejected end-to-end.

## Audit obligations

Every successful `tokenize` / `detokenize` increments a metric and emits a trace span tagged with the `request_uuid`. Operators can join tokenizer traces with gateway traces by the same `request_uuid` to reconstruct the full request handling. No plaintext appears in either set of telemetry.

## Cryptographic primitives

- **AES-256-SIV-CMAC** (RFC 5297). Nonce-misuse-resistant deterministic authenticated encryption. We pass a nil nonce, so the synthetic IV is derived from `(K_request, AAD, plaintext)` and the same triple always produces the same ciphertext — that's the equality property the LLM relies on within a request. AES-SIV-CMAC's 16-byte synthetic IV is both the integrity check and the IV; there is no separate auth tag to lose.
- **AES-256-GCM** (via `go-kms-wrapping`) for `K_master` envelope encryption in the `aead-local` backend. Vault Transit / AWS KMS / PKCS#11 backends use whatever primitive their KMS exposes; we treat them as opaque oracles.
