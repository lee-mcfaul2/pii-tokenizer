# pii-tokenizer

Crypto plane for the AI Agent Security Platform. Tokenizes PII via AES-SIV-CMAC (RFC 5297) under per-request keys, with the per-request keys themselves wrapped by an HSM-backed `K_master`.

Spec: `ai-security/docs/superpowers/specs/2026-05-12-pii-tokenizer-design.md` in the umbrella workspace.

## Quickstart

```
make build
make test
make run-local       # boots tokenizer + Redis via Docker Compose with aead-local backend
```

## Layout

- `cmd/tokenizer/` — entrypoint
- `internal/` — core packages (crypto, kmaster, store, api, …)
- `tools/rotate-kmaster/` — K_master rotation CLI
- `deploy/` — Dockerfile + Helm chart fragment
- `api/` — OpenAPI spec
- `tests/` — integration, e2e, conformance
- `docs/` — api.md, backends.md, rotation.md, failure-modes.md, threat-model.md

## License

Apache 2.0.
