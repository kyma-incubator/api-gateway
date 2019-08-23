package processing

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

type factory struct {
	Client client.Client
}

type ProcessingStrategy interface {
	Process(api *gatewayv2alpha1.Api) error
}

func NewProcessingStrategyFactory(client client.Client) *factory {
	return &factory{Client: client}
}

func (f *factory) NewProcessingStrategy(strategyName string) (ProcessingStrategy, error) {
	switch strategyName {
	case gatewayv2alpha1.PASSTHROUGH:
		fmt.Println("PASSTHROUGH mode detected")
		return &passthrough{Client: f.Client}, nil
	default:
		err := fmt.Errorf("Unsupported mode: %s", strategyName)
		return nil, err
	}
}
