{{/*
Comma-separated list of every fully qualified hostname this gateway serves, for
external-dns's hostname annotation. Kept in one place so the Gateway's listeners and the
DNS records it publishes can never disagree about the hostname format.
*/}}
{{- define "oracle-gateway.hostnames" -}}
{{- $domain := .Values.global.oracle.domain -}}
{{- $names := list -}}
{{- range .Values.hosts -}}
{{- $names = append $names (printf "%s.%s" . $domain) -}}
{{- end -}}
{{- join "," $names -}}
{{- end -}}
