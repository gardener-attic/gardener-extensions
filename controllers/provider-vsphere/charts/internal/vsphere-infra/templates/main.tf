// nsxt with ip pools, as long as pull request is not accepted, use https://github.com/MartinWeindel/terraform-provider-nsxt/releases/download/v1.1.x-alpha1/terraform-provider-nsxt_linux_amd64
provider "nsxt" {
  host                     = "{{ required "nsxt.host is required" .Values.nsxt.host }}"
  username                 = "${var.USER_NAME}"
  password                 = "${var.PASSWORD}"
  allow_unverified_ssl     = {{ required "nsxt.insecure is required" .Values.nsxt.insecure }}
  max_retries              = 10
  retry_min_delay          = 500
  retry_max_delay          = 5000
  retry_on_status_codes    = [429]
  tolerate_partial_success = true
}

# Inputs
variable "nsx_full_cluster_name" {
    default = "{{ .Values.nsxt.namePrefix }}_{{ .Values.clusterName }}"
}
variable "nsx_tag_scope" {
    default = "nameprefix"
}
variable "nsx_tag" {
    default = "{{ required "nsxt.namePrefix is required" .Values.nsxt.namePrefix }}"
}
variable "nsx_tag_shoot" {
    default = "{{ required "clusterName is required" .Values.clusterName }}"
}
variable "nsx_t1_router_name" {
    default = "{{ .Values.nsxt.namePrefix }}_{{ .Values.clusterName }}"
}
variable "nsx_networks_worker" {
    default = "{{ required "networks.worker is required" .Values.networks.worker }}"
}
variable "nsx_networks_worker_suffix" {
    default = "{{ regexFind "/[0-9]+" .Values.networks.worker }}"
}
variable "nsx_snat_destination_network" {
    default = "0.0.0.0/0"
}

# Prerequisites
data "nsxt_transport_zone" "cluster" {
    display_name = "{{ required "vsphere.nsxt.transportZone is required" .Values.nsxt.transportZone }}"
}

data "nsxt_logical_tier0_router" "cluster" {
    display_name = "{{ required "vsphere.nsxt.logicalTier0Router is required" .Values.nsxt.logicalTier0Router }}"
}

data "nsxt_edge_cluster" "cluster" {
    display_name = "{{ required "vsphere.nsxt.edgeCluster is required" .Values.nsxt.edgeCluster }}"
}

data "nsxt_ip_pool" "snat_pool" {
    display_name = "{{ required "vsphere.nsxt.snatIpPool is required" .Values.nsxt.snatIpPool }}"
}

#################################
# create logical router & switch
#################################

resource "nsxt_logical_switch" "switch" {
  admin_state = "UP"
  description = "logical switch for gardener cluster"
  display_name = "${var.nsx_full_cluster_name}"
  transport_zone_id = "${data.nsxt_transport_zone.cluster.id}"
  replication_mode = "MTEP"

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}

resource "nsxt_logical_router_link_port_on_tier0" "external" {
  description       = "downlink for tier0 router to tier1 router ${var.nsx_t1_router_name}"
  display_name      = "${var.nsx_t1_router_name}-downlink"
  logical_router_id = "${data.nsxt_logical_tier0_router.cluster.id}"

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}

resource "nsxt_logical_tier1_router" "router" {
  description                 = "tier1 router for gardener cluster"
  display_name                = "${var.nsx_t1_router_name}"
  failover_mode               = "PREEMPTIVE"
  edge_cluster_id             = "${data.nsxt_edge_cluster.cluster.id}"
  enable_router_advertisement = true
  advertise_connected_routes  = false
  advertise_static_routes     = true
  advertise_nat_routes        = true
  advertise_lb_vip_routes     = true
  advertise_lb_snat_ip_routes = true

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}

resource "nsxt_logical_router_link_port_on_tier1" "router" {
  description                   = "uplink for tier1 router ${var.nsx_t1_router_name} to tier0 router"
  display_name                  = "${var.nsx_t1_router_name}-uplink"
  logical_router_id             = "${nsxt_logical_tier1_router.router.id}"
  linked_logical_router_port_id = "${nsxt_logical_router_link_port_on_tier0.external.id}"

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}

# Create a switchport on our logical switch
resource "nsxt_logical_port" "switch" {
  admin_state       = "UP"
  description       = "Logical port for cluster switch"
  display_name      = "${var.nsx_full_cluster_name}_LP1"
  logical_switch_id = "${nsxt_logical_switch.switch.id}"

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}

