{{- define "database-controller.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "database-controller.version" -}}
{{- if .Values.image.tag -}}
{{ .Values.image.tag }}
{{- else -}}
{{ .Chart.Version }}
{{- end -}}
{{- end -}}
