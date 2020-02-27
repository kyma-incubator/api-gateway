package validation

import (
	"fmt"
	"regexp"
)

const (
	labelKeyRegexpString   = "^[a-zA-Z0-9][-A-Za-z0-9_./]{0,61}[a-zA-Z0-9]$"
	labelValueRegexpString = "^[a-zA-Z0-9][-A-Za-z0-9_.]{0,61}[a-zA-Z0-9]$"
)

var (
	labelKeyRegexp   = regexp.MustCompile(labelKeyRegexpString)
	labelValueRegexp = regexp.MustCompile(labelValueRegexpString)
)

//VerifyLabelKey returns error if the provided string is not a proper k8s label key
func VerifyLabelKey(key string) error {
	if !labelKeyRegexp.MatchString(key) {
		return fmt.Errorf("key '%s' is not a proper k8s label key", key)
	}
	return nil
}

//VerifyLabelValue returns error if the provided string is not a proper k8s label value
func VerifyLabelValue(value string) error {
	if !labelValueRegexp.MatchString(value) {
		return fmt.Errorf("value '%s' is not a proper k8s label value", value)
	}
	return nil
}
