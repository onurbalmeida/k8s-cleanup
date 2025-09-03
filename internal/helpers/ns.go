package helpers

import v1 "k8s.io/api/core/v1"

func FilterNamespaces(nsList *v1.NamespaceList, exclude []string) []string {
	if nsList == nil || len(nsList.Items) == 0 {
		return []string{}
	}
	ex := make(map[string]struct{}, len(exclude))
	for _, n := range exclude {
		ex[n] = struct{}{}
	}
	out := make([]string, 0, len(nsList.Items))
	for _, n := range nsList.Items {
		if _, drop := ex[n.Name]; drop {
			continue
		}
		out = append(out, n.Name)
	}
	return out
}
