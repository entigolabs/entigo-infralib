{{/*
The "@sha256:..." suffix of a provider package, from the generated values-images.yaml. Empty when
the package has no digest there, or the file is not among the value files: the package is then
installed by its tag alone.
Use on the line after the reference: {{- include "crossplane-google.digest" (list $ "<package>:<version>") }}
*/}}
{{- define "crossplane-google.digest" -}}
{{- with index (index . 0).Values.imageDigests (index . 1) }}@{{ . }}{{ end -}}
{{- end }}
