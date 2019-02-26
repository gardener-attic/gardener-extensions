package awsclient

import (
	"context"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/sts"
)

// ClientInterface is an interface which must be implemented by AWS clients.
type ClientInterface interface {
	GetAccountID(ctx context.Context) (string, error)
	GetInternetGateway(ctx context.Context, vpcID string) (string, error)

	// The following functions are only temporary needed due to https://github.com/gardener/gardener/issues/129.
	ListKubernetesELBs(ctx context.Context, vpcID, clusterName string) ([]string, error)
	ListKubernetesSecurityGroups(ctx context.Context, vpcID, clusterName string) ([]string, error)
	DeleteELB(ctx context.Context, name string) error
	DeleteSecurityGroup(ctx context.Context, id string) error
}

// Client is a struct containing several clients for the different AWS services it needs to interact with.
// * EC2 is the standard client for the EC2 service.
// * ELB is the standard client for the ELB service.
// * STS is the standard client for the STS service.
type Client struct {
	EC2 *ec2.EC2
	ELB *elb.ELB
	STS *sts.STS
}
