//go:generate mockgen -package=client -destination=mocks.go github.com/gardener/gardener-extensions/controllers/provider-gcp/pkg/internal/client Interface,FirewallsService,FirewallsListCall,FirewallsDeleteCall

package client
