{{/* vim: set filetype=mustache: */}}
{{- define "cloudflare-operator.fullname" -}}
{{- "cloudflare-operator" -}}
{{- end -}}

{{- define "cloudflare-operator.image" -}}
{{- $tag := .Values.image.tag | default .Chart.AppVersion -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end -}}
