package validation

import (
	"fmt"
	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

type factory struct {}

type Factory interface {
	NewValidationStrategy(strategyName string) (ValidationStrategy, error)
}

type ValidationStrategy interface {
	Validate(config *runtime.RawExtension) error
}

func NewValidationStrategyFactory() Factory {
	return &factory{}
}

func (f *factory) NewValidationStrategy(strategyName string) (ValidationStrategy, error) {
	switch strategyName {
	case gatewayv2alpha1.PASSTHROUGH:
		fmt.Println("PASSTHROUGH mode detected")
		return &passthrough{}, nil
	case gatewayv2alpha1.JWT:
		fmt.Println("JWT mode detected")
		return &jwt{}, nil
	case gatewayv2alpha1.OAUTH:
		fmt.Println("OAUTH mode detected")
		return &oauth{}, nil
	default:
		err := fmt.Errorf("Unsupported mode: %s", strategyName)
		return nil, err
	}
}

type passthrough struct {}

func (p *passthrough) Validate(config *runtime.RawExtension) error {
	if config != nil {
		return fmt.Errorf("passthrough mode requires empty configuration")
	}
	return nil
}

type oauth struct {}

func (o *oauth) Validate(config *runtime.RawExtension) error {
	return nil
}

type jwt struct {}

func (j *jwt) Validate(config *runtime.RawExtension) error {
	return nil
}
