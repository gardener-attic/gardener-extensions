{{- define "cloud-provider-config"}}
cloud: AZUREPUBLICCLOUD
resourceGroup: "{{ .Values.resourceGroup }}"
location: "{{ .Values.region }}"
vnetName: "{{ .Values.vnetName }}"
subnetName: "{{ .Values.subnetName }}"
securityGroupName: "{{ .Values.securityGroupName }}"
routeTableName: "{{ .Values.routeTableName }}"
primaryAvailabilitySetName: "{{ .Values.availabilitySetName }}"
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
