package clients

import (
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	oryClient "github.com/kyma-incubator/api-gateway/internal/clients/ory"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(crClient client.Client) *ExternalCRClients {
	return &ExternalCRClients{
		virtualService:       istioClient.ForVirtualService(crClient),
		authenticationPolicy: istioClient.ForAuthenticationPolicy(crClient),
		accessRule:           oryClient.ForAccessRule(crClient),
	}
}

//Exposes clients for external Custom Resources (e.g. Istio VirtualService)
type ExternalCRClients struct {
	virtualService       *istioClient.VirtualService
	authenticationPolicy *istioClient.AuthenticationPolicy
	accessRule           *oryClient.AccessRule
}

func (c *ExternalCRClients) ForVirtualService() *istioClient.VirtualService {
	return c.virtualService
}

func (c *ExternalCRClients) ForAuthenticationPolicy() *istioClient.AuthenticationPolicy {
	return c.authenticationPolicy
}

func (c *ExternalCRClients) ForAccessRule() *oryClient.AccessRule {
	return c.accessRule
}
