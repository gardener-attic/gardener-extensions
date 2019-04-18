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

package cmd

import (
	"testing"

	"github.com/gardener/gardener-extensions/pkg/util/test"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webhook Cmd Suite")
}

var _ = Describe("Options", func() {
	var (
		ctrl *gomock.Controller
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("WebhookConfigOptions", func() {
		const (
			name             = "foo"
			port             = 9999
			certDir          = "/cert"
			namespace        = "default"
			serviceSelectors = `{"app":"kubernetes"}`
			host             = "bar"
		)

		Describe("#Completed", func() {
			It("should yield correct WebhookServerConfig after completion in service mode", func() {
				command := test.NewCommandBuilder(name).
					Flag(PortFlag, port).
					Flag(CertDirFlag, certDir).
					Flag(ModeFlag, ServiceMode).
					Flag(NameFlag, name).
					Flag(NamespaceFlag, namespace).
					Flag(ServiceSelectorsFlag, serviceSelectors).
					Command().
					Slice()
				fs := pflag.NewFlagSet(name, pflag.ExitOnError)
				opts := WebhookServerOptions{}

				// Parse command into options
				opts.AddFlags(fs)
				err := fs.Parse(command)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts).To(Equal(WebhookServerOptions{
					Port:             port,
					CertDir:          certDir,
					Mode:             ServiceMode,
					Name:             name,
					Namespace:        namespace,
					ServiceSelectors: serviceSelectors,
				}))

				// Complete the options
				err = opts.Complete()
				Expect(err).NotTo(HaveOccurred())

				// Check Completed result
				Expect(opts.Completed()).To(Equal(&WebhookServerConfig{
					Port:    port,
					CertDir: certDir,
					BootstrapOptions: &webhook.BootstrapOptions{
						MutatingWebhookConfigName: name,
						Secret:                    &types.NamespacedName{Namespace: namespace, Name: name},
						Service: &webhook.Service{
							Name:      name,
							Namespace: namespace,
							Selectors: map[string]string{"app": "kubernetes"},
						},
					},
				}))
			})
		})

		Describe("#Completed", func() {
			It("should yield correct WebhookServerConfig after completion in url mode", func() {
				command := test.NewCommandBuilder(name).
					Flag(PortFlag, port).
					Flag(CertDirFlag, certDir).
					Flag(ModeFlag, URLMode).
					Flag(NameFlag, name).
					Flag(HostFlag, host).
					Command().
					Slice()
				fs := pflag.NewFlagSet(name, pflag.ExitOnError)
				opts := WebhookServerOptions{}

				// Parse command into options
				opts.AddFlags(fs)
				err := fs.Parse(command)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts).To(Equal(WebhookServerOptions{
					Port:    port,
					CertDir: certDir,
					Mode:    URLMode,
					Name:    name,
					Host:    host,
				}))

				// Complete the options
				err = opts.Complete()
				Expect(err).NotTo(HaveOccurred())

				// Check Completed result
				h := host
				Expect(opts.Completed()).To(Equal(&WebhookServerConfig{
					Port:    port,
					CertDir: certDir,
					BootstrapOptions: &webhook.BootstrapOptions{
						MutatingWebhookConfigName: name,
						Host:                      &h,
					},
				}))
			})
		})
	})
})
