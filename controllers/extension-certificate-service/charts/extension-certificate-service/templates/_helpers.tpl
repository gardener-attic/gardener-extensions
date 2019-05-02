{{- define "certconfig" -}}
---
apiVersion: certificate-service.extensions.config.gardener.cloud/v1alpha1
kind: Configuration
spec:
  lifecycleSync: {{ required ".Values.certificateConfig.lifecycleSync is required" .Values.certificateConfig.lifecycleSync }}
  serviceSync: {{ required ".Values.certificateConfig.serviceSync is required" .Values.certificateConfig.serviceSync }}
  issuerName: {{ required ".Values.certificateConfig.issuerName is required" .Values.certificateConfig.issuerName }}
  namespaceRef: {{ .Release.Namespace }}
  resourceNamespace: {{ required ".Values.certificateConfig.issuerName is required" .Values.certificateConfig.resourceNamespace }}
  acme:
    email: {{ required ".Values.certificateConfig.acme.email is required" .Values.certificateConfig.acme.email }}
    server: {{ required ".Values.certificateConfig.acme.server is required" .Values.certificateConfig.acme.server }}
    {{- if .Values.certificateConfig.acme.privateKey }}
    privateKey: | 
{{ .Values.certificateConfig.acme.privateKey | trim | indent 6 }}
    {{- end }}
  providers:
    {{- if .Values.certificateConfig.providers.clouddns }}
    clouddns:
    {{- range .Values.certificateConfig.providers.clouddns }}
    - name: {{ required ".name is required" .name }}
      {{- if not .domains }}
      {{ required ".domains is required" .domains }}
      {{- end }}
      domains: 
      {{- range .domains }}
      - {{ . }}
      {{- end }}
      project: {{ required ".project is required" .project }}
      serviceAccount: |
{{required ".serviceAccount is required" .serviceAccount | trim | indent 8 }}
    {{- end }}
    {{- end }}
    {{- if .Values.certificateConfig.providers.route53 }}
    route53:
    {{- range .Values.certificateConfig.providers.route53 }}
    - name: {{ required ".name is required" .name }}
      {{- if not .domains }}
      {{ required ".domains is required" .domains }}
      {{- end }}
      domains: 
      {{- range .domains }}
      - {{ . }}
      {{- end }}
      region: {{ required ".region is required" .region }}
      accessKeyID: {{ required ".accessKeyID is required" .accessKeyID }}
      secretAccessKey: {{ required ".secretAccessKey is required" .secretAccessKey }}
    {{- end }}
    {{- end }}
{{- end }}

{{-  define "image" -}}
  {{- if hasPrefix "sha256:" .Values.image.tag }}
  {{- printf "%s@%s" .Values.image.repository .Values.image.tag }}
  {{- else }}
  {{- printf "%s:%s" .Values.image.repository .Values.image.tag }}
  {{- end }}
{{- end }}
