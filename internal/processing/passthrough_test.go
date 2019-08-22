package processing_test

import (
	"testing"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

// spec:
//   service:
//     host: imgur.com
//     name: imgur
//     port: 443
//   auth:
//     name: PASSTHROUGH
//   gateway: kyma-gateway.kyma-system.svc.cluster.local

func TestGenerateVirtualService(t *testing.T) {
	exampleAPI := &gatewayv2alpha1.Api{}
}
