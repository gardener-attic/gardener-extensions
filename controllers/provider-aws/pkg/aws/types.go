package aws

import (
	"path/filepath"

	"github.com/gardener/gardener/pkg/utils/imagevector"
	"k8s.io/apimachinery/pkg/util/runtime"
)

const (
	// TerraformerImageName is the name of the Terraformer image.
	TerraformerImageName = "terraformer"
	// AccessKeyID is a constant for the key in a cloud provider secret and backup secret that holds the AWS access key id.
	AccessKeyID = "accessKeyID"
	// SecretAccessKey is a constant for the key in a cloud provider secret and backup secret that holds the AWS secret access key.
	SecretAccessKey = "secretAccessKey"
	// Region is a constant for the key in a backup secret that holds the AWS region.
	Region = "region"
	// TerrformerPurposeInfra is a constant for the complete Terraform setup with purpose 'infrastructure'.
	TerrformerPurposeInfra = "infra"
	// TerraformVariablesKey variables key
	TerraformVariablesKey = "aws-infra/templates/variables.tf"
	// TerraformMainKey main key
	TerraformMainKey = "aws-infra/templates/main.tf"
	// TerraformTFVarsKey tf variables key
	TerraformTFVarsKey = "aws-infra/templates/terraform.tfvars"
	// VPCIDKey is the vpc_id tf state key
	VPCIDKey = "vpc_id"
	// SubnetPublic is the key for accessing public subnets from outputs in terraform
	SubnetPublic = "subnet_public_utility_z0"
	// SubnetNodes is the key for accessing subnet nodes from outputs in terraform
	SubnetNodes = "subnet_nodes_z0"
	// SecurityGroupsNodes is the key for accessing nodes security groups from outputs in terraform
	SecurityGroupsNodes = "security_group_nodes"
	// SSHKeyName key for accessing SSH key name from outputs in terraform
	SSHKeyName = "keyName"
	// IAMInstanceProfileNodes key for accessing Nodes Instance profile from outputs in terraform
	IAMInstanceProfileNodes = "iamInstanceProfileNodes"
	// IAMInstanceProfileBastions key for accessing Bastions Instance profile from outputs in terraform
	IAMInstanceProfileBastions = "iamInstanceProfileBastions"
	// NodesRole role for nodes
	NodesRole = "nodes_role_arn"
	// BastionsRole role for bastions
	BastionsRole = "bastions_role_arn"
)

var (
	// ImageVector is the image vector that contains all the image
	ImageVector imagevector.ImageVector
	// ChartImagesPath is the path to the Helm charts.
	ChartImagesPath = filepath.Join("controllers", "provider-aws", "charts", "images.yaml")
	// TerraformersChartsPath is the terraform charts path
	TerraformersChartsPath = filepath.Join("controllers", "provider-aws", "charts", "internal")
)

func init() {
	var err error
	ImageVector, err = imagevector.ReadImageVector(ChartImagesPath)
	runtime.Must(err)
}
