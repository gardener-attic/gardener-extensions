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

package app

import (
	"context"
	"github.com/gardener/gardener-extensions/controllers/os-suse-jeos/pkg/customizer"
	"github.com/gardener/gardener-extensions/controllers/os-suse-jeos/pkg/generator"
	"github.com/gardener/gardener-extensions/pkg/controller/cmd"
	"github.com/gardener/gardener-extensions/pkg/controller/operatingsystemconfig/oscommon/app"
	"github.com/spf13/cobra"
)

// NewControllerCommand returns a new Command with a new Generator
func NewControllerCommand(ctx context.Context) *cobra.Command {
	c, err := customizer.NewCustomizer()
	if err != nil {
		cmd.LogErrAndExit(err, "Could not create JeOS CloudInit customizer")
	}

	g, err := generator.NewCloudInitGenerator()
	if err != nil {
		cmd.LogErrAndExit(err, "Could not create Generator")
	}

	return app.NewControllerCommand(ctx, "suse-jeos", g, c)
}
