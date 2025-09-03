package engine

import (
	"context"
	"time"

	"github.com/onurbalmeida/k8s-cleanup/internal/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Config struct {
	OlderThan         time.Duration
	Kinds             []string
	AllNamespaces     bool
	Namespaces        []string
	ExcludeNamespaces []string
	LabelSelector     string
	FieldSelector     string
	IncludeCompleted  bool
	IncludeFailed     bool
	IncludeEvicted    bool
	ProtectKey        string
	ProtectVal        string
}

type Candidate struct {
	Kind      string
	Namespace string
	Name      string
	State     string
	Age       time.Duration
}

type Engine struct {
	kube kubernetes.Interface
	cfg  Config
}

func New(kube kubernetes.Interface, cfg Config) *Engine {
	return &Engine{kube: kube, cfg: cfg}
}

func (e *Engine) FindCandidates(ctx context.Context) ([]Candidate, error) {
	namespaces, err := e.resolveNamespaces(ctx)
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().Add(-e.cfg.OlderThan)
	var out []Candidate

	if helpers.HasKind(e.cfg.Kinds, "pod") {
		for _, ns := range namespaces {
			list, err := e.kube.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
				LabelSelector: e.cfg.LabelSelector,
				FieldSelector: e.cfg.FieldSelector,
			})
			if err != nil {
				return nil, err
			}
			for i := range list.Items {
				p := list.Items[i]
				if e.protected(p.Labels) {
					continue
				}
				state := helpers.PodState(&p)
				if !e.stateIncluded(state) {
					continue
				}
				ts := p.CreationTimestamp.Time
				if p.Status.StartTime != nil {
					ts = p.Status.StartTime.Time
				}
				if ts.After(cutoff) {
					continue
				}
				out = append(out, Candidate{
					Kind:      "pod",
					Namespace: ns,
					Name:      p.Name,
					State:     state,
					Age:       time.Since(ts),
				})
			}
		}
	}

	if helpers.HasKind(e.cfg.Kinds, "job") {
		for _, ns := range namespaces {
			list, err := e.kube.BatchV1().Jobs(ns).List(ctx, metav1.ListOptions{
				LabelSelector: e.cfg.LabelSelector,
				FieldSelector: e.cfg.FieldSelector,
			})
			if err != nil {
				return nil, err
			}
			for i := range list.Items {
				j := list.Items[i]
				if e.protected(j.Labels) {
					continue
				}
				state := helpers.JobState(&j)
				if !e.stateIncluded(state) {
					continue
				}
				ref := helpers.JobRefTime(&j)
				if ref.After(cutoff) {
					continue
				}
				out = append(out, Candidate{
					Kind:      "job",
					Namespace: ns,
					Name:      j.Name,
					State:     state,
					Age:       time.Since(ref),
				})
			}
		}
	}

	return out, nil
}

func (e *Engine) Delete(ctx context.Context, c Candidate) error {
	pp := metav1.DeletePropagationForeground
	switch c.Kind {
	case "pod":
		return e.kube.CoreV1().Pods(c.Namespace).Delete(ctx, c.Name, metav1.DeleteOptions{PropagationPolicy: &pp})
	case "job":
		return e.kube.BatchV1().Jobs(c.Namespace).Delete(ctx, c.Name, metav1.DeleteOptions{PropagationPolicy: &pp})
	default:
		return nil
	}
}

func (e *Engine) resolveNamespaces(ctx context.Context) ([]string, error) {
	if e.cfg.AllNamespaces {
		nsList, err := e.kube.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, err
		}
		return helpers.FilterNamespaces(nsList, e.cfg.ExcludeNamespaces), nil
	}
	if len(e.cfg.Namespaces) == 0 {
		return []string{"default"}, nil
	}
	return e.cfg.Namespaces, nil
}

func (e *Engine) stateIncluded(state string) bool {
	s := lower(state)
	switch s {
	case "succeeded":
		return e.cfg.IncludeCompleted
	case "failed":
		return e.cfg.IncludeFailed
	case "evicted":
		return e.cfg.IncludeEvicted
	default:
		return false
	}
}

func (e *Engine) protected(labels map[string]string) bool {
	if e.cfg.ProtectKey == "" {
		return false
	}
	if e.cfg.ProtectVal == "" {
		_, ok := labels[e.cfg.ProtectKey]
		return ok
	}
	return labels[e.cfg.ProtectKey] == e.cfg.ProtectVal
}

func lower(s string) string {
	b := []byte(s)
	for i := range b {
		if 'A' <= b[i] && b[i] <= 'Z' {
			b[i] = b[i] + 32
		}
	}
	return string(b)
}
