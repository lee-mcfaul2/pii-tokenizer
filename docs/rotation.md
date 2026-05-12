# K_master rotation

## Why rotate

`K_master` is a long-lived key. Rotation limits the blast radius of a compromise and is required for some compliance regimes.

## When to rotate

- **Scheduled:** at least annually.
- **On suspicion:** any time a compromise of the cluster, HSM, or operator credentials is suspected.
- **After offboarding:** when a key holder leaves the team.

## Rotation by backend

See [`backends.md`](./backends.md) for the specific rotation command per backend.

## Observing rotation

Two Prometheus gauges show the state of rotation:

- `tokenizer_kmaster_versions_loaded` — count of versions the tokenizer can currently `Unwrap`. Steady state is 1.
- `tokenizer_kmaster_version_in_use{version=N}` — per-version gauge. Set when any Wrap or Unwrap touched version N in the past 60 seconds.

### Healthy rotation timeline

```
t=0     pre-rotation:    versions_loaded=1, in_use{v=1}=non-zero
t=0     rotate           (Vault/AWS automatic; aead-local: rotate-kmaster rotate-aead)
t=1s    versions_loaded=2, in_use{v=1}=non-zero, in_use{v=2}=non-zero
t=2h    most new requests use v=2; v=1 still serves expiring requests
t=2h+TTL  in_use{v=1}=0 for safe window
t=24h   safe to retire v=1 (aead-local: remove the file)
        versions_loaded=1 again
```

### Drift detection

Alert when `tokenizer_kmaster_versions_loaded > 1` for more than 24h. Indicates incomplete cleanup.

## Failure recovery

If a rotation goes wrong (e.g., new version unreadable, current pointer wrong):

1. Restore the previous Secret/key state from backup.
2. Tokenizer pods reload on Secret update or are restarted.
3. Investigate root cause before retrying.

For Vault/AWS backends, rollback is handled by the backend (Vault keeps prior versions; AWS KMS retains old material).
