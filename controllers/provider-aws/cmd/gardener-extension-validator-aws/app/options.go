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
	"github.com/spf13/pflag"
)

const (
	// CertDirFlag is the name of the command line flag to specify the directory that contains the webhook server key and certificate.
	CertDirFlag = "webhook-config-cert-dir"
)

// ServerOptions are command line options that can be set for ServerConfig.
type ServerOptions struct {
	// CertDir is the directory that contains the webhook server key and certificate.
	CertDir string

	config *ServerConfig
}

// ServerConfig is a completed webhook server configuration.
type ServerConfig struct {
	// CertDir is the directory that contains the webhook server key and certificate.
	CertDir string
}

// Complete implements Completer.Complete.
func (w *ServerOptions) Complete() error {
	w.config = &ServerConfig{
		CertDir: w.CertDir,
	}

	return nil
}

// Completed returns the completed ServerConfig. Only call this if `Complete` was successful.
func (w *ServerOptions) Completed() *ServerConfig {
	return w.config
}

// AddFlags implements Flagger.AddFlags.
// TODO: (timuthy) This flag can be removed and added to ServerOptions as soons as we use Controller-Runtime v0.2.2
// https://github.com/kubernetes-sigs/controller-runtime/pull/569
func (w *ServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&w.CertDir, CertDirFlag, w.CertDir, "The directory that contains the webhook server key and certificate.")
}
