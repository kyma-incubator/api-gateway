package validation

import "k8s.io/apimachinery/pkg/runtime"

type oauth struct{}

func (o *oauth) Validate(attrPath string, accStrConfig *runtime.RawExtension) []Failure {
	return nil
}
