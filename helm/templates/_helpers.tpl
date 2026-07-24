{{- define "tsa.name" -}}tyk-sre-assignment{{- end -}}
{{- define "tsa.labels" -}}
app.kubernetes.io/name: {{ include "tsa.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}
{{- define "tsa.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{ .Values.image.repository }}:{{ $tag }}
{{- end -}}
