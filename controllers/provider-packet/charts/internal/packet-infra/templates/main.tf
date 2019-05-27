provider "packet" {
  auth_token = "${var.PACKET_API_KEY}"
}

// Deploy a new ssh key
resource "packet_project_ssh_key" "publickey" {
  name = "{{ required "clusterName is required" .Values.clusterName }}-ssh-publickey"
  public_key = "{{ required "sshPublicKey is required" .Values.sshPublicKey }}"
  project_id = "{{ required "packet.projectID is required" .Values.packet.projectID }}"
}

//=====================================================================
//= Output variables
//=====================================================================

output "{{ .Values.outputKeys.sshKeyID }}" {
  value = "${packet_project_ssh_key.publickey.id}"
}
