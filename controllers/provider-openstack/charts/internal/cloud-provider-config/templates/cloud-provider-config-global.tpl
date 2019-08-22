{{- define "cloud-provider-config-global"}}
[Global]
auth-url="{{ .Values.authUrl }}"
{{- if .Values.domainName }}
domain-name="{{ .Values.domainName }}"
{{- end }}
{{- if .Values.domainID }}
domain-id="{{ .Values.domainID }}"
{{- end }}
{{- if .Values.tenantName }}
tenant-name="{{ .Values.tenantName }}"
{{- end }}
{{- if .Values.tenantID }}
tenant-id="{{ .Values.tenantID }}"
{{- end }}
{{- if .Values.userDomainName }}
user-domain-name="{{ .Values.userDomainName }}"
{{- end }}
{{- if .Values.userDomainID }}
user-domain-id="{{ .Values.userDomainID }}"
{{- end }}
username="{{ .Values.username }}"
password="{{ .Values.password }}"
[LoadBalancer]
create-monitor=true
monitor-delay=60s
monitor-timeout=30s
monitor-max-retries=5
lb-version=v2
lb-provider="{{ .Values.lbProvider }}"
floating-network-id="{{ .Values.floatingNetworkID }}"
{{- end }}