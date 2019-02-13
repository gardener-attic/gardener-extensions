package v1alpha1

import (
	"github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureConfig infrastructure configuration resource
type InfrastructureConfig struct {
	metav1.TypeMeta `json:",inline"`
	Networks        Networks `json:"networks"`
}

// CIDR is a string alias.
type CIDR string

// Networks holds information about the Kubernetes and infrastructure networks.
type Networks struct {
	v1alpha1.K8SNetworks `json:",inline"`
	// VPC indicates whether to use an existing VPC or create a new one.
	VPC VPC `json:"vpc"`
	// Internal is a list of private subnets to create (used for internal load balancers).
	Internal []CIDR `json:"internal"`
	// Public is sa list of public subnets to create (used for bastion and load balancers).
	Public []CIDR `json:"public"`
	// Workers is a list of worker subnets (private) to create (used for the VMs).
	Workers []CIDR `json:"workers"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// InfrastructureStatus contains information about created infrastructure resources.
type InfrastructureStatus struct {
	metav1.TypeMeta `json:",inline"`

	// EC2 contains information about the created AWS EC2 resources.
	EC2 EC2 `json:"ec2"`
	// IAM contains information about the created AWS IAM resources.
	IAM IAM `json:"iam"`
	// VPC contains information about the created AWS VPC and some related resources.
	VPC VPC `json:"vpc"`
}

// EC2 contains information about the created AWS EC2 resources.
type EC2 struct {
	// KeyName is the name of the SSH key that has been created.
	KeyName string `json:"keyName"`
}

// IAM contains information about the created AWS IAM resources.
type IAM struct {
	// InstanceProfiles is a list of AWS IAM instance profiles.
	InstanceProfiles []InstanceProfile `json:"instanceProfiles"`
	// Roles is a list of AWS IAM roles.
	Roles []Role `json:"roles"`
}

// VPC contains information about the created AWS VPC and some related resources.
type VPC struct {
	// ID is the VPC id.
	ID string `json:"id"`
	// CIDR is the VPC CIDR
	CIDR CIDR `json:"cidr"`
	// Subnets is a list of subnets that have been created.
	Subnets []Subnet `json:"subnets"`
	// SecurityGroups is a list of security groups that have been created.
	SecurityGroups []SecurityGroup `json:"securityGroups"`
}

// InstanceProfile is an AWS IAM instance profile.
type InstanceProfile struct {
	// Purpose is a logical description of the instance profile.
	Purpose string `json:"purpose"`
	// Name is the name for this instance profile.
	Name string `json:"name"`
}

// Role is an AWS IAM role.
type Role struct {
	// Purpose is a logical description of the role.
	Purpose string `json:"purpose"`
	// ARN is the AWS Resource Name for this role.
	ARN string `json:"arn"`
}

// Subnet is an AWS subnet related to a VPC.
type Subnet struct {
	// Name is a logical name of the subnet.
	Name string `json:"name"`
	// ID is the subnet id.
	ID string `json:"id"`
	// Zone is the availability zone into which the subnet has been created.
	Zone string `json:"zone"`
}

// SecurityGroup is an AWS security group related to a VPC.
type SecurityGroup struct {
	// Name is a logical name of the subnet.
	Name string `json:"name"`
	// ID is the subnet id.
	ID string `json:"id"`
}
