# Failure modes runbook

| `error_type` | Trigger | Immediate action (on-call) | Root-cause checks |
|---|---|---|---|
| `SCOPE_NOT_FOUND` | UUID has no Redis entry | None (caller bug or expired). Log only. | Caller's request flow — did `init_request` ever run? Is the UUID typo'd? |
| `SCOPE_EXPIRED` | Entry past `exp` | None. Caller should release+re-init. | Are TTLs configured too short for the workload? |
| `AAD_MISMATCH` | Token tampered / wrong UUID / wrong type | **Page** — possible attack | Inspect audit log for the request_uuid; check for malformed-token volume spikes. |
| `SCHEMA_VALIDATION_FAILED` | Bad request body | None — caller bug. | Check the gateway's tokenizer-client codegen against the OpenAPI spec. |
| `INVALID_PII_TYPE` | Type not in enum | None — caller bug. | Caller is using a type the tokenizer doesn't know; bump schema lib or fix caller. |
| `KMASTER_UNAVAILABLE` | KMS/HSM errored | **Page** if rate > X/min. The tokenizer becomes unable to wrap or unwrap → cascading failure to gateway → 503s upstream. | Backend-specific: Vault auth lapsed? AWS KMS throttled? HSM hung? |
| `REDIS_UNAVAILABLE` | Redis client error | **Page** — this is P0. All operations fail. | Redis pod healthy? Network policy correct? AOF disk full? |
| `INTERNAL_ERROR` | Unhandled panic / unexpected | **Page** | Check container logs and stack traces. |

## Generic ops checklist

1. `kubectl get pods -n platform -l app=pii-tokenizer` — all healthy?
2. `kubectl logs -n platform deploy/pii-tokenizer --tail=200` — recent errors?
3. `kubectl exec -n platform deploy/redis -- redis-cli ping` — Redis up?
4. Check Prometheus dashboards for `tokenizer_errors_total` and `tokenizer_redis_errors_total`.
5. If on `aead-local`: `kubectl get secret pii-tokenizer-kmaster -o yaml` — Secret intact?
6. If on Vault/AWS/PKCS11: backend-specific health checks (see `backends.md`).

## Restart-only mitigation

If a tokenizer pod is in a wedged state but other replicas are healthy, restarting the wedged pod is safe — there is no per-pod state. `kubectl delete pod ...` triggers a fresh start.
