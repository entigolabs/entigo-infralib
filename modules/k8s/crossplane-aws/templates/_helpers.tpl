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

{{/*
The "@sha256:..." suffix of a provider package, from the generated values-images.yaml. Empty when
the package has no digest there, or the file is not among the value files: the package is then
installed by its tag alone.
Use on the line after the reference: {{- include "crossplane-aws.digest" (list $ "<package>:<version>") }}
*/}}
{{- define "crossplane-aws.digest" -}}
{{- with index (index . 0).Values.imageDigests (index . 1) }}@{{ . }}{{ end -}}
{{- end }}
