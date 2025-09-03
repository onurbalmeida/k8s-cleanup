package helpers

import (
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

func PodState(p *corev1.Pod) string {
	if strings.EqualFold(p.Status.Reason, "Evicted") {
		return "Evicted"
	}
	return string(p.Status.Phase)
}

func JobState(j *batchv1.Job) string {
	for _, c := range j.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			return "Succeeded"
		}
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return "Failed"
		}
	}
	return "Active"
}

func JobRefTime(j *batchv1.Job) time.Time {
	if j.Status.CompletionTime != nil {
		return j.Status.CompletionTime.Time
	}
	return j.CreationTimestamp.Time
}
