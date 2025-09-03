//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type cleanupRecord struct {
	Resource  string `json:"resource"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	State     string `json:"state"`
	Deleted   bool   `json:"deleted"`
	DryRun    bool   `json:"dryRun"`
}

type k8sList[T any] struct {
	Items []T `json:"items"`
}

func lookPath(t *testing.T, name string) string {
	t.Helper()
	p, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("%s not found in PATH; skipping e2e", name)
	}
	return p
}

func run(ctx context.Context, dir, name string, args []string, stdin string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestEndToEnd_Kind(t *testing.T) {
	lookPath(t, "docker")
	lookPath(t, "kind")
	lookPath(t, "kubectl")

	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(root, "bin", "k8s-cleanup-e2e")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if out, err := run(ctx, root, "go", []string{"build", "-o", bin, "."}, ""); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	cluster := fmt.Sprintf("e2e-%d", time.Now().UnixNano())
	if out, err := run(ctx, root, "kind", []string{"create", "cluster", "--name", cluster}, ""); err != nil {
		t.Fatalf("kind create failed: %v\n%s", err, out)
	}
	defer func() {
		_, _ = run(context.Background(), root, "kind", []string{"delete", "cluster", "--name", cluster}, "")
	}()

	if _, err := run(ctx, root, "kubectl", []string{"config", "use-context", "kind-" + cluster}, ""); err != nil {
		t.Fatalf("kubectl use-context failed: %v", err)
	}

	ns := "cleanup-test"
	_, _ = run(ctx, root, "kubectl", []string{"create", "ns", ns}, "")

	jobsYAML := `
apiVersion: batch/v1
kind: Job
metadata:
  name: job-success
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
metadata:
  name: job-fail
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: c
        image: busybox
        command: ["sh","-c","exit 1"]
  backoffLimit: 0
`
	if out, err := run(ctx, root, "kubectl", []string{"-n", ns, "apply", "-f", "-"}, jobsYAML); err != nil {
		t.Fatalf("apply jobs failed: %v\n%s", err, out)
	}

	_, _ = run(ctx, root, "kubectl", []string{"-n", ns, "wait", "--for=condition=complete", "job/job-success", "--timeout=90s"}, "")
	time.Sleep(3 * time.Second)

	out, err := run(ctx, root, bin, []string{
		"run",
		"--namespace", ns,
		"--older-than", "1s",
		"--output", "json",
		"--log-level", "error",
	}, "")
	if err != nil {
		t.Fatalf("dry-run failed: %v\n%s", err, out)
	}
	var results []cleanupRecord
	if err := json.Unmarshal([]byte(out), &results); err != nil {
		t.Fatalf("json parse failed: %v\n%s", err, out)
	}
	if len(results) < 2 {
		t.Fatalf("expected >=2 candidates, got %d\n%s", len(results), out)
	}

	out, err = run(ctx, root, bin, []string{
		"run",
		"--namespace", ns,
		"--older-than", "1s",
		"--dry-run=false",
	}, "")
	if err != nil {
		if ee, ok := err.(*exec.ExitError); !ok || ee.ExitCode() != 2 {
			t.Fatalf("real run failed: %v\n%s", err, out)
		}
	}

	jout, err := run(ctx, root, "kubectl", []string{"-n", ns, "get", "jobs", "-o", "json"}, "")
	if err != nil {
		t.Fatalf("kubectl get jobs failed: %v\n%s", err, jout)
	}
	var jobList k8sList[struct{}]
	if err := json.Unmarshal([]byte(jout), &jobList); err != nil {
		t.Fatalf("jobs json parse: %v\n%s", err, jout)
	}
	if len(jobList.Items) != 0 {
		t.Fatalf("expected 0 jobs after cleanup, got %d", len(jobList.Items))
	}
}
