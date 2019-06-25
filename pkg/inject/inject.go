package inject

import (
	"context"
	"github.com/gardener/gardener-extensions/pkg/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type WithClient struct {
	Client client.Client
}

func (w *WithClient) InjectClient(c client.Client) error {
	w.Client = c
	return nil
}

type WithEmbeddedClient struct {
	client.Client
}

func (w *WithEmbeddedClient) InjectClient(c client.Client) error {
	w.Client = c
	return nil
}

type WithStopChannel struct {
	StopChannel <-chan struct{}
}

func (w *WithStopChannel) InjectStopChannel(stopChan <-chan struct{}) error {
	w.StopChannel = stopChan
	return nil
}

type WithContext struct {
	Context context.Context
}

func (w *WithContext) InjectStopChannel(stopChan <-chan struct{}) error {
	w.Context = util.ContextFromStopChannel(stopChan)
	return nil
}
