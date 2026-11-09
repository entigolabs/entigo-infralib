{{- define "crossplane-oracle.providers" -}}
{{- $all := concat .Values.global.requiredProviders .Values.global.extraProviders }}
{{- if .Values.global.createObservers }}
{{- $all = concat $all .Values.global.observerProviders }}
{{- end }}
{{- $unique := list }}
{{- range $p := $all }}
{{- if not (has $p $unique) }}
{{- $unique = append $unique $p }}
{{- end }}
{{- end }}
{{- toYaml $unique }}
{{- end }}
