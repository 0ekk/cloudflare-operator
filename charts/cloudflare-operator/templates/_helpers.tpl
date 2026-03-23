{{/* vim: set filetype=mustache: */}}
{{- define "cloudflare-operator.fullname" -}}
{{- "cloudflare-operator" -}}
{{- end -}}

{{- define "cloudflare-operator.image" -}}
{{- $tag := .Values.image.tag | default (printf "v%s" .Chart.Version) -}}
{{- printf "%s:%s" .Values.image.repository $tag -}}
{{- end -}}
