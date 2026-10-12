{{/*
The upbound-provider-aws-* subpackages to install: requiredProviders, extraProviders and, when
observers are enabled, observerProviders, deduplicated in that order.
Use as: include "crossplane-aws.providers" . | fromYamlArray
*/}}
{{- define "crossplane-aws.providers" -}}
{{- $all := concat .Values.global.requiredProviders .Values.global.extraProviders }}
{{- if or .Values.global.createObservers .Values.global.createNamespacedObservers }}
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
