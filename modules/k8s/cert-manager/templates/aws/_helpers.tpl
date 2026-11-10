{{/*
"true" when the ACME ClusterIssuer and its Route53 identity are created on AWS, empty otherwise
*/}}
{{- define "cert-manager.awsIssuer" -}}
{{- if and (eq .Values.global.cloudProvider "aws") .Values.clusterIssuer.create }}true{{- end }}
{{- end -}}
