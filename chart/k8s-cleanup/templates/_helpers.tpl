{{- define "k8s-cleanup.name" -}}
{{- .Chart.Name -}}
{{- end -}}

{{- define "k8s-cleanup.fullname" -}}
{{- printf "%s-%s" .Release.Name (include "k8s-cleanup.name" .) | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "k8s-cleanup.labels" -}}
app.kubernetes.io/name: {{ include "k8s-cleanup.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
app.kubernetes.io/managed-by: Helm
{{- end -}}
