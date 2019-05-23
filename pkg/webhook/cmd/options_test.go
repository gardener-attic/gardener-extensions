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
	mockwebhook "github.com/gardener/gardener-extensions/pkg/mock/controller-runtime/webhook"
	mockextensionswebhook "github.com/gardener/gardener-extensions/pkg/mock/gardener-extensions/webhook"
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
			It("should yield correct ServerConfig after completion in service mode", func() {
				command := test.NewCommandBuilder(name).
					Flags(
						test.IntFlag(PortFlag, port),
						test.StringFlag(CertDirFlag, certDir),
						test.StringFlag(ModeFlag, ServiceMode),
						test.StringFlag(NameFlag, name),
						test.StringFlag(NamespaceFlag, namespace),
						test.StringFlag(ServiceSelectorsFlag, serviceSelectors),
					).
					Command().
					Slice()
				fs := pflag.NewFlagSet(name, pflag.ExitOnError)
				opts := ServerOptions{}

				// Parse command into options
				opts.AddFlags(fs)
				err := fs.Parse(command)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts).To(Equal(ServerOptions{
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
				Expect(opts.Completed()).To(Equal(&ServerConfig{
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
			It("should yield correct ServerConfig after completion in url mode", func() {
				command := test.NewCommandBuilder(name).
					Flags(
						test.IntFlag(PortFlag, port),
						test.StringFlag(CertDirFlag, certDir),
						test.StringFlag(ModeFlag, URLMode),
						test.StringFlag(NameFlag, name),
						test.StringFlag(HostFlag, host),
					).
					Command().
					Slice()
				fs := pflag.NewFlagSet(name, pflag.ExitOnError)
				opts := ServerOptions{}

				// Parse command into options
				opts.AddFlags(fs)
				err := fs.Parse(command)
				Expect(err).NotTo(HaveOccurred())
				Expect(opts).To(Equal(ServerOptions{
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
				Expect(opts.Completed()).To(Equal(&ServerConfig{
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

	Context("SwitchOptions", func() {
		const commandName = "test"

		Describe("#AddFlags", func() {
			It("should correctly parse the flags", func() {
				var (
					name1    = "foo"
					name2    = "bar"
					switches = NewSwitchOptions(
						Switch(name1, nil),
						Switch(name2, nil),
					)
				)

				fs := pflag.NewFlagSet(commandName, pflag.ContinueOnError)
				switches.AddFlags(fs)

				err := fs.Parse(test.NewCommandBuilder(commandName).
					Flags(
						test.StringSliceFlag(DisableFlag, name1, name2),
					).
					Command().
					Slice())

				Expect(err).NotTo(HaveOccurred())
				Expect(switches.Complete()).To(Succeed())

				Expect(switches.Disabled).To(Equal([]string{name1, name2}))
			})

			It("should error on an unknown webhook", func() {
				switches := NewSwitchOptions()

				fs := pflag.NewFlagSet(commandName, pflag.ContinueOnError)
				switches.AddFlags(fs)

				err := fs.Parse(test.NewCommandBuilder(commandName).
					Flags(
						test.StringSliceFlag(DisableFlag, "unknown"),
					).
					Command().
					Slice())

				Expect(err).NotTo(HaveOccurred())
				Expect(switches.Complete()).To(HaveOccurred())
			})
		})

		Describe("#AddToManager", func() {
			It("should return a configuration that does not add the disabled webhooks", func() {
				var (
					f1 = mockextensionswebhook.NewMockFactory(ctrl)
					f2 = mockextensionswebhook.NewMockFactory(ctrl)

					name1 = "name1"
					name2 = "name2"

					switches = NewSwitchOptions(
						Switch(name1, f1.Do),
						Switch(name2, f2.Do),
					)

					wh1 = mockwebhook.NewMockWebhook(ctrl)
				)

				f1.EXPECT().Do(nil).Return(wh1, nil)

				switches.Disabled = []string{name2}

				Expect(switches.Complete()).To(Succeed())

				webhooks, err := switches.Completed().WebhooksFactory(nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(webhooks).To(Equal([]webhook.Webhook{wh1}))
			})
		})
	})
})
