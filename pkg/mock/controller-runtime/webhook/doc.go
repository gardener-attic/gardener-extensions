//go:generate mockgen -package=webhook -destination=mocks.go sigs.k8s.io/controller-runtime/pkg/webhook Webhook

package webhook
