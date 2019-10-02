{{- define "vsphere-infra.dnsServers" }}
{{- if .Values.nsxt.dnsServers }}
{{- range .Values.nsxt.dnsServers }}"{{ . }}", {{ end }}
{{- end }}
{{- end -}}
