package validation

import (
	"encoding/json"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

//jwt is used to validate accessStrategy of type gatewayv2alpha1.Jwt
type jwt struct{}

func (j *jwt) Validate(attributePath string, accStrConfig *runtime.RawExtension) []Failure {
	var problems []Failure

	var template gatewayv2alpha1.JWTModeConfig

	if !configNotEmpty(accStrConfig) {
		problems = append(problems, Failure{AttributePath: attributePath, Message: "supplied config cannot be empty"})
		return problems
	}
	err := json.Unmarshal(accStrConfig.Raw, &template)
	if err != nil {
		problems = append(problems, Failure{AttributePath: attributePath, Message: "Can't read json: " + err.Error()})
		return problems
	}
	if !isValidURL(template.Issuer) {
		problems = append(problems, Failure{AttributePath: attributePath + ".issuer", Message: "issuer field is empty or not a valid url"})
	}
	return problems
}
