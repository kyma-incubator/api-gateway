package clients

import (
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(crClient client.Client) *ExternalCRD {
	return &ExternalCRD{
		virtualService: istioClient.ForVirtualService(crClient),
	}
}

//Exposes clients for external CRDs (e.g. Istio VirtualService)
type ExternalCRD struct {
	virtualService *istioClient.VirtualService
}

func (c *ExternalCRD) ForVirtualService() *istioClient.VirtualService {
	return c.virtualService
}
