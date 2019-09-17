package validation

import (
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

//All known validators for AccessStrategies
var vldAllow = &allow{}
var vldJWT = &jwt{}
var vldOAuth = &oauth{}

type accessStrategyValidator interface {
	Validate(attrPath string, accStrConfig *runtime.RawExtension) []Failure
}

//configNotEmpty Verify if the config object is not empty
func configNotEmpty(config *runtime.RawExtension) bool {
	if config == nil {
		return false
	}
	return len(config.Raw) != 0
}

//Gate is used to validate github.com/kyma-incubator/api-gateway/api/v2alpha1/Gate instances
type Gate struct {
}

//Validate performs Gate validation
func (v *Gate) Validate(gate *gatewayv2alpha1.Gate) []Failure {

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

func (v *Gate) validateService(service *gatewayv2alpha1.Service) []Failure {
	return nil
}

func (v *Gate) validateGateway(gateway *string) []Failure {
	return nil
}

func (v *Gate) validateRules(rules []gatewayv2alpha1.Rule) []Failure {
	var problems []Failure

	if len(rules) == 0 {
		problems = append(problems, Failure{AttributePath: ".rules", Message: "No rules defined"})
		return problems
	}

	if hasDuplicates(rules) {
		problems = append(problems, Failure{AttributePath: ".rules", Message: "multiple rules defined for the same path"})
	}

	for i, r := range rules {
		attributePath := fmt.Sprintf(".rules[%d]", i)
		problems = append(problems, v.validateMethods(attributePath+".methods", r.Methods)...)
		problems = append(problems, v.validateAccessStrategies(attributePath+".accessStrategies", r.AccessStrategies)...)
	}

	return problems
}

func (v *Gate) validateMethods(attributePath string, methods []string) []Failure {
	var problems []Failure

	if len(methods) == 0 {
		problems = append(problems, Failure{AttributePath: attributePath, Message: "No methods defined for rule"})
	}

	return problems
}

func (v *Gate) validateAccessStrategies(attributePath string, accessStrategies []*rulev1alpha1.Authenticator) []Failure {
	var problems []Failure

	if len(accessStrategies) == 0 {
		problems = append(problems, Failure{AttributePath: attributePath, Message: "No accessStrategies defined"})
		return problems
	}

	for i, r := range accessStrategies {
		strategyAttrPath := attributePath + fmt.Sprintf("[%d]", i)
		problems = append(problems, v.validateAccessStrategy(strategyAttrPath, r)...)
	}

	return problems
}

func (v *Gate) validateAccessStrategy(attributePath string, accessStrategy *rulev1alpha1.Authenticator) []Failure {
	var problems []Failure

	var vld accessStrategyValidator

	switch accessStrategy.Handler.Name {
	case gatewayv2alpha1.Allow:
		vld = vldAllow
	case gatewayv2alpha1.Jwt:
		vld = vldJWT
	case gatewayv2alpha1.Oauth:
		vld = vldOAuth
	default:
		problems = append(problems, Failure{AttributePath: attributePath, Message: fmt.Sprintf("Unsupported accessStrategy: %s", accessStrategy.Handler.Name)})
		return problems
	}

	return vld.Validate(attributePath+".config", accessStrategy.Handler.Config)
}
