package aws

import (
	"github.com/gardener/gardener/pkg/apis/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta
	Networks Networks
}

// Networks holds information about the Kubernetes and infrastructure networks.
type Networks struct {
	core.K8SNetworks
	// VPC indicates whether to use an existing VPC or create a new one.
	VPC VPC
	// Internal is a list of private subnets to create (used for internal load balancers).
	Internal []core.CIDR
	// Public is a list of public subnets to create (used for bastion and load balancers).
	Public []core.CIDR
	// Workers is a list of worker subnets (private) to create (used for the VMs).
	Workers []core.CIDR
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureStatus contains information about created infrastructure resources.
type InfrastructureStatus struct {
	metav1.TypeMeta

	// EC2 contains information about the created AWS EC2 resources.
	EC2 EC2
	// IAM contains information about the created AWS IAM resources.
	IAM IAM
	// VPC contains information about the created AWS VPC and some related resources.
	VPC VPC
}

// EC2 contains information about the created AWS EC2 resources.
type EC2 struct {
	// KeyName is the name of the SSH key that has been created.
	KeyName string
}

// IAM contains information about the created AWS IAM resources.
type IAM struct {
	// InstanceProfiles is a list of AWS IAM instance profiles.
	InstanceProfiles []InstanceProfile
	// Roles is a list of AWS IAM roles.
	Roles []Role
}

// VPC contains information about the created AWS VPC and some related resources.
type VPC struct {
	// ID is the VPC id.
	ID string
	// CIDR is the VPC CIDR
	CIDR core.CIDR
	// Subnets is a list of subnets that have been created.
	Subnets []Subnet
	// SecurityGroups is a list of security groups that have been created.
	SecurityGroups []SecurityGroup
}

// InstanceProfile is an AWS IAM instance profile.
type InstanceProfile struct {
	// Purpose is a logical description of the instance profile.
	Purpose string
	// Name is the name for this instance profile.
	Name string
}

// Role is an AWS IAM role.
type Role struct {
	// Purpose is a logical description of the role.
	Purpose string
	// ARN is the AWS Resource Name for this role.
	ARN string
}

// Subnet is an AWS subnet related to a VPC.
type Subnet struct {
	// Name is a logical name of the subnet.
	Name string
	// ID is the subnet id.
	ID string
	// Zone is the availability zone into which the subnet has been created.
	Zone string
}

// SecurityGroup is an AWS security group related to a VPC.
type SecurityGroup struct {
	// Name is a logical name of the subnet.
	Name string
	// ID is the subnet id.
	ID string
}
