# k8s-cleanup

[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/k8s-cleanup)](https://artifacthub.io/packages/search?repo=k8s-cleanup)
[![CI](https://github.com/onurbalmeida/k8s-cleanup/actions/workflows/ci.yml/badge.svg)](https://github.com/onurbalmeida/k8s-cleanup/actions/workflows/ci.yml)
[![Release](https://github.com/onurbalmeida/k8s-cleanup/actions/workflows/release.yml/badge.svg)](https://github.com/onurbalmeida/k8s-cleanup/actions/workflows/release.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/onurbalmeida/k8s-cleanup.svg)](https://pkg.go.dev/github.com/onurbalmeida/k8s-cleanup)
[![Go Report Card](https://goreportcard.com/badge/github.com/onurbalmeida/k8s-cleanup)](https://goreportcard.com/report/github.com/onurbalmeida/k8s-cleanup)
[![Version](https://img.shields.io/github/v/tag/onurbalmeida/k8s-cleanup?label=version)](https://github.com/onurbalmeida/k8s-cleanup/releases)
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](LICENSE)
[![GHCR](https://img.shields.io/badge/OCI-ghcr.io%2Fonurbalmeida%2Fk8s--cleanup-5965e0?logo=github)](https://ghcr.io/onurbalmeida/k8s-cleanup)
![Arch](https://img.shields.io/badge/arch-amd64%20%7C%20arm64-informational)

A fast, safe CLI to clean up **old Pods and Jobs** in Kubernetes. Defaults to **dry-run**, supports **label/field selectors**, **all-namespaces**, **concurrency**, **JSON audit**, and can run as a **CronJob via Helm**.

---

## Features

- Clean up Pods (`Completed`, `Failed`, `Evicted`) and Jobs (`Succeeded`, `Failed`)
- Dry-run by default, with JSON output and NDJSON audit file
- All-namespaces mode with exclusions and label/field selectors
- Concurrency for faster deletions
- Exit codes that integrate with CI
- Shell completions and one-line `version` like `kind`

## Install

### Option 1: Binary (Releases)
Download from [Releases](https://github.com/onurbalmeida/k8s-cleanup/releases) for your OS/arch, then:

```bash
chmod +x k8s-cleanup-<os>-<arch>
sudo mv k8s-cleanup-<os>-<arch> /usr/local/bin/k8s-cleanup
```

### Option 2: Go
```bash
go install github.com/onurbalmeida/k8s-cleanup@latest
```

### Option 3: Docker/OCI (GHCR)
```bash
docker run --rm -v $HOME/.kube:/root/.kube:ro \
  ghcr.io/onurbalmeida/k8s-cleanup:<tag> \
  run --all-namespaces --older-than 24h --dry-run
```

### Option 4: Helm (CronJob)
```bash
helm registry login ghcr.io
helm install cleanup oci://ghcr.io/onurbalmeida/charts/k8s-cleanup --version <chart_version> \
  --set image.repository=ghcr.io/onurbalmeida/k8s-cleanup \
  --set image.tag=<app_version> \
  --set schedule="0 2 * * *" \
  --set args.allNamespaces=true \
  --set "args.excludeNamespaces={kube-system,kube-public,local-path-storage}"
```

> Image and chart are published by CI for tags `vX.Y.Z`.

---

## Quickstart

Show help:
```bash
k8s-cleanup --help
k8s-cleanup run --help
```

Dry-run on all namespaces, 24h+:
```bash
k8s-cleanup run --all-namespaces --older-than 24h
```

Delete only Completed pods older than 7d in one namespace:
```bash
k8s-cleanup run --namespace my-ns --older-than 7d --completed --failed=false --evicted=false --dry-run=false
```

Label selector and audit to file:
```bash
k8s-cleanup run --all-namespaces --label-selector app=myapp --older-than 48h --output json --audit-file audit.ndjson
```

Show version:
```bash
k8s-cleanup version
k8s-cleanup --version
```

Generate shell completion:
```bash
k8s-cleanup completion zsh > /usr/local/share/zsh/site-functions/_k8s-cleanup
```

---

## Usage

```
k8s-cleanup run [flags]

Flags:
  --dry-run                         Simulate without deleting (default true)
  --older-than string               Age threshold (e.g., 30m, 24h, 7d) (default "24h")
  --kind strings                    Resource kinds: pod,job (default [pod,job])
  --namespace string                Target namespace (default "default")
  --all-namespaces                  Process all namespaces
  --exclude-ns strings              Namespaces to exclude (default [kube-system,kube-public])
  --label-selector string           Label selector
  --field-selector string           Field selector
  --completed                       Include Completed/Succeeded (default true)
  --failed                          Include Failed (default true)
  --evicted                         Include Evicted pods (default true)
  --protect string                  Protect resources with this label key[=value] (default "keep=true")
  --concurrency int                 Concurrent deletions (default 10)
  --output string                   Output format: text|json (default "text")
  --audit-file string               Write NDJSON audit events to file
  --exit-nonzero-on-changes         Exit with code 2 if there are candidates (dry-run)
  --log-level string                Log level: trace|debug|info|warn|error (default "info")
```

Global flags:
```
--config string                     Config file (YAML)
--log-level string                  Log level for all commands
```

### JSON Output
```json
[
  {
    "resource":"pod",
    "namespace":"cleanup-test",
    "name":"job-success-abc12",
    "state":"Succeeded",
    "age": 3600000000000,
    "deleted":false,
    "dryRun":true,
    "ts":"2025-09-03T10:00:00Z"
  }
]
```

### Exit Codes
- `0` no candidates / no changes
- `2` changes detected or performed
- `3` errors occurred

---

## Configuration

`k8s-cleanup` reads a YAML config if present (`--config`, `./cleanup.yaml`, `$HOME/cleanup.yaml`, `$HOME/.config/k8s-cleanup/cleanup.yaml`). Flags override config.

Example `cleanup.yaml`:
```yaml
olderThan: 24h
kinds: [pod, job]
allNamespaces: true
excludeNamespaces: [kube-system, kube-public, local-path-storage]
completed: true
failed: true
evicted: true
protectLabel: keep=true
concurrency: 20
output: text
exitNonZeroOnChanges: false
```

Environment:
- `KUBECONFIG` to point to your kubeconfig

---

## RBAC

Minimal RBAC if running in-cluster:
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata: { name: k8s-cleanup }
rules:
- apiGroups: [""]       # core
  resources: ["pods","namespaces"]
  verbs: ["get","list","watch","delete"]
- apiGroups: ["batch"]
  resources: ["jobs","cronjobs"]
  verbs: ["get","list","watch","delete"]
---
apiVersion: v1
kind: ServiceAccount
metadata: { name: k8s-cleanup, namespace: ops }
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata: { name: k8s-cleanup }
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8s-cleanup
subjects:
- kind: ServiceAccount
  name: k8s-cleanup
  namespace: ops
```

---

## Development

Prereqs: Go ≥ 1.22, Docker (for e2e), kind, kubectl.

```bash
make tidy
make test
make build
make e2e-go
```

### Run e2e locally
```bash
go test -v -tags=e2e ./test/e2e
```

### Lint
```bash
golangci-lint run
```

---

## CI/CD

- `ci.yml` runs unit tests on PRs and main, e2e on main, and publishes images `:main` and `:sha-XXXX` to GHCR.
- `release.yml` on tag `vX.Y.Z` publishes multi-arch images `:vX.Y.Z` and `:latest`, attaches binaries to GitHub Release, and pushes the Helm chart to GHCR as OCI.

Release:
```bash
git tag v0.1.0
git push origin v0.1.0
```

---

## Roadmap

- Krew plugin (`kubectl cleanup`)
- More resource kinds (PVCs by finalizer/age)
- TTL policies per namespace via config
- Slack/Webhook notifications

---

## Contributing

PRs and issues are welcome. Please run tests locally before opening a PR.

---

## License

Apache-2.0
