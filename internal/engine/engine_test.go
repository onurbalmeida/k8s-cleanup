package engine

import (
	"context"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func ns(name string) *corev1.Namespace {
	return &corev1.Namespace{ObjectMeta: meta.ObjectMeta{Name: name}}
}

func pod(ns, name string, phase corev1.PodPhase, reason string, started time.Time, labels map[string]string) *corev1.Pod {
	p := &corev1.Pod{
		ObjectMeta: meta.ObjectMeta{Name: name, Namespace: ns, Labels: labels},
		Status:     corev1.PodStatus{Phase: phase, Reason: reason},
	}
	t := meta.NewTime(started)
	p.Status.StartTime = &t
	return p
}

func job(ns, name string, state string, ref time.Time, labels map[string]string) *batchv1.Job {
	j := &batchv1.Job{
		ObjectMeta: meta.ObjectMeta{Name: name, Namespace: ns, Labels: labels},
	}
	switch state {
	case "Succeeded":
		j.Status.Conditions = append(j.Status.Conditions, batchv1.JobCondition{
			Type: batchv1.JobComplete, Status: corev1.ConditionTrue,
		})
		t := meta.NewTime(ref)
		j.Status.CompletionTime = &t
	case "Failed":
		j.Status.Conditions = append(j.Status.Conditions, batchv1.JobCondition{
			Type: batchv1.JobFailed, Status: corev1.ConditionTrue,
		})
		t := meta.NewTime(ref)
		j.Status.CompletionTime = &t
	default:
	}
	return j
}

func Test_FindCandidates_Pods(t *testing.T) {
	c := fake.NewSimpleClientset(
		ns("test"),
		pod("test", "p-ok", corev1.PodSucceeded, "", time.Now().Add(-2*time.Hour), nil),
		pod("test", "p-fail", corev1.PodFailed, "", time.Now().Add(-3*time.Hour), nil),
		pod("test", "p-ev", corev1.PodRunning, "Evicted", time.Now().Add(-4*time.Hour), nil),
		pod("test", "p-new", corev1.PodSucceeded, "", time.Now().Add(-10*time.Minute), nil),
	)
	cfg := Config{
		OlderThan:        time.Hour,
		Kinds:            []string{"pod"},
		Namespaces:       []string{"test"},
		IncludeCompleted: true,
		IncludeFailed:    true,
		IncludeEvicted:   true,
		ProtectKey:       "",
		ProtectVal:       "",
	}
	e := New(c, cfg)
	list, err := e.FindCandidates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 3 {
		t.Fatalf("want 3 candidates, got %d", len(list))
	}
}

func Test_FindCandidates_ProtectLabel(t *testing.T) {
	c := fake.NewSimpleClientset(
		ns("test"),
		pod("test", "p1", corev1.PodSucceeded, "", time.Now().Add(-2*time.Hour), map[string]string{"keep": "true"}),
		pod("test", "p2", corev1.PodSucceeded, "", time.Now().Add(-2*time.Hour), nil),
	)
	cfg := Config{
		OlderThan:        time.Hour,
		Kinds:            []string{"pod"},
		Namespaces:       []string{"test"},
		IncludeCompleted: true,
		ProtectKey:       "keep",
		ProtectVal:       "true",
	}
	e := New(c, cfg)
	list, err := e.FindCandidates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != "p2" {
		t.Fatalf("protect label not respected: %+v", list)
	}
}

func Test_FindCandidates_Jobs(t *testing.T) {
	c := fake.NewSimpleClientset(
		ns("test"),
		job("test", "j-ok", "Succeeded", time.Now().Add(-25*time.Hour), nil),
		job("test", "j-fail", "Failed", time.Now().Add(-26*time.Hour), nil),
		job("test", "j-new", "Succeeded", time.Now().Add(-10*time.Minute), nil),
	)
	cfg := Config{
		OlderThan:        24 * time.Hour,
		Kinds:            []string{"job"},
		Namespaces:       []string{"test"},
		IncludeCompleted: true,
		IncludeFailed:    true,
	}
	e := New(c, cfg)
	list, err := e.FindCandidates(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2 candidates, got %d", len(list))
	}
}

func Test_Delete_Pod_And_Job(t *testing.T) {
	c := fake.NewSimpleClientset(
		ns("test"),
		pod("test", "p-del", corev1.PodFailed, "", time.Now().Add(-2*time.Hour), nil),
		job("test", "j-del", "Failed", time.Now().Add(-2*time.Hour), nil),
	)
	e := New(c, Config{})
	if err := e.Delete(context.Background(), Candidate{Kind: "pod", Namespace: "test", Name: "p-del"}); err != nil {
		t.Fatal(err)
	}
	if err := e.Delete(context.Background(), Candidate{Kind: "job", Namespace: "test", Name: "j-del"}); err != nil {
		t.Fatal(err)
	}
}
