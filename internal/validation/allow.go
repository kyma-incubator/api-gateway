package validation

import (
	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

//allow is used to validate accessStrategy of type gatewayv2alpha1.Allow
type allow struct{}

func (a *allow) Validate(attrPath string, accStrConfig *runtime.RawExtension) []Failure {
	var problems []Failure

	if len(accStrConfig.Raw) > 0 {
		problems = append(problems, Failure{AttributePath: attrPath, Message: "strategy: " + gatewayv2alpha1.Allow + " does not support configuration"})
	}

	return problems
}
