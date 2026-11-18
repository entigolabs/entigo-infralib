{{/*
The CAs the AWS Private CA issuer may issue from, as a JSON list: global.aws.privateCA.arns, else the CA of the default
private domain of the route53 module.
*/}}
{{- define "cert-manager.privateCAArns" -}}
{{- $arns := .Values.global.aws.privateCA.arns -}}
{{- if and (not $arns) .Values.global.aws.intCertificateAuthorityArn -}}
{{- $arns = list .Values.global.aws.intCertificateAuthorityArn -}}
{{- end -}}
{{- $arns | default list | toJson -}}
{{- end -}}

{{/*
spec.allowed of the CertificateRequestPolicy, as JSON: global.aws.privateCA.policy.allowed, else the names of the
default private domain of the route53 module with the usages of the default end entity template.
*/}}
{{- define "cert-manager.privateCAPolicyAllowed" -}}
{{- $allowed := .Values.global.aws.privateCA.policy.allowed -}}
{{- $domain := .Values.global.aws.intCertZoneName -}}
{{- if and (not $allowed) $domain -}}
{{- $names := list (printf "*.%s" $domain) $domain -}}
{{- $allowed = dict "commonName" (dict "validations" (list (dict "rule" (printf "self == %s || self.endsWith(%s)" ($domain | quote) (printf ".%s" $domain | quote)) "message" (printf "must be %s or a name under it" $domain)))) "dnsNames" (dict "values" $names) "usages" (list "server auth" "client auth") -}}
{{- end -}}
{{- $allowed | default dict | toJson -}}
{{- end -}}
