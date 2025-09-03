# k8s-cleanup (Helm Chart)

![Helm](https://img.shields.io/badge/helm-oci-blue)
![Registry](https://img.shields.io/badge/registry-ghcr.io-black)
![License](https://img.shields.io/badge/license-Apache--2.0-green)

Cleanup **old Pods and Jobs** in Kubernetes. Deploy as a **one‑shot Job** or a **scheduled CronJob**. Built in Go for speed, safety, and clear logs.

---

## 🚀 Quick start

> Requirements: **Helm ≥ 3.8** (OCI support), **Kubernetes ≥ 1.24**.

### Run once (Job)
```bash
helm install cleanup oci://ghcr.io/onurbalmeida/charts/k8s-cleanup   --version 1.2.3   -n ops --create-namespace   --set schedule=""   --set image.repository=ghcr.io/onurbalmeida/k8s-cleanup   --set image.tag=v1.2.3   --set args.allNamespaces=false   --set args.namespace=cleanup-test   --set args.olderThan=12h   --set args.dryRun=false
```

### Run on a schedule (CronJob, daily at 02:00)
```bash
helm install cleanup oci://ghcr.io/onurbalmeida/charts/k8s-cleanup   --version 1.2.3   -n ops --create-namespace   --set schedule="0 2 * * *"   --set image.repository=ghcr.io/onurbalmeida/k8s-cleanup   --set image.tag=v1.2.3   --set args.allNamespaces=true   --set args.olderThan=24h   --set args.dryRun=false
```

Inspect the chart directly from the registry:
```bash
helm show chart  oci://ghcr.io/onurbalmeida/charts/k8s-cleanup --version 1.2.3
helm show values oci://ghcr.io/onurbalmeida/charts/k8s-cleanup --version 1.2.3
```

> 💡 **Note:** In the examples below we use `--version 1.2.3` and `image.tag=v1.2.3` only as placeholders.
> Always replace them with the actual version you want to use to ensure reproducible results.

---

## ✨ Features

- Deletes **Pods** and **Jobs** older than a given threshold (`--older-than`).
- Filters by state (**Succeeded**, **Failed**, **Evicted**).
- Works **cluster‑wide** or **per namespace**.
- **Protection selector** (e.g., keep anything labeled `keep=true`).
- Concurrency‑limited deletions, **dry‑run** mode, optional **JSON** output/audit.

---

## 🔧 Common recipes

Single namespace cleanup:
```bash
helm upgrade --install cleanup oci://ghcr.io/onurbalmeida/charts/k8s-cleanup   --version 1.2.3 -n ops   --set schedule=""   --set image.tag=v1.2.3   --set args.allNamespaces=false   --set args.namespace=cleanup-test   --set args.olderThan=12h   --set args.failed=true --set args.completed=true --set args.evicted=true   --set args.dryRun=false
```

Cluster‑wide but skip system + Helm namespace (`ops`):
```bash
helm upgrade --install cleanup oci://ghcr.io/onurbalmeida/charts/k8s-cleanup   --version 1.2.3 -n ops   --set schedule="*/30 * * * *"   --set image.tag=v1.2.3   --set args.allNamespaces=true   --set args.excludeNamespaces="{ops,kube-system,kube-public,local-path-storage}"   --set args.olderThan=24h --set args.dryRun=false
```

Only Jobs (skip Pods):
```bash
--set args.kinds="{job}"
```

Protect by label (default keeps `keep=true`):
```bash
--set args.protect="keep=true"
# or
--set args.protect="do-not-delete=true"
```

Dry‑run with JSON output and non‑zero exit if changes:
```bash
--set args.dryRun=true --set args.output=json --set args.exitNonZeroOnChanges=true
```

---

## 🧰 Values reference

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| `schedule` | string | `"0 2 * * *"` | When set, deploy as **CronJob**; empty for **one‑shot Job** |
| `image.repository` | string | `ghcr.io/onurbalmeida/k8s-cleanup` | Container image repository |
| `image.tag` | string | chart `.appVersion` | Container image tag (e.g., `v1.2.3`) |
| `image.pullPolicy` | string | `IfNotPresent` | Image pull policy |
| `args.dryRun` | bool | `true` | Don’t delete, only report |
| `args.olderThan` | string | `"24h"` | Age threshold (`30m`, `12h`, `7d`, …) |
| `args.kinds` | list(string) | `["pod","job"]` | Kinds to target |
| `args.allNamespaces` | bool | `true` | If true, process all namespaces |
| `args.namespace` | string | `""` | Namespace when `allNamespaces=false` |
| `args.excludeNamespaces` | list(string) | `["kube-system","kube-public","local-path-storage"]` | Namespaces to skip (consider adding your Helm ns) |
| `args.labelSelector` | string | `""` | Label selector |
| `args.fieldSelector` | string | `""` | Field selector |
| `args.completed` | bool | `true` | Include **Completed** |
| `args.failed` | bool | `true` | Include **Failed** |
| `args.evicted` | bool | `true` | Include **Evicted** |
| `args.protect` | string | `"keep=true"` | Skip resources matching this selector |
| `args.concurrency` | int | `10` | Max parallel deletions |
| `args.output` | string | `"text"` | Output: `text` or `json` |
| `args.auditFile` | string | `""` | Write JSON audit to file (container FS) |
| `args.exitNonZeroOnChanges` | bool | `false` | Exit code `1` if something would be/was deleted |
| `args.logLevel` | string | `"info"` | `debug`, `info`, `warn`, `error` |
| `args.extra` | list(string) | `[]` | Extra raw CLI args |
| `serviceAccount.create` | bool | `true` | Create a ServiceAccount |
| `serviceAccount.name` | string | `""` | Use existing SA |
| `rbac.create` | bool | `true` | Create ClusterRole/Binding |
| `resources` | object | `{}` | Pod resources |
| `nodeSelector` | object | `{}` | Pod node selector |
| `tolerations` | list | `[]` | Pod tolerations |
| `affinity` | object | `{}` | Pod affinity |

Show defaults straight from registry:
```bash
helm show values oci://ghcr.io/onurbalmeida/charts/k8s-cleanup --version 1.2.3
```

---

## 📝 Operational notes

- **Job vs CronJob**: set `schedule=""` for a one‑shot Job; set a cron expression for a CronJob.
- **Upgrading a Job**: Jobs have an **immutable Pod template**. To change image/args: delete the existing Job or switch to CronJob.
- **Exclude your release namespace**: when running cluster‑wide, add your Helm namespace to `args.excludeNamespaces` to avoid touching its own Job pods.
- **Audit file**: if you use `args.auditFile`, mount a volume to persist it.

---

## 🔍 Troubleshooting

- `ImagePullBackOff`: ensure `image.tag` exists and the image is public on GHCR.
- `cannot patch ... Job ... field is immutable`: delete the Job before upgrade or use a CronJob.
- “Nothing deleted”: confirm `--older-than`, `--dry-run=false`, selectors, and namespaces.

Debug commands:
```bash
helm status cleanup -n ops
helm get values cleanup -n ops -a
helm get manifest cleanup -n ops
kubectl -n ops logs job/cleanup-k8s-cleanup
```

---

## 🔖 Versioning

- **Chart version**: Helm packaging (por exemplo `1.2.3`).
- **AppVersion**: container image tag (por exemplo `v1.2.3`).

⚠️ Always use the actual version you intend to deploy instead of the placeholder.

---

## 📜 License

Apache 2.0. See the repository LICENSE.
