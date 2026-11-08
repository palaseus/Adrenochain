# Adrenochain Kubernetes Manifests

This directory contains Kubernetes deployment manifests for Adrenochain, referenced by
`scripts/deploy.sh` in `kubernetes` mode:

```bash
./scripts/deploy.sh kubernetes
```

## Manifests

| File | Purpose |
|------|---------|
| `namespace.yaml` | `adrenochain` namespace |
| `rbac.yaml` | ServiceAccount + Role/RoleBinding (least-privilege) |
| `configmap.yaml` | Production configuration (`production.yaml`) |
| `secret.example.yaml` | **Example** secret template — copy and seal before use |
| `deployment.yaml` | Deployment, ClusterIP + LoadBalancer Services, probes, resource limits |
| `autoscaling.yaml` | HPA, PodDisruptionBudget, NetworkPolicy |
| `kustomization.yaml` | Kustomize entry point |

## Prerequisites

- `kubectl` configured against your cluster
- Docker image `adrenochain:latest` built and pushed to a registry reachable by the cluster
  (update `images[].newName`/`newTag` in `kustomization.yaml`)

## Secrets management

Real secrets must never be committed to git:

```bash
# 1. Copy the example
cp scripts/deployment/k8s/secret.example.yaml scripts/deployment/k8s/secret.yaml

# 2. Base64-encode your values and edit secret.yaml
echo -n "0xYourAdmin..." | base64

# 3. Apply
kubectl apply -k scripts/deployment/k8s/ -n adrenochain
```

Prefer a secrets manager (HashiCorp Vault, AWS Secrets Manager, Sealed Secrets, or
External Secrets) over raw Kubernetes Secrets in production.

## Storage

The Deployment mounts `emptyDir` volumes for `/app/data` and `/app/logs`. These are
ephemeral and are **not** suitable for production persistence. For production, replace
the `emptyDir` volumes with a `PersistentVolumeClaim` backed by your cluster's default
StorageClass, or bind to hostPath/NFS/CSI volumes:

```yaml
volumes:
  - name: data
    persistentVolumeClaim:
      claimName: adrenochain-data
  - name: logs
    persistentVolumeClaim:
      claimName: adrenochain-logs
```

See `docs/DEPLOYMENT.md` for the full Docker volume and deployment documentation.
