package validation_test

import (
	"testing"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	"gotest.tools/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	validYaml    = ``
	notValidYaml = `
config:
  foo: bar
`
)

func TestPassthroughValidate(t *testing.T) {
	factory := validation.NewValidationStrategyFactory()
	strategy, err := factory.NewValidationStrategy(gatewayv2alpha1.PASSTHROUGH)
	assert.NilError(t, err)

	valid := &runtime.RawExtension{Raw: []byte(validYaml)}
	assert.NilError(t, strategy.Validate(valid))

	notValid := &runtime.RawExtension{Raw: []byte(notValidYaml)}
	assert.Error(t, strategy.Validate(notValid), "passthrough mode requires empty configuration")
}
