package validation

import (
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

//configNotEmpty Verify if the config object is not empty
func configNotEmpty(config *runtime.RawExtension) bool {
	if config == nil {
		return false
	}
	return len(config.Raw) != 0
}

//Validator is used to validate github.com/kyma-incubator/api-gateway/api/v2alpha1/Gate instances
type Validator struct {
}

//Validate performs Gate validation
func (v *Validator) Validate(gate *gatewayv2alpha1.Gate) []Failure {

	res := []Failure{}
	//Validate service
	res = append(res, v.validateService(gate.Spec.Service)...)
	//Validate Gateway
	res = append(res, v.validateGateway(gate.Spec.Gateway)...)
	//Validate Rules
	res = append(res, v.validateRules(gate.Spec.Rules)...)

	return res
}

//Failure carries validation failures for a single attribute of an object.
type Failure struct {
	AttributePath string
	Message       string
}

func (v *Validator) validateService(service *gatewayv2alpha1.Service) []Failure {
	return nil
}

func (v *Validator) validateGateway(gateway *string) []Failure {
	return nil
}

func (v *Validator) validateRules(rules []gatewayv2alpha1.Rule) []Failure {
	var problems []Failure
	if len(rules) == 0 {
		problems = append(problems, Failure{AttributePath: ".rules", Message: "No rules defined"})
		return problems
	}

	for i, r := range rules {
		if len(r.Methods) == 0 {
			attributePath := fmt.Sprintf(".rules[%d]", i)
			problems = append(problems, Failure{AttributePath: attributePath, Message: "No methods defined"})
		}
	}

	return problems
}
