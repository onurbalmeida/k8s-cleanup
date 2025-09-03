package helpers

import "testing"

func TestHasKind(t *testing.T) {
	cases := []struct {
		kinds []string
		k     string
		want  bool
	}{
		{[]string{"pod", "job"}, "pod", true},
		{[]string{"pod", "job"}, "pods", true},
		{[]string{"Pod", "Job"}, "job", true},
		{[]string{"pod"}, "job", false},
	}
	for _, c := range cases {
		if got := HasKind(c.kinds, c.k); got != c.want {
			t.Fatalf("HasKind(%v,%q) got %v want %v", c.kinds, c.k, got, c.want)
		}
	}
}

func TestParseKV(t *testing.T) {
	k, v := ParseKV("keep=true")
	if k != "keep" || v != "true" {
		t.Fatalf("got %q=%q", k, v)
	}
	k, v = ParseKV("ttl")
	if k != "ttl" || v != "" {
		t.Fatalf("got %q=%q", k, v)
	}
}
