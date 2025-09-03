package helpers

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestPodState(t *testing.T) {
	p := corev1.Pod{}
	p.Status.Phase = corev1.PodSucceeded
	if s := PodState(&p); s != "Succeeded" {
		t.Fatal(s)
	}
	p.Status.Reason = "Evicted"
	if s := PodState(&p); s != "Evicted" {
		t.Fatal(s)
	}
}
