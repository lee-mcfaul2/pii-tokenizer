# pii-tokenizer HTTP API

Source of truth: [`api/openapi.yaml`](../api/openapi.yaml).

## Endpoints

| Endpoint | Method | Purpose |
|---|---|---|
| `/v1/init_request` | POST | Create per-request K_request |
| `/v1/tokenize` | POST | Encrypt plaintext → token |
| `/v1/detokenize` | POST | Decrypt token → plaintext |
| `/v1/release_request` | POST | Best-effort delete; TTL backstop |
| `/healthz` | GET | Liveness |
| `/readyz` | GET | Readiness (Redis + K_master) |
| `/metrics` | GET | Prometheus |

## Caller authentication

Linkerd `AuthorizationPolicy` (in the Helm chart) restricts the caller to `agent-gateway`'s mesh identity. The tokenizer process does not re-check caller identity — the mesh is trusted.

## Error catalog

| `error_type` | HTTP | Trigger |
|---|---|---|
| `SCOPE_NOT_FOUND` | 404 | UUID has no Redis entry |
| `SCOPE_EXPIRED` | 410 | Entry past `exp` |
| `AAD_MISMATCH` | 400 | Token tampered / wrong UUID / wrong type |
| `SCHEMA_VALIDATION_FAILED` | 400 | Bad request body |
| `INVALID_PII_TYPE` | 400 | Type not in enum |
| `KMASTER_UNAVAILABLE` | 503 | KMS/HSM errored (retriable) |
| `REDIS_UNAVAILABLE` | 503 | Redis errored (retriable) |
| `INTERNAL_ERROR` | 500 | Unhandled (retriable) |

## Idempotency

`/v1/init_request` is idempotent: re-calling with the same `request_uuid` returns the existing `expires_at` without extending it.
