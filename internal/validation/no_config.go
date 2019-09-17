package validation

import (
	"github.com/ory/oathkeeper-maester/api/v1alpha1"
)

//noConfig is an accessStrategy validator that does not accept nested config
type noConfigAccStrValidator struct{}

func (a *noConfigAccStrValidator) Validate(attrPath string, handler *v1alpha1.Handler) []Failure {
	var problems []Failure

	if handler.Config != nil && len(handler.Config.Raw) > 0 {
		problems = append(problems, Failure{AttributePath: attrPath, Message: "strategy: " + handler.Name + " does not support configuration"})
	}

	return problems
}
