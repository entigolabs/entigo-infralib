{{/*
The list of upbound-provider-aws-* subpackages to install: requiredProviders, extraProviders and,
when observers are enabled, observerProviders, deduplicated in that order.
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
Resources for one subpackage: providerResources.default deep-merged with providerResources.providers.<name>.
*/}}
{{- define "crossplane-aws.providerResources" -}}
{{- $override := index .Values.providerResources.providers .provider | default dict }}
{{- toYaml (mergeOverwrite (deepCopy .Values.providerResources.default) $override) }}
{{- end }}

{{/*
Name of the DeploymentRuntimeConfig a subpackage uses: its own when it has a providerResources override,
the shared one otherwise.
*/}}
{{- define "crossplane-aws.runtimeConfigName" -}}
{{- if hasKey .Values.providerResources.providers .provider }}
{{- printf "%s-%s" .Release.Name .provider }}
{{- else }}
{{- .Release.Name }}
{{- end }}
{{- end }}

{{/*
A DeploymentRuntimeConfig. Takes a dict: root (the top-level context), name, resources, and
serviceAccount (true renders the serviceAccountTemplate that creates the shared IRSA ServiceAccount).
*/}}
{{- define "crossplane-aws.runtimeConfig" -}}
apiVersion: pkg.crossplane.io/v1beta1
kind: DeploymentRuntimeConfig
metadata:
  name: {{ .name }}
  annotations:
    argocd.argoproj.io/sync-options: SkipDryRunOnMissingResource=true
    argocd.argoproj.io/sync-wave: '1'
    helm.sh/resource-policy: keep
spec:
  deploymentTemplate:
    spec:
      selector: {}
      strategy: {}
      template:
        spec:
          affinity:
            nodeAffinity:
              preferredDuringSchedulingIgnoredDuringExecution:
                - preference:
                    matchExpressions:
                      - key: tools
                        operator: In
                        values:
                          - 'true'
                  weight: 90
          containers:
            - name: package-runtime
              args:
                {{- with .root.Values.global.deploymentRuntimeConfig.args }}
                {{- toYaml . | nindent 16 }}
                {{- end }}
              resources:
                {{- toYaml .resources | nindent 16 }}
              securityContext:
                runAsNonRoot: true
                seccompProfile:
                  type: RuntimeDefault
          securityContext:
            fsGroup: 2000
            runAsNonRoot: true
            seccompProfile:
              type: RuntimeDefault
          serviceAccountName: {{ .root.Release.Name }}
          tolerations:
            - effect: NoSchedule
              key: tools
              operator: Equal
              value: 'true'
  serviceTemplate: {}
{{- if .serviceAccount }}
  serviceAccountTemplate:
    metadata:
      name: {{ .root.Release.Name }}
      annotations:
        eks.amazonaws.com/role-arn: {{ .root.Values.global.aws.role }}
{{- end }}
{{- end }}
