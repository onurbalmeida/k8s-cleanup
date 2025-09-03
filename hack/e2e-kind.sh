#!/usr/bin/env bash
set -euo pipefail

CLUSTER=${CLUSTER:-"dev-cleanup"}
NS="cleanup-test"

echo ">> ensure kind cluster"
kind get clusters | grep -q "^${CLUSTER}$" || kind create cluster --name "${CLUSTER}"
kubectl config use-context "kind-${CLUSTER}"

echo ">> namespace"
kubectl get ns "${NS}" >/dev/null 2>&1 || kubectl create ns "${NS}"

echo ">> create jobs"
cat <<'YAML' | kubectl -n "${NS}" apply -f -
apiVersion: batch/v1
kind: Job
metadata: { name: job-success }
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: c
        image: busybox
        command: ["sh","-c","echo ok && sleep 1"]
  backoffLimit: 0
---
apiVersion: batch/v1
kind: Job
metadata: { name: job-fail }
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: c
        image: busybox
        command: ["sh","-c","exit 1"]
  backoffLimit: 0
YAML

echo ">> wait a bit"
sleep 5

echo ">> dry-run candidates"
go run . run --namespace "${NS}" --older-than 1s --output json

echo ">> real delete"
go run . run --namespace "${NS}" --older-than 1s --dry-run=false

echo ">> pods after"
kubectl -n "${NS}" get pods || true

echo ">> done"
