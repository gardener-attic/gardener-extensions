// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/slb"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/sts"
	alicloudvpc "github.com/aliyun/alibaba-cloud-sdk-go/services/vpc"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// FactoryFunc is a function that implements the Factory interface. Used for consuming the
// `alicloudvpc.NewClientWithAccessKey` function.
type FactoryFunc func(region, accessKeyID, accessKeySecret string) (*alicloudvpc.Client, error)

// NewVPC implements Factory.
func (f FactoryFunc) NewVPC(region, accessKeyID, accessKeySecret string) (VPC, error) {
	return f(region, accessKeyID, accessKeySecret)
}

// DefaultFactory instantiates a default Factory.
func DefaultFactory() Factory {
	return FactoryFunc(alicloudvpc.NewClientWithAccessKey)
}

type storageClient struct {
	client *oss.Client
}

// NewStorageClientFromSecretRef creates a new Alicloud storage Client using the credentials from <secretRef>.
func NewStorageClientFromSecretRef(ctx context.Context, client client.Client, secretRef *corev1.SecretReference, region string) (Storage, error) {
	credentials, err := alicloud.ReadCredentialsFromSecretRef(ctx, client, secretRef)
	if err != nil {
		return nil, err
	}

	ossClient, err := oss.New(ComputeStorageEndpoint(region), credentials.AccessKeyID, credentials.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	storageClient := &storageClient{
		client: ossClient,
	}
	return storageClient, nil
}

// DeleteObjectsWithPrefix deletes the s3 objects with the specific <prefix> from <bucketName>. If it does not exist,
// no error is returned.
func (c *storageClient) DeleteObjectsWithPrefix(ctx context.Context, bucketName, prefix string) error {
	bucket, err := c.client.Bucket(bucketName)
	if err != nil {
		return err
	}

	var expirationOption oss.Option
	t, ok := ctx.Deadline()
	if ok {
		expirationOption = oss.Expires(t)
	}

	marker := ""
	for {
		lsRes, err := bucket.ListObjects(oss.Marker(marker), oss.Prefix(prefix), oss.MaxKeys(1000), expirationOption)

		if err != nil {
			return err
		}

		var objectKeys []string
		for _, object := range lsRes.Objects {
			objectKeys = append(objectKeys, object.Key)
		}

		if len(objectKeys) != 0 {
			if _, err := bucket.DeleteObjects(objectKeys, oss.DeleteObjectsQuiet(true), expirationOption); err != nil {
				return err
			}
		}

		if lsRes.IsTruncated {
			marker = lsRes.NextMarker
		} else {
			return nil
		}
	}
}

// CreateBucketIfNotExists creates the OSS bucket with name <bucketName> in <region>. If it already exist,
// no error is returned.
func (c *storageClient) CreateBucketIfNotExists(ctx context.Context, bucketName string) error {
	var expirationOption oss.Option
	t, ok := ctx.Deadline()
	if ok {
		expirationOption = oss.Expires(t)
	}

	if err := c.client.CreateBucket(bucketName, oss.StorageClass(oss.StorageStandard), expirationOption); err != nil {
		if ossErr, ok := err.(oss.ServiceError); !ok {
			return err
		} else if ossErr.StatusCode != http.StatusConflict {
			return err
		}
	}

	rules := []oss.LifecycleRule{
		{
			Prefix: "",
			Status: "Enabled",
			AbortMultipartUpload: &oss.LifecycleAbortMultipartUpload{
				Days: 7,
			},
		},
	}
	return c.client.SetBucketLifecycle(bucketName, rules)
}

// DeleteBucketIfExists deletes the Alicloud OSS bucket with name <bucketName>. If it does not exist,
// no error is returned.
func (c *storageClient) DeleteBucketIfExists(ctx context.Context, bucketName string) error {
	if err := c.client.DeleteBucket(bucketName); err != nil {
		if ossErr, ok := err.(oss.ServiceError); ok {
			switch ossErr.StatusCode {
			case http.StatusNotFound:
				return nil

			case http.StatusConflict:
				if err := c.DeleteObjectsWithPrefix(ctx, bucketName, ""); err != nil {
					return err
				}
				return c.DeleteBucketIfExists(ctx, bucketName)

			default:
				return ossErr
			}
		}
	}
	return nil
}

// ComputeStorageEndpoint computes the OSS storage endpoint based on the given region.
func ComputeStorageEndpoint(region string) string {
	return fmt.Sprintf("https://oss-%s.aliyuncs.com/", region)
}

type clientFactory struct {
}

// NewClientFactory creates a new clientFactory instance that can be used to instantiate Alicloud clients
func NewClientFactory() ClientFactory {
	return &clientFactory{}
}

type ecsClient struct {
	client *ecs.Client
}

type stsClient struct {
	client *sts.Client
}

type slbClient struct {
	client *slb.Client
}

// NewECSClient creates a new ECS client with given region, AccessKeyID, and AccessKeySecret
func (f *clientFactory) NewECSClient(ctx context.Context, region, accessKeyID, accessKeySecret string) (ECS, error) {
	client, err := ecs.NewClientWithAccessKey(region, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}

	return &ecsClient{
		client: client,
	}, nil
}

// CheckIfImageExists checks whether given imageID can be accessed by the client
func (c *ecsClient) CheckIfImageExists(ctx context.Context, imageID string) (bool, error) {
	request := ecs.CreateDescribeImagesRequest()
	request.ImageId = imageID
	request.SetScheme("HTTPS")
	response, err := c.client.DescribeImages(request)
	if err != nil {
		return false, err
	}
	return response.TotalCount > 0, nil
}

// ShareImageToAccount shares the given image to target account from current client
func (c *ecsClient) ShareImageToAccount(ctx context.Context, regionID, imageID, accountID string) error {
	request := ecs.CreateModifyImageSharePermissionRequest()
	request.RegionId = regionID
	request.ImageId = imageID
	request.AddAccount = &[]string{accountID}
	request.SetScheme("HTTPS")
	_, err := c.client.ModifyImageSharePermission(request)
	return err
}

// NewSTSClient creates a new STS client with given region, AccessKeyID, and AccessKeySecret
func (f *clientFactory) NewSTSClient(ctx context.Context, region, accessKeyID, accessKeySecret string) (STS, error) {
	client, err := sts.NewClientWithAccessKey(region, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}

	return &stsClient{
		client: client,
	}, nil
}

// GetAccountIDFromCallerIdentity gets caller's accountID
func (c *stsClient) GetAccountIDFromCallerIdentity(ctx context.Context) (string, error) {
	request := sts.CreateGetCallerIdentityRequest()
	request.SetScheme("HTTPS")
	response, err := c.client.GetCallerIdentity(request)
	if err != nil {
		return "", err
	}
	return response.AccountId, nil
}

// NewSLBClient creates a new SLB client with given region, AccessKeyID, and AccessKeySecret
func (f *clientFactory) NewSLBClient(ctx context.Context, region, accessKeyID, accessKeySecret string) (SLB, error) {
	client, err := slb.NewClientWithAccessKey(region, accessKeyID, accessKeySecret)
	if err != nil {
		return nil, err
	}

	return &slbClient{
		client: client,
	}, nil
}

// GetLoadBalancerIDs gets LoadBalancerIDs from all LoadBalancers in the given region
func (c *slbClient) GetLoadBalancerIDs(ctx context.Context, region string) ([]string, error) {
	var (
		loadBalancerIDs []string
		pageNumber      = 1
		pageSize        = 100
		request         = slb.CreateDescribeLoadBalancersRequest()
	)
	request.SetScheme("HTTPS")
	request.RegionId = region
	request.PageSize = requests.NewInteger(pageSize)

	for {
		request.PageNumber = requests.NewInteger(pageNumber)
		response, err := c.client.DescribeLoadBalancers(request)
		if err != nil {
			return nil, err
		}
		for _, loadBalancer := range response.LoadBalancers.LoadBalancer {
			loadBalancerIDs = append(loadBalancerIDs, loadBalancer.LoadBalancerId)
		}

		if pageNumber*pageSize >= response.TotalCount {
			break
		}
		pageNumber++
	}
	return loadBalancerIDs, nil
}

// GetFirstVServerGroupName gets the VServerGroupName of the first VServerGroup in the LoadBalancer with given region and loadBalancerID
func (c *slbClient) GetFirstVServerGroupName(ctx context.Context, region, loadBalancerID string) (string, error) {
	request := slb.CreateDescribeVServerGroupsRequest()
	request.SetScheme("HTTPS")
	request.RegionId = region
	request.LoadBalancerId = loadBalancerID
	response, err := c.client.DescribeVServerGroups(request)
	if err != nil {
		return "", err
	}
	if len(response.VServerGroups.VServerGroup) == 0 {
		return "", nil
	}
	return response.VServerGroups.VServerGroup[0].VServerGroupName, nil
}

// DeleteLoadBalancer deletes the LoadBalancer with given region and loadBalancerID
func (c *slbClient) DeleteLoadBalancer(ctx context.Context, region, loadBalancerID string) error {
	request := slb.CreateDeleteLoadBalancerRequest()
	request.SetScheme("HTTPS")
	request.RegionId = region
	request.LoadBalancerId = loadBalancerID
	_, err := c.client.DeleteLoadBalancer(request)
	return err
}
