package helpers

import "strings"

func normalizeKind(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.TrimSuffix(s, "s")
}

func HasKind(kinds []string, k string) bool {
	nk := normalizeKind(k)
	for _, x := range kinds {
		if normalizeKind(x) == nk {
			return true
		}
	}
	return false
}

func ParseKV(s string) (string, string) {
	if s == "" {
		return "", ""
	}
	parts := strings.SplitN(s, "=", 2)
	if len(parts) == 1 {
		return strings.TrimSpace(parts[0]), ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
