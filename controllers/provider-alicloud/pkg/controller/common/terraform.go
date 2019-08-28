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

package common

import (
	"time"

	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/alicloud"
	"github.com/gardener/gardener-extensions/controllers/provider-alicloud/pkg/imagevector"
	"github.com/gardener/gardener-extensions/pkg/gardener/terraformer"

	"github.com/gardener/gardener/pkg/logger"
	"k8s.io/client-go/rest"
)

const (
	TerraformVarAccessKeyID     = "TF_VAR_ACCESS_KEY_ID"
	TerraformVarAccessKeySecret = "TF_VAR_ACCESS_KEY_SECRET"
)

// NewTerraformer creates a new Terraformer and initializes it with the credentials.
func NewTerraformer(factory terraformer.Factory, config *rest.Config, credentials *alicloud.Credentials, purpose, namespace, name string) (terraformer.Interface, error) {
	tf, err := factory.NewForConfig(logger.NewLogger("info"), config, purpose, namespace, name, imagevector.TerraformerImage())
	if err != nil {
		return nil, err
	}

	variablesEnvironment := map[string]string{
		TerraformVarAccessKeyID:     credentials.AccessKeyID,
		TerraformVarAccessKeySecret: credentials.AccessKeySecret,
	}

	return tf.
		SetVariablesEnvironment(variablesEnvironment).
		SetJobBackoffLimit(1).
		SetActiveDeadlineSeconds(900).
		SetDeadlineCleaning(5 * time.Minute).
		SetDeadlinePod(5 * time.Minute).
		SetDeadlineJob(15 * time.Minute), nil
}
