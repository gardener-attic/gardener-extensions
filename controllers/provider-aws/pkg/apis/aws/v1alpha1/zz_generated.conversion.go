// +build !ignore_autogenerated

/*
Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	aws "github.com/gardener/gardener-extensions/controllers/provider-aws/pkg/apis/aws"
	core "github.com/gardener/gardener/pkg/apis/core"
	corev1alpha1 "github.com/gardener/gardener/pkg/apis/core/v1alpha1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*CloudControllerManagerConfig)(nil), (*aws.CloudControllerManagerConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_CloudControllerManagerConfig_To_aws_CloudControllerManagerConfig(a.(*CloudControllerManagerConfig), b.(*aws.CloudControllerManagerConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.CloudControllerManagerConfig)(nil), (*CloudControllerManagerConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_CloudControllerManagerConfig_To_v1alpha1_CloudControllerManagerConfig(a.(*aws.CloudControllerManagerConfig), b.(*CloudControllerManagerConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ControlPlaneConfig)(nil), (*aws.ControlPlaneConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ControlPlaneConfig_To_aws_ControlPlaneConfig(a.(*ControlPlaneConfig), b.(*aws.ControlPlaneConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.ControlPlaneConfig)(nil), (*ControlPlaneConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(a.(*aws.ControlPlaneConfig), b.(*ControlPlaneConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*EC2)(nil), (*aws.EC2)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_EC2_To_aws_EC2(a.(*EC2), b.(*aws.EC2), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.EC2)(nil), (*EC2)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_EC2_To_v1alpha1_EC2(a.(*aws.EC2), b.(*EC2), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*IAM)(nil), (*aws.IAM)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_IAM_To_aws_IAM(a.(*IAM), b.(*aws.IAM), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.IAM)(nil), (*IAM)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_IAM_To_v1alpha1_IAM(a.(*aws.IAM), b.(*IAM), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*InfrastructureConfig)(nil), (*aws.InfrastructureConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_InfrastructureConfig_To_aws_InfrastructureConfig(a.(*InfrastructureConfig), b.(*aws.InfrastructureConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.InfrastructureConfig)(nil), (*InfrastructureConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(a.(*aws.InfrastructureConfig), b.(*InfrastructureConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*InfrastructureStatus)(nil), (*aws.InfrastructureStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_InfrastructureStatus_To_aws_InfrastructureStatus(a.(*InfrastructureStatus), b.(*aws.InfrastructureStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.InfrastructureStatus)(nil), (*InfrastructureStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(a.(*aws.InfrastructureStatus), b.(*InfrastructureStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*InstanceProfile)(nil), (*aws.InstanceProfile)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_InstanceProfile_To_aws_InstanceProfile(a.(*InstanceProfile), b.(*aws.InstanceProfile), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.InstanceProfile)(nil), (*InstanceProfile)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_InstanceProfile_To_v1alpha1_InstanceProfile(a.(*aws.InstanceProfile), b.(*InstanceProfile), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MachineImage)(nil), (*aws.MachineImage)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MachineImage_To_aws_MachineImage(a.(*MachineImage), b.(*aws.MachineImage), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.MachineImage)(nil), (*MachineImage)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_MachineImage_To_v1alpha1_MachineImage(a.(*aws.MachineImage), b.(*MachineImage), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MachineImageVersion)(nil), (*aws.MachineImageVersion)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MachineImageVersion_To_aws_MachineImageVersion(a.(*MachineImageVersion), b.(*aws.MachineImageVersion), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.MachineImageVersion)(nil), (*MachineImageVersion)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_MachineImageVersion_To_v1alpha1_MachineImageVersion(a.(*aws.MachineImageVersion), b.(*MachineImageVersion), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*MachineImages)(nil), (*aws.MachineImages)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_MachineImages_To_aws_MachineImages(a.(*MachineImages), b.(*aws.MachineImages), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.MachineImages)(nil), (*MachineImages)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_MachineImages_To_v1alpha1_MachineImages(a.(*aws.MachineImages), b.(*MachineImages), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*Networks)(nil), (*aws.Networks)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Networks_To_aws_Networks(a.(*Networks), b.(*aws.Networks), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.Networks)(nil), (*Networks)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_Networks_To_v1alpha1_Networks(a.(*aws.Networks), b.(*Networks), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ProviderProfileConfig)(nil), (*aws.ProviderProfileConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ProviderProfileConfig_To_aws_ProviderProfileConfig(a.(*ProviderProfileConfig), b.(*aws.ProviderProfileConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.ProviderProfileConfig)(nil), (*ProviderProfileConfig)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_ProviderProfileConfig_To_v1alpha1_ProviderProfileConfig(a.(*aws.ProviderProfileConfig), b.(*ProviderProfileConfig), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*RegionAMIMapping)(nil), (*aws.RegionAMIMapping)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_RegionAMIMapping_To_aws_RegionAMIMapping(a.(*RegionAMIMapping), b.(*aws.RegionAMIMapping), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.RegionAMIMapping)(nil), (*RegionAMIMapping)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_RegionAMIMapping_To_v1alpha1_RegionAMIMapping(a.(*aws.RegionAMIMapping), b.(*RegionAMIMapping), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*Role)(nil), (*aws.Role)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Role_To_aws_Role(a.(*Role), b.(*aws.Role), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.Role)(nil), (*Role)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_Role_To_v1alpha1_Role(a.(*aws.Role), b.(*Role), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*SecurityGroup)(nil), (*aws.SecurityGroup)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_SecurityGroup_To_aws_SecurityGroup(a.(*SecurityGroup), b.(*aws.SecurityGroup), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.SecurityGroup)(nil), (*SecurityGroup)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_SecurityGroup_To_v1alpha1_SecurityGroup(a.(*aws.SecurityGroup), b.(*SecurityGroup), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*Subnet)(nil), (*aws.Subnet)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Subnet_To_aws_Subnet(a.(*Subnet), b.(*aws.Subnet), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.Subnet)(nil), (*Subnet)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_Subnet_To_v1alpha1_Subnet(a.(*aws.Subnet), b.(*Subnet), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VPC)(nil), (*aws.VPC)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VPC_To_aws_VPC(a.(*VPC), b.(*aws.VPC), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.VPC)(nil), (*VPC)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_VPC_To_v1alpha1_VPC(a.(*aws.VPC), b.(*VPC), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*VPCStatus)(nil), (*aws.VPCStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_VPCStatus_To_aws_VPCStatus(a.(*VPCStatus), b.(*aws.VPCStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.VPCStatus)(nil), (*VPCStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_VPCStatus_To_v1alpha1_VPCStatus(a.(*aws.VPCStatus), b.(*VPCStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*WorkerStatus)(nil), (*aws.WorkerStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_WorkerStatus_To_aws_WorkerStatus(a.(*WorkerStatus), b.(*aws.WorkerStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.WorkerStatus)(nil), (*WorkerStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_WorkerStatus_To_v1alpha1_WorkerStatus(a.(*aws.WorkerStatus), b.(*WorkerStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*Zone)(nil), (*aws.Zone)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Zone_To_aws_Zone(a.(*Zone), b.(*aws.Zone), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*aws.Zone)(nil), (*Zone)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_aws_Zone_To_v1alpha1_Zone(a.(*aws.Zone), b.(*Zone), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_CloudControllerManagerConfig_To_aws_CloudControllerManagerConfig(in *CloudControllerManagerConfig, out *aws.CloudControllerManagerConfig, s conversion.Scope) error {
	out.FeatureGates = *(*map[string]bool)(unsafe.Pointer(&in.FeatureGates))
	return nil
}

// Convert_v1alpha1_CloudControllerManagerConfig_To_aws_CloudControllerManagerConfig is an autogenerated conversion function.
func Convert_v1alpha1_CloudControllerManagerConfig_To_aws_CloudControllerManagerConfig(in *CloudControllerManagerConfig, out *aws.CloudControllerManagerConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_CloudControllerManagerConfig_To_aws_CloudControllerManagerConfig(in, out, s)
}

func autoConvert_aws_CloudControllerManagerConfig_To_v1alpha1_CloudControllerManagerConfig(in *aws.CloudControllerManagerConfig, out *CloudControllerManagerConfig, s conversion.Scope) error {
	out.FeatureGates = *(*map[string]bool)(unsafe.Pointer(&in.FeatureGates))
	return nil
}

// Convert_aws_CloudControllerManagerConfig_To_v1alpha1_CloudControllerManagerConfig is an autogenerated conversion function.
func Convert_aws_CloudControllerManagerConfig_To_v1alpha1_CloudControllerManagerConfig(in *aws.CloudControllerManagerConfig, out *CloudControllerManagerConfig, s conversion.Scope) error {
	return autoConvert_aws_CloudControllerManagerConfig_To_v1alpha1_CloudControllerManagerConfig(in, out, s)
}

func autoConvert_v1alpha1_ControlPlaneConfig_To_aws_ControlPlaneConfig(in *ControlPlaneConfig, out *aws.ControlPlaneConfig, s conversion.Scope) error {
	out.CloudControllerManager = (*aws.CloudControllerManagerConfig)(unsafe.Pointer(in.CloudControllerManager))
	return nil
}

// Convert_v1alpha1_ControlPlaneConfig_To_aws_ControlPlaneConfig is an autogenerated conversion function.
func Convert_v1alpha1_ControlPlaneConfig_To_aws_ControlPlaneConfig(in *ControlPlaneConfig, out *aws.ControlPlaneConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_ControlPlaneConfig_To_aws_ControlPlaneConfig(in, out, s)
}

func autoConvert_aws_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(in *aws.ControlPlaneConfig, out *ControlPlaneConfig, s conversion.Scope) error {
	out.CloudControllerManager = (*CloudControllerManagerConfig)(unsafe.Pointer(in.CloudControllerManager))
	return nil
}

// Convert_aws_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig is an autogenerated conversion function.
func Convert_aws_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(in *aws.ControlPlaneConfig, out *ControlPlaneConfig, s conversion.Scope) error {
	return autoConvert_aws_ControlPlaneConfig_To_v1alpha1_ControlPlaneConfig(in, out, s)
}

func autoConvert_v1alpha1_EC2_To_aws_EC2(in *EC2, out *aws.EC2, s conversion.Scope) error {
	out.KeyName = in.KeyName
	return nil
}

// Convert_v1alpha1_EC2_To_aws_EC2 is an autogenerated conversion function.
func Convert_v1alpha1_EC2_To_aws_EC2(in *EC2, out *aws.EC2, s conversion.Scope) error {
	return autoConvert_v1alpha1_EC2_To_aws_EC2(in, out, s)
}

func autoConvert_aws_EC2_To_v1alpha1_EC2(in *aws.EC2, out *EC2, s conversion.Scope) error {
	out.KeyName = in.KeyName
	return nil
}

// Convert_aws_EC2_To_v1alpha1_EC2 is an autogenerated conversion function.
func Convert_aws_EC2_To_v1alpha1_EC2(in *aws.EC2, out *EC2, s conversion.Scope) error {
	return autoConvert_aws_EC2_To_v1alpha1_EC2(in, out, s)
}

func autoConvert_v1alpha1_IAM_To_aws_IAM(in *IAM, out *aws.IAM, s conversion.Scope) error {
	out.InstanceProfiles = *(*[]aws.InstanceProfile)(unsafe.Pointer(&in.InstanceProfiles))
	out.Roles = *(*[]aws.Role)(unsafe.Pointer(&in.Roles))
	return nil
}

// Convert_v1alpha1_IAM_To_aws_IAM is an autogenerated conversion function.
func Convert_v1alpha1_IAM_To_aws_IAM(in *IAM, out *aws.IAM, s conversion.Scope) error {
	return autoConvert_v1alpha1_IAM_To_aws_IAM(in, out, s)
}

func autoConvert_aws_IAM_To_v1alpha1_IAM(in *aws.IAM, out *IAM, s conversion.Scope) error {
	out.InstanceProfiles = *(*[]InstanceProfile)(unsafe.Pointer(&in.InstanceProfiles))
	out.Roles = *(*[]Role)(unsafe.Pointer(&in.Roles))
	return nil
}

// Convert_aws_IAM_To_v1alpha1_IAM is an autogenerated conversion function.
func Convert_aws_IAM_To_v1alpha1_IAM(in *aws.IAM, out *IAM, s conversion.Scope) error {
	return autoConvert_aws_IAM_To_v1alpha1_IAM(in, out, s)
}

func autoConvert_v1alpha1_InfrastructureConfig_To_aws_InfrastructureConfig(in *InfrastructureConfig, out *aws.InfrastructureConfig, s conversion.Scope) error {
	if err := Convert_v1alpha1_Networks_To_aws_Networks(&in.Networks, &out.Networks, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_InfrastructureConfig_To_aws_InfrastructureConfig is an autogenerated conversion function.
func Convert_v1alpha1_InfrastructureConfig_To_aws_InfrastructureConfig(in *InfrastructureConfig, out *aws.InfrastructureConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_InfrastructureConfig_To_aws_InfrastructureConfig(in, out, s)
}

func autoConvert_aws_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(in *aws.InfrastructureConfig, out *InfrastructureConfig, s conversion.Scope) error {
	if err := Convert_aws_Networks_To_v1alpha1_Networks(&in.Networks, &out.Networks, s); err != nil {
		return err
	}
	return nil
}

// Convert_aws_InfrastructureConfig_To_v1alpha1_InfrastructureConfig is an autogenerated conversion function.
func Convert_aws_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(in *aws.InfrastructureConfig, out *InfrastructureConfig, s conversion.Scope) error {
	return autoConvert_aws_InfrastructureConfig_To_v1alpha1_InfrastructureConfig(in, out, s)
}

func autoConvert_v1alpha1_InfrastructureStatus_To_aws_InfrastructureStatus(in *InfrastructureStatus, out *aws.InfrastructureStatus, s conversion.Scope) error {
	if err := Convert_v1alpha1_EC2_To_aws_EC2(&in.EC2, &out.EC2, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_IAM_To_aws_IAM(&in.IAM, &out.IAM, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_VPCStatus_To_aws_VPCStatus(&in.VPC, &out.VPC, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_InfrastructureStatus_To_aws_InfrastructureStatus is an autogenerated conversion function.
func Convert_v1alpha1_InfrastructureStatus_To_aws_InfrastructureStatus(in *InfrastructureStatus, out *aws.InfrastructureStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_InfrastructureStatus_To_aws_InfrastructureStatus(in, out, s)
}

func autoConvert_aws_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in *aws.InfrastructureStatus, out *InfrastructureStatus, s conversion.Scope) error {
	if err := Convert_aws_EC2_To_v1alpha1_EC2(&in.EC2, &out.EC2, s); err != nil {
		return err
	}
	if err := Convert_aws_IAM_To_v1alpha1_IAM(&in.IAM, &out.IAM, s); err != nil {
		return err
	}
	if err := Convert_aws_VPCStatus_To_v1alpha1_VPCStatus(&in.VPC, &out.VPC, s); err != nil {
		return err
	}
	return nil
}

// Convert_aws_InfrastructureStatus_To_v1alpha1_InfrastructureStatus is an autogenerated conversion function.
func Convert_aws_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in *aws.InfrastructureStatus, out *InfrastructureStatus, s conversion.Scope) error {
	return autoConvert_aws_InfrastructureStatus_To_v1alpha1_InfrastructureStatus(in, out, s)
}

func autoConvert_v1alpha1_InstanceProfile_To_aws_InstanceProfile(in *InstanceProfile, out *aws.InstanceProfile, s conversion.Scope) error {
	out.Purpose = in.Purpose
	out.Name = in.Name
	return nil
}

// Convert_v1alpha1_InstanceProfile_To_aws_InstanceProfile is an autogenerated conversion function.
func Convert_v1alpha1_InstanceProfile_To_aws_InstanceProfile(in *InstanceProfile, out *aws.InstanceProfile, s conversion.Scope) error {
	return autoConvert_v1alpha1_InstanceProfile_To_aws_InstanceProfile(in, out, s)
}

func autoConvert_aws_InstanceProfile_To_v1alpha1_InstanceProfile(in *aws.InstanceProfile, out *InstanceProfile, s conversion.Scope) error {
	out.Purpose = in.Purpose
	out.Name = in.Name
	return nil
}

// Convert_aws_InstanceProfile_To_v1alpha1_InstanceProfile is an autogenerated conversion function.
func Convert_aws_InstanceProfile_To_v1alpha1_InstanceProfile(in *aws.InstanceProfile, out *InstanceProfile, s conversion.Scope) error {
	return autoConvert_aws_InstanceProfile_To_v1alpha1_InstanceProfile(in, out, s)
}

func autoConvert_v1alpha1_MachineImage_To_aws_MachineImage(in *MachineImage, out *aws.MachineImage, s conversion.Scope) error {
	out.Name = in.Name
	out.Version = in.Version
	out.AMI = in.AMI
	return nil
}

// Convert_v1alpha1_MachineImage_To_aws_MachineImage is an autogenerated conversion function.
func Convert_v1alpha1_MachineImage_To_aws_MachineImage(in *MachineImage, out *aws.MachineImage, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineImage_To_aws_MachineImage(in, out, s)
}

func autoConvert_aws_MachineImage_To_v1alpha1_MachineImage(in *aws.MachineImage, out *MachineImage, s conversion.Scope) error {
	out.Name = in.Name
	out.Version = in.Version
	out.AMI = in.AMI
	return nil
}

// Convert_aws_MachineImage_To_v1alpha1_MachineImage is an autogenerated conversion function.
func Convert_aws_MachineImage_To_v1alpha1_MachineImage(in *aws.MachineImage, out *MachineImage, s conversion.Scope) error {
	return autoConvert_aws_MachineImage_To_v1alpha1_MachineImage(in, out, s)
}

func autoConvert_v1alpha1_MachineImageVersion_To_aws_MachineImageVersion(in *MachineImageVersion, out *aws.MachineImageVersion, s conversion.Scope) error {
	out.Version = in.Version
	out.Regions = *(*[]aws.RegionAMIMapping)(unsafe.Pointer(&in.Regions))
	return nil
}

// Convert_v1alpha1_MachineImageVersion_To_aws_MachineImageVersion is an autogenerated conversion function.
func Convert_v1alpha1_MachineImageVersion_To_aws_MachineImageVersion(in *MachineImageVersion, out *aws.MachineImageVersion, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineImageVersion_To_aws_MachineImageVersion(in, out, s)
}

func autoConvert_aws_MachineImageVersion_To_v1alpha1_MachineImageVersion(in *aws.MachineImageVersion, out *MachineImageVersion, s conversion.Scope) error {
	out.Version = in.Version
	out.Regions = *(*[]RegionAMIMapping)(unsafe.Pointer(&in.Regions))
	return nil
}

// Convert_aws_MachineImageVersion_To_v1alpha1_MachineImageVersion is an autogenerated conversion function.
func Convert_aws_MachineImageVersion_To_v1alpha1_MachineImageVersion(in *aws.MachineImageVersion, out *MachineImageVersion, s conversion.Scope) error {
	return autoConvert_aws_MachineImageVersion_To_v1alpha1_MachineImageVersion(in, out, s)
}

func autoConvert_v1alpha1_MachineImages_To_aws_MachineImages(in *MachineImages, out *aws.MachineImages, s conversion.Scope) error {
	out.Name = in.Name
	out.Versions = *(*[]aws.MachineImageVersion)(unsafe.Pointer(&in.Versions))
	return nil
}

// Convert_v1alpha1_MachineImages_To_aws_MachineImages is an autogenerated conversion function.
func Convert_v1alpha1_MachineImages_To_aws_MachineImages(in *MachineImages, out *aws.MachineImages, s conversion.Scope) error {
	return autoConvert_v1alpha1_MachineImages_To_aws_MachineImages(in, out, s)
}

func autoConvert_aws_MachineImages_To_v1alpha1_MachineImages(in *aws.MachineImages, out *MachineImages, s conversion.Scope) error {
	out.Name = in.Name
	out.Versions = *(*[]MachineImageVersion)(unsafe.Pointer(&in.Versions))
	return nil
}

// Convert_aws_MachineImages_To_v1alpha1_MachineImages is an autogenerated conversion function.
func Convert_aws_MachineImages_To_v1alpha1_MachineImages(in *aws.MachineImages, out *MachineImages, s conversion.Scope) error {
	return autoConvert_aws_MachineImages_To_v1alpha1_MachineImages(in, out, s)
}

func autoConvert_v1alpha1_Networks_To_aws_Networks(in *Networks, out *aws.Networks, s conversion.Scope) error {
	if err := Convert_v1alpha1_VPC_To_aws_VPC(&in.VPC, &out.VPC, s); err != nil {
		return err
	}
	out.Zones = *(*[]aws.Zone)(unsafe.Pointer(&in.Zones))
	return nil
}

// Convert_v1alpha1_Networks_To_aws_Networks is an autogenerated conversion function.
func Convert_v1alpha1_Networks_To_aws_Networks(in *Networks, out *aws.Networks, s conversion.Scope) error {
	return autoConvert_v1alpha1_Networks_To_aws_Networks(in, out, s)
}

func autoConvert_aws_Networks_To_v1alpha1_Networks(in *aws.Networks, out *Networks, s conversion.Scope) error {
	if err := Convert_aws_VPC_To_v1alpha1_VPC(&in.VPC, &out.VPC, s); err != nil {
		return err
	}
	out.Zones = *(*[]Zone)(unsafe.Pointer(&in.Zones))
	return nil
}

// Convert_aws_Networks_To_v1alpha1_Networks is an autogenerated conversion function.
func Convert_aws_Networks_To_v1alpha1_Networks(in *aws.Networks, out *Networks, s conversion.Scope) error {
	return autoConvert_aws_Networks_To_v1alpha1_Networks(in, out, s)
}

func autoConvert_v1alpha1_ProviderProfileConfig_To_aws_ProviderProfileConfig(in *ProviderProfileConfig, out *aws.ProviderProfileConfig, s conversion.Scope) error {
	out.MachineImages = *(*[]aws.MachineImages)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_v1alpha1_ProviderProfileConfig_To_aws_ProviderProfileConfig is an autogenerated conversion function.
func Convert_v1alpha1_ProviderProfileConfig_To_aws_ProviderProfileConfig(in *ProviderProfileConfig, out *aws.ProviderProfileConfig, s conversion.Scope) error {
	return autoConvert_v1alpha1_ProviderProfileConfig_To_aws_ProviderProfileConfig(in, out, s)
}

func autoConvert_aws_ProviderProfileConfig_To_v1alpha1_ProviderProfileConfig(in *aws.ProviderProfileConfig, out *ProviderProfileConfig, s conversion.Scope) error {
	out.MachineImages = *(*[]MachineImages)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_aws_ProviderProfileConfig_To_v1alpha1_ProviderProfileConfig is an autogenerated conversion function.
func Convert_aws_ProviderProfileConfig_To_v1alpha1_ProviderProfileConfig(in *aws.ProviderProfileConfig, out *ProviderProfileConfig, s conversion.Scope) error {
	return autoConvert_aws_ProviderProfileConfig_To_v1alpha1_ProviderProfileConfig(in, out, s)
}

func autoConvert_v1alpha1_RegionAMIMapping_To_aws_RegionAMIMapping(in *RegionAMIMapping, out *aws.RegionAMIMapping, s conversion.Scope) error {
	out.Name = in.Name
	out.AMI = in.AMI
	return nil
}

// Convert_v1alpha1_RegionAMIMapping_To_aws_RegionAMIMapping is an autogenerated conversion function.
func Convert_v1alpha1_RegionAMIMapping_To_aws_RegionAMIMapping(in *RegionAMIMapping, out *aws.RegionAMIMapping, s conversion.Scope) error {
	return autoConvert_v1alpha1_RegionAMIMapping_To_aws_RegionAMIMapping(in, out, s)
}

func autoConvert_aws_RegionAMIMapping_To_v1alpha1_RegionAMIMapping(in *aws.RegionAMIMapping, out *RegionAMIMapping, s conversion.Scope) error {
	out.Name = in.Name
	out.AMI = in.AMI
	return nil
}

// Convert_aws_RegionAMIMapping_To_v1alpha1_RegionAMIMapping is an autogenerated conversion function.
func Convert_aws_RegionAMIMapping_To_v1alpha1_RegionAMIMapping(in *aws.RegionAMIMapping, out *RegionAMIMapping, s conversion.Scope) error {
	return autoConvert_aws_RegionAMIMapping_To_v1alpha1_RegionAMIMapping(in, out, s)
}

func autoConvert_v1alpha1_Role_To_aws_Role(in *Role, out *aws.Role, s conversion.Scope) error {
	out.Purpose = in.Purpose
	out.ARN = in.ARN
	return nil
}

// Convert_v1alpha1_Role_To_aws_Role is an autogenerated conversion function.
func Convert_v1alpha1_Role_To_aws_Role(in *Role, out *aws.Role, s conversion.Scope) error {
	return autoConvert_v1alpha1_Role_To_aws_Role(in, out, s)
}

func autoConvert_aws_Role_To_v1alpha1_Role(in *aws.Role, out *Role, s conversion.Scope) error {
	out.Purpose = in.Purpose
	out.ARN = in.ARN
	return nil
}

// Convert_aws_Role_To_v1alpha1_Role is an autogenerated conversion function.
func Convert_aws_Role_To_v1alpha1_Role(in *aws.Role, out *Role, s conversion.Scope) error {
	return autoConvert_aws_Role_To_v1alpha1_Role(in, out, s)
}

func autoConvert_v1alpha1_SecurityGroup_To_aws_SecurityGroup(in *SecurityGroup, out *aws.SecurityGroup, s conversion.Scope) error {
	out.Purpose = in.Purpose
	out.ID = in.ID
	return nil
}

// Convert_v1alpha1_SecurityGroup_To_aws_SecurityGroup is an autogenerated conversion function.
func Convert_v1alpha1_SecurityGroup_To_aws_SecurityGroup(in *SecurityGroup, out *aws.SecurityGroup, s conversion.Scope) error {
	return autoConvert_v1alpha1_SecurityGroup_To_aws_SecurityGroup(in, out, s)
}

func autoConvert_aws_SecurityGroup_To_v1alpha1_SecurityGroup(in *aws.SecurityGroup, out *SecurityGroup, s conversion.Scope) error {
	out.Purpose = in.Purpose
	out.ID = in.ID
	return nil
}

// Convert_aws_SecurityGroup_To_v1alpha1_SecurityGroup is an autogenerated conversion function.
func Convert_aws_SecurityGroup_To_v1alpha1_SecurityGroup(in *aws.SecurityGroup, out *SecurityGroup, s conversion.Scope) error {
	return autoConvert_aws_SecurityGroup_To_v1alpha1_SecurityGroup(in, out, s)
}

func autoConvert_v1alpha1_Subnet_To_aws_Subnet(in *Subnet, out *aws.Subnet, s conversion.Scope) error {
	out.Purpose = in.Purpose
	out.ID = in.ID
	out.Zone = in.Zone
	return nil
}

// Convert_v1alpha1_Subnet_To_aws_Subnet is an autogenerated conversion function.
func Convert_v1alpha1_Subnet_To_aws_Subnet(in *Subnet, out *aws.Subnet, s conversion.Scope) error {
	return autoConvert_v1alpha1_Subnet_To_aws_Subnet(in, out, s)
}

func autoConvert_aws_Subnet_To_v1alpha1_Subnet(in *aws.Subnet, out *Subnet, s conversion.Scope) error {
	out.Purpose = in.Purpose
	out.ID = in.ID
	out.Zone = in.Zone
	return nil
}

// Convert_aws_Subnet_To_v1alpha1_Subnet is an autogenerated conversion function.
func Convert_aws_Subnet_To_v1alpha1_Subnet(in *aws.Subnet, out *Subnet, s conversion.Scope) error {
	return autoConvert_aws_Subnet_To_v1alpha1_Subnet(in, out, s)
}

func autoConvert_v1alpha1_VPC_To_aws_VPC(in *VPC, out *aws.VPC, s conversion.Scope) error {
	out.ID = (*string)(unsafe.Pointer(in.ID))
	out.CIDR = (*core.CIDR)(unsafe.Pointer(in.CIDR))
	return nil
}

// Convert_v1alpha1_VPC_To_aws_VPC is an autogenerated conversion function.
func Convert_v1alpha1_VPC_To_aws_VPC(in *VPC, out *aws.VPC, s conversion.Scope) error {
	return autoConvert_v1alpha1_VPC_To_aws_VPC(in, out, s)
}

func autoConvert_aws_VPC_To_v1alpha1_VPC(in *aws.VPC, out *VPC, s conversion.Scope) error {
	out.ID = (*string)(unsafe.Pointer(in.ID))
	out.CIDR = (*corev1alpha1.CIDR)(unsafe.Pointer(in.CIDR))
	return nil
}

// Convert_aws_VPC_To_v1alpha1_VPC is an autogenerated conversion function.
func Convert_aws_VPC_To_v1alpha1_VPC(in *aws.VPC, out *VPC, s conversion.Scope) error {
	return autoConvert_aws_VPC_To_v1alpha1_VPC(in, out, s)
}

func autoConvert_v1alpha1_VPCStatus_To_aws_VPCStatus(in *VPCStatus, out *aws.VPCStatus, s conversion.Scope) error {
	out.ID = in.ID
	out.Subnets = *(*[]aws.Subnet)(unsafe.Pointer(&in.Subnets))
	out.SecurityGroups = *(*[]aws.SecurityGroup)(unsafe.Pointer(&in.SecurityGroups))
	return nil
}

// Convert_v1alpha1_VPCStatus_To_aws_VPCStatus is an autogenerated conversion function.
func Convert_v1alpha1_VPCStatus_To_aws_VPCStatus(in *VPCStatus, out *aws.VPCStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_VPCStatus_To_aws_VPCStatus(in, out, s)
}

func autoConvert_aws_VPCStatus_To_v1alpha1_VPCStatus(in *aws.VPCStatus, out *VPCStatus, s conversion.Scope) error {
	out.ID = in.ID
	out.Subnets = *(*[]Subnet)(unsafe.Pointer(&in.Subnets))
	out.SecurityGroups = *(*[]SecurityGroup)(unsafe.Pointer(&in.SecurityGroups))
	return nil
}

// Convert_aws_VPCStatus_To_v1alpha1_VPCStatus is an autogenerated conversion function.
func Convert_aws_VPCStatus_To_v1alpha1_VPCStatus(in *aws.VPCStatus, out *VPCStatus, s conversion.Scope) error {
	return autoConvert_aws_VPCStatus_To_v1alpha1_VPCStatus(in, out, s)
}

func autoConvert_v1alpha1_WorkerStatus_To_aws_WorkerStatus(in *WorkerStatus, out *aws.WorkerStatus, s conversion.Scope) error {
	out.MachineImages = *(*[]aws.MachineImage)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_v1alpha1_WorkerStatus_To_aws_WorkerStatus is an autogenerated conversion function.
func Convert_v1alpha1_WorkerStatus_To_aws_WorkerStatus(in *WorkerStatus, out *aws.WorkerStatus, s conversion.Scope) error {
	return autoConvert_v1alpha1_WorkerStatus_To_aws_WorkerStatus(in, out, s)
}

func autoConvert_aws_WorkerStatus_To_v1alpha1_WorkerStatus(in *aws.WorkerStatus, out *WorkerStatus, s conversion.Scope) error {
	out.MachineImages = *(*[]MachineImage)(unsafe.Pointer(&in.MachineImages))
	return nil
}

// Convert_aws_WorkerStatus_To_v1alpha1_WorkerStatus is an autogenerated conversion function.
func Convert_aws_WorkerStatus_To_v1alpha1_WorkerStatus(in *aws.WorkerStatus, out *WorkerStatus, s conversion.Scope) error {
	return autoConvert_aws_WorkerStatus_To_v1alpha1_WorkerStatus(in, out, s)
}

func autoConvert_v1alpha1_Zone_To_aws_Zone(in *Zone, out *aws.Zone, s conversion.Scope) error {
	out.Name = in.Name
	out.Internal = core.CIDR(in.Internal)
	out.Public = core.CIDR(in.Public)
	out.Workers = core.CIDR(in.Workers)
	return nil
}

// Convert_v1alpha1_Zone_To_aws_Zone is an autogenerated conversion function.
func Convert_v1alpha1_Zone_To_aws_Zone(in *Zone, out *aws.Zone, s conversion.Scope) error {
	return autoConvert_v1alpha1_Zone_To_aws_Zone(in, out, s)
}

func autoConvert_aws_Zone_To_v1alpha1_Zone(in *aws.Zone, out *Zone, s conversion.Scope) error {
	out.Name = in.Name
	out.Internal = corev1alpha1.CIDR(in.Internal)
	out.Public = corev1alpha1.CIDR(in.Public)
	out.Workers = corev1alpha1.CIDR(in.Workers)
	return nil
}

// Convert_aws_Zone_To_v1alpha1_Zone is an autogenerated conversion function.
func Convert_aws_Zone_To_v1alpha1_Zone(in *aws.Zone, out *Zone, s conversion.Scope) error {
	return autoConvert_aws_Zone_To_v1alpha1_Zone(in, out, s)
}
