{{/*
The ACME account of the ClusterIssuer: server, registration email and External Account Binding. Each cloud's
templates/<cloud>/clusterIssuer.yaml adds its DNS-01 solvers
*/}}
{{- define "cert-manager.acme" -}}
server: {{ .Values.acme.server }}
{{- with .Values.acme.email }}
email: {{ . }}
{{- end }}
privateKeySecretRef:
  name: {{ .Values.clusterIssuer.name }}-account-key
{{- with .Values.acme.externalAccountBinding }}
{{- if .keyID }}
externalAccountBinding:
  keyID: {{ .keyID }}
  keySecretRef:
    {{- toYaml .keySecretRef | nindent 4 }}
{{- end }}
{{- end }}
{{- end -}}

{{/*
"true" when the ACME ClusterIssuer and its Route53 identity are created on AWS, empty otherwise
*/}}
{{- define "cert-manager.awsIssuer" -}}
{{- if and (eq .Values.global.cloudProvider "aws") .Values.clusterIssuer.create }}true{{- end }}
{{- end -}}
