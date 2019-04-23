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

package infrastructure

import (
	"context"
	"fmt"

	mockgcpclient "github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/mock/client"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/api/compute/v1"
)

var _ = Describe("Infrastructure", func() {
	var (
		ctrl *gomock.Controller
	)
	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
	})
	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("#ListKubernetesFirewalls", func() {
		It("should list all kubernetes related firewall names", func() {
			var (
				ctx       = context.TODO()
				projectID = "foo"
				network   = "bar"

				firewallName  = fmt.Sprintf("%sfw", KubernetesFirewallNamePrefix)
				firewallNames = []string{firewallName}

				client            = mockgcpclient.NewMockInterface(ctrl)
				firewalls         = mockgcpclient.NewMockFirewallsService(ctrl)
				firewallsListCall = mockgcpclient.NewMockFirewallsListCall(ctrl)
			)

			gomock.InOrder(
				client.EXPECT().Firewalls().Return(firewalls),
				firewalls.EXPECT().List(projectID).Return(firewallsListCall),
				firewallsListCall.EXPECT().Pages(ctx, gomock.AssignableToTypeOf(func(*compute.FirewallList) error { return nil })).
					DoAndReturn(func(_ context.Context, f func(*compute.FirewallList) error) error {
						return f(&compute.FirewallList{
							Items: []*compute.Firewall{
								{Name: firewallName, Network: network},
							},
						})
					}),
			)

			actual, err := ListKubernetesFirewalls(ctx, client, projectID, network)

			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(firewallNames))
		})
	})

	Describe("#ListKubernetesRoutes", func() {
		It("should list all kubernetes related route names with the shoot namespace as prefix", func() {
			var (
				ctx       = context.TODO()
				projectID = "foo"
				network   = "shoot--foo--bar"

				routeName  = fmt.Sprintf("shoot--foo--bar-2690fa98-450f-11e9-8ebe-ce2a79d67b14")
				routeNames = []string{routeName}

				client         = mockgcpclient.NewMockInterface(ctrl)
				routes         = mockgcpclient.NewMockRoutesService(ctrl)
				routesListCall = mockgcpclient.NewMockRoutesListCall(ctrl)
			)

			gomock.InOrder(
				client.EXPECT().Routes().Return(routes),
				routes.EXPECT().List(projectID).Return(routesListCall),
				routesListCall.EXPECT().Pages(ctx, gomock.AssignableToTypeOf(func(*compute.RouteList) error { return nil })).
					DoAndReturn(func(_ context.Context, f func(*compute.RouteList) error) error {
						return f(&compute.RouteList{
							Items: []*compute.Route{
								{Name: routeName, Network: network},
							},
						})
					}),
			)

			actual, err := ListKubernetesRoutes(ctx, client, projectID, network)

			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(routeNames))
		})
	})

	Describe("#DeleteFirewalls", func() {
		It("should delete all firewalls", func() {
			var (
				ctx       = context.TODO()
				projectID = "foo"

				firewallName  = fmt.Sprintf("%sfw", KubernetesFirewallNamePrefix)
				firewallNames = []string{firewallName}

				client              = mockgcpclient.NewMockInterface(ctrl)
				firewalls           = mockgcpclient.NewMockFirewallsService(ctrl)
				firewallsDeleteCall = mockgcpclient.NewMockFirewallsDeleteCall(ctrl)
			)

			gomock.InOrder(
				client.EXPECT().Firewalls().Return(firewalls),
				firewalls.EXPECT().Delete(projectID, firewallName).Return(firewallsDeleteCall),
				firewallsDeleteCall.EXPECT().Context(ctx).Return(firewallsDeleteCall),
				firewallsDeleteCall.EXPECT().Do(),
			)

			Expect(DeleteFirewalls(ctx, client, projectID, firewallNames)).To(Succeed())
		})
	})

	Describe("#DeleteRoutes", func() {
		It("should delete all routess", func() {
			var (
				ctx       = context.TODO()
				projectID = "foo"

				routeName  = fmt.Sprintf("shoot--foo--bar-2690fa98-450f-11e9-8ebe-ce2a79d67b14")
				routeNames = []string{routeName}

				client           = mockgcpclient.NewMockInterface(ctrl)
				routes           = mockgcpclient.NewMockRoutesService(ctrl)
				routesDeleteCall = mockgcpclient.NewMockRoutesDeleteCall(ctrl)
			)

			gomock.InOrder(
				client.EXPECT().Routes().Return(routes),
				routes.EXPECT().Delete(projectID, routeName).Return(routesDeleteCall),
				routesDeleteCall.EXPECT().Context(ctx).Return(routesDeleteCall),
				routesDeleteCall.EXPECT().Do(),
			)

			Expect(DeleteRoutes(ctx, client, projectID, routeNames)).To(Succeed())
		})
	})
})
