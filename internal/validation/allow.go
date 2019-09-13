package validation

import (
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

//Allow is used to validate accessStrategy of type gatewayv2alpha1.Allow
type Allow struct{}

//Validate performs the validation
func (a *Allow) Validate(gate *gatewayv2alpha1.Gate) error {
	if len(gate.Spec.Rules) != 1 {
		return fmt.Errorf("supplied config should contain exactly one path")
	}
	if hasDuplicates(gate.Spec.Rules) {
		return fmt.Errorf("supplied config is invalid: multiple definitions of the same path detected")
	}
	return nil
}
