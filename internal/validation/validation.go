package validation

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	nilTemplate OauthModeConfig
)

//OauthModeConfig ?
type OauthModeConfig struct {
	Foo *string `json:"foo"`
}

//ValidatePassthoughMode ?
func ValidatePassthroughMode(config *runtime.RawExtension) error {
	if config != nil {
		return fmt.Errorf("passthrough mode requires empty configuration")
	}
	return nil
}

func ValidateOauthMode(config *runtime.RawExtension) error {
	template := &OauthModeConfig{}

	if config == nil {
		return fmt.Errorf("Config empty!")
	}
	if len(config.Raw) == 0 {
		return fmt.Errorf("Config empty!")
	}

	//Check if the supplied data is castable to OauthModeConfig
	err := json.Unmarshal(config.Raw, template)
	if err != nil {
		return errors.WithStack(err)
	}
	// If not, the result is an empty template object.
	// Check if template is empty
	if *template == nilTemplate {
		return fmt.Errorf("Supplied config does not match internal template!")
	}
	return nil
}
