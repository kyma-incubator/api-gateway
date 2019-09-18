package validation

import (
	"bytes"
	"fmt"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/ory/oathkeeper-maester/api/v1alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

//Validators for AccessStrategies
var vldNoConfig = &noConfigAccStrValidator{}
var vldJWT = &jwtAccStrValidator{}
var vldDummy = &dummyAccStrValidator{}

type accessStrategyValidator interface {
	Validate(attrPath string, Handler *v1alpha1.Handler) []Failure
}

//configNotEmpty Verify if the config object is not empty
func configEmpty(config *runtime.RawExtension) bool {

	return config == nil ||
		len(config.Raw) == 0 ||
		bytes.Equal(config.Raw, []byte("null")) ||
		bytes.Equal(config.Raw, []byte("{}"))
}

//configNotEmpty Verify if the config object is not empty
func configNotEmpty(config *runtime.RawExtension) bool {
	return !configEmpty(config)
}

//APIRule is used to validate github.com/kyma-incubator/api-gateway/api/v1alpha1/APIRule instances
type APIRule struct {
}

//Validate performs Gate validation
func (v *APIRule) Validate(api *gatewayv1alpha1.APIRule) []Failure {

	res := []Failure{}
	//Validate service
	res = append(res, v.validateService(".spec.service", api.Spec.Service)...)
	//Validate Gateway
	res = append(res, v.validateGateway(".spec.gateway", api.Spec.Gateway)...)
	//Validate Rules
	res = append(res, v.validateRules(".spec.rules", api.Spec.Rules)...)

	return res
}

//Failure carries validation failures for a single attribute of an object.
type Failure struct {
	AttributePath string
	Message       string
}

func (v *APIRule) validateService(attributePath string, service *gatewayv1alpha1.Service) []Failure {
	return nil
}

func (v *APIRule) validateGateway(attributePath string, gateway *string) []Failure {
	return nil
}

func (v *APIRule) validateRules(attributePath string, rules []gatewayv1alpha1.Rule) []Failure {
	var problems []Failure

	if len(rules) == 0 {
		problems = append(problems, Failure{AttributePath: attributePath, Message: "No rules defined"})
		return problems
	}

	if hasDuplicates(rules) {
		problems = append(problems, Failure{AttributePath: attributePath, Message: "multiple rules defined for the same path"})
	}

	for i, r := range rules {
		attrPath := fmt.Sprintf("%s[%d]", attributePath, i)
		problems = append(problems, v.validateMethods(attrPath+".methods", r.Methods)...)
		problems = append(problems, v.validateAccessStrategies(attrPath+".accessStrategies", r.AccessStrategies)...)
	}

	return problems
}

func (v *APIRule) validateMethods(attributePath string, methods []string) []Failure {
	return nil
}

func (v *APIRule) validateAccessStrategies(attributePath string, accessStrategies []*rulev1alpha1.Authenticator) []Failure {
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

func (v *APIRule) validateAccessStrategy(attributePath string, accessStrategy *rulev1alpha1.Authenticator) []Failure {
	var problems []Failure

	var vld accessStrategyValidator

	switch accessStrategy.Handler.Name {
	case "allow": //our internal constant, does not exist in ORY
		vld = vldNoConfig
	case "noop":
		vld = vldNoConfig
	case "unauthorized":
		vld = vldNoConfig
	case "anonymous":
		vld = vldNoConfig
	case "cookie_session":
		vld = vldNoConfig
	case "oauth2_client_credentials":
		vld = vldDummy
	case "oauth2_introspection":
		vld = vldDummy
	case "jwt":
		vld = vldJWT
	default:
		problems = append(problems, Failure{AttributePath: attributePath + ".handler", Message: fmt.Sprintf("Unsupported accessStrategy: %s", accessStrategy.Handler.Name)})
		return problems
	}

	return vld.Validate(attributePath, accessStrategy.Handler)
}