# Create downlink port on the T1 router and connect it to the switchport we created earlier
resource "nsxt_logical_router_downlink_port" "downlink_port" {
  description                   = "Downlink port for cluster switch"
  display_name                  = "${var.nsx_full_cluster_name}_DP1"
  logical_router_id             = "${nsxt_logical_tier1_router.router.id}"
  linked_logical_switch_port_id = "${nsxt_logical_port.switch.id}"
  ip_address                    = "${cidrhost(var.nsx_networks_worker, 1)}${var.nsx_networks_worker_suffix}"

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}

# IP address of all nodes for SNAT
resource "nsxt_ip_pool_allocation_ip_address" "snat" {
  ip_pool_id = "${data.nsxt_ip_pool.snat_pool.id}"
}

resource "nsxt_nat_rule" "cluster-snat" {
  logical_router_id         = "${nsxt_logical_tier1_router.router.id}"
  description               = "snat rule for subnet ${var.nsx_tag_shoot}"
  display_name              = "${var.nsx_full_cluster_name}_NR"
  action                    = "SNAT"
  enabled                   = true
  logging                   = true
  nat_pass                  = false
  translated_network        = "${nsxt_ip_pool_allocation_ip_address.snat.allocation_id}/32"
  match_destination_network = "${var.nsx_snat_destination_network}"
  match_source_network      = "${var.nsx_networks_worker}"

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}

# install a DHCP server
resource "nsxt_dhcp_server_profile" "profile" {
  display_name     = "${var.nsx_full_cluster_name}"
  description      = "dhcp server profile of ${var.nsx_full_cluster_name}"
  edge_cluster_id = "${data.nsxt_edge_cluster.cluster.id}"

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}

resource "nsxt_logical_dhcp_server" "dhcpserver" {
  display_name     = "${var.nsx_full_cluster_name}"
  description      = "logical dhcp server of ${var.nsx_full_cluster_name}"
  dhcp_profile_id  = "${nsxt_dhcp_server_profile.profile.id}"
  dhcp_server_ip   = "${cidrhost(var.nsx_networks_worker, 2)}${var.nsx_networks_worker_suffix}"
  gateway_ip       = "${cidrhost(var.nsx_networks_worker, 1)}"

  {{- if .Values.nsxt.dnsServers }}
  dns_name_servers = [{{- include "vsphere-infra.dnsServers" . | trimSuffix ", " }}]
  {{- else }}
  dns_name_servers = []
  {{- end }}

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}

resource "nsxt_logical_dhcp_port" "dhcpserver" {
  admin_state       = "UP"
  description       = "LP1 for dhcp server of ${var.nsx_full_cluster_name}"
  display_name      = "${var.nsx_full_cluster_name}_LP_DHCP"
  logical_switch_id = "${nsxt_logical_switch.switch.id}"
  dhcp_server_id    = "${nsxt_logical_dhcp_server.dhcpserver.id}"

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}

resource "nsxt_dhcp_server_ip_pool" "dhcp_pool" {
  display_name           = "${var.nsx_full_cluster_name}"
  description            = "dhcp ip pool for ${var.nsx_full_cluster_name}"
  logical_dhcp_server_id = "${nsxt_logical_dhcp_server.dhcpserver.id}"
  gateway_ip             = "${nsxt_logical_dhcp_server.dhcpserver.gateway_ip}"
  lease_time             = 7200
  error_threshold        = 98
  warning_threshold      = 70

  ip_range {
    start = "${cidrhost(var.nsx_networks_worker, 10)}"
    end   = "${cidrhost(var.nsx_networks_worker, -1)}"
  }

  #dhcp_generic_option {
  #  code   = "119" # 119 = domain search list
  #  values = ["my.domain.com"]
  #}

  tag {
    scope = "${var.nsx_tag_scope}"
    tag = "${var.nsx_tag}"
  }
  tag {
    scope = "shoot"
    tag   = "${var.nsx_tag_shoot}"
  }
}


//=====================================================================
//= Output variables
//=====================================================================

output "network_name" {
  value = "${var.nsx_full_cluster_name}"
}

output "logical_router_id" {
  value = "${nsxt_logical_tier1_router.router.id}"
}

output "logical_switch_id" {
  value = "${nsxt_logical_switch.switch.id}"
}

