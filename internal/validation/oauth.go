package validation

import (
	"encoding/json"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

type oauth struct{}

func (o *oauth) Validate(config *runtime.RawExtension) error {
	template := &gatewayv2alpha1.OauthModeConfig{}

	if !configNotEmpty(config) {
		return fmt.Errorf("Config empty!")
	}

	//Check if the supplied data is castable to OauthModeConfig
	err := json.Unmarshal(config.Raw, &template.Paths)
	if err != nil {
		return errors.WithStack(err)
	}
	// If not, the result is an empty template object.
	// Check if template is empty
	if len(template.Paths) == 0 {
		return fmt.Errorf("Supplied config does not match internal template!\n%v", len(template.Paths))
	}

	return nil
}
