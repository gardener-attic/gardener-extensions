{{- define "cloud-provider-config"}}
cloud: AZUREPUBLICCLOUD
location: "{{ .Values.region }}"
resourceGroup: "{{ .Values.resourceGroup }}"
routeTableName: "{{ .Values.routeTableName }}"
securityGroupName: "{{ .Values.securityGroupName }}"
loadBalancerSku: "{{ .Values.loadBalancerSku }}"
subnetName: "{{ .Values.subnetName }}"
vnetName: "{{ .Values.vnetName }}"
{{- if hasKey .Values "availabilitySetName" }}
primaryAvailabilitySetName: "{{ .Values.availabilitySetName }}"
{{- end }}
cloudProviderBackoff: true
cloudProviderBackoffRetries: 6
cloudProviderBackoffExponent: 1.5
cloudProviderBackoffDuration: 5
cloudProviderBackoffJitter: 1.0
cloudProviderRateLimit: true
cloudProviderRateLimitQPS: 10.0
cloudProviderRateLimitBucket: 100
cloudProviderRateLimitQPSWrite: 10.0
cloudProviderRateLimitBucketWrite: 100
{{- if semverCompare ">= 1.14" .Values.kubernetesVersion }}
cloudProviderBackoffMode: v2
{{- end }}
{{- end }}
