# K_master backends

The tokenizer's master key (`K_master`) is the KEK that wraps each per-request `K_request`. Four backends ship; selectable via Helm value `kmaster.backend`.

## aead-local (default for demo)

- 32-byte AES key stored as a K8s `Secret`.
- Mounted read-only at `/etc/tokenizer/kmaster/` with files `k_master_v<N>` and `current`.
- **Not production-grade**: keys live in etcd.

### Setup

```bash
kubectl create secret generic pii-tokenizer-kmaster -n platform \
  --from-file=k_master_v1=<(openssl rand 32) \
  --from-literal=current=v1
```

### Rotation

```bash
kubectl exec -n platform deploy/pii-tokenizer -- rotate-kmaster rotate-aead -d /etc/tokenizer/kmaster
```

Tokenizer pods reload on Secret update.

## vault-transit (production)

- HashiCorp Vault Transit secrets engine.
- Tokenizer authenticates via K8s ServiceAccount (Vault's Kubernetes auth method).

### Setup

```bash
vault secrets enable transit
vault write -f transit/keys/pii-tokenizer
vault write auth/kubernetes/role/pii-tokenizer \
  bound_service_account_names=pii-tokenizer-sa \
  bound_service_account_namespaces=platform \
  policies=pii-tokenizer-transit \
  ttl=1h
```

Set Helm values:
```yaml
kmaster:
  backend: vault-transit
  vault:
    addr: http://vault.platform.svc.cluster.local:8200
    keyName: pii-tokenizer
```

### Rotation

```bash
vault write -f transit/keys/pii-tokenizer/rotate
```

Tokenizer is unaware; Vault handles versioning transparently.

## aws-kms (production)

- AWS KMS CMK.
- Auth via IRSA (annotation on `pii-tokenizer-sa`).

### Setup

1. Create CMK in AWS KMS console / Terraform.
2. Annotate the SA: `eks.amazonaws.com/role-arn: arn:aws:iam::123:role/pii-tokenizer`.
3. IAM role allows `kms:Encrypt` and `kms:Decrypt` on the CMK.

Set Helm values:
```yaml
kmaster:
  backend: aws-kms
  awsKms:
    region: us-east-1
    keyId: arn:aws:kms:us-east-1:123:key/abcd-1234
```

### Rotation

AWS KMS handles key rotation automatically when enabled:
```bash
aws kms enable-key-rotation --key-id <CMK-ARN>
```

## pkcs11 (deferred)

The pkcs11 backend is currently a no-op stub — `hashicorp/go-kms-wrapping` does not publish a pkcs11 wrapper on the Go proxy, and our threat model is satisfied by directing operators that need HSM-grade key custody to `vault-transit` (configured with an HSM-backed Vault auto-unseal) or `aws-kms` (CloudHSM-backed CMK). Calling `NewPKCS11` returns a clear error pointing to those alternatives.

To enable a direct PKCS#11 integration in a future revision, replace `internal/kmaster/pkcs11.go` with a constructor that wires `github.com/miekg/pkcs11` into a `Wrapper`.
