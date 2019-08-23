package processing

import (
	"testing"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	apiName                 = "some-api"
	apiUID        types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	apiNamespace            = "some-namespace"
	apiAPIVersion           = "gateway.kyma-project.io/v2alpha1"
	apiKind                 = "Gate"
	apiGateway              = "some-gateway"
	serviceName             = "example-service"
	serviceHost             = "myService.myDomain.com"
	servicePort   int32     = 8080
)

func TestGenerateVirtualService(t *testing.T) {
	exampleAPI := &gatewayv2alpha1.Gate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiName,
			UID:       apiUID,
			Namespace: apiNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiAPIVersion,
			Kind:       apiKind,
		},
		Spec: gatewayv2alpha1.GateSpec{
			Gateway: &apiGateway,
			Service: &gatewayv2alpha1.Service{
				Name: &serviceName,
				Host: &serviceHost,
				Port: &servicePort,
			},
		},
	}
	strategyPassthrough := &passthrough{}
	vs := strategyPassthrough.generateVirtualService(exampleAPI)

	assert.Equal(t, len(vs.Spec.Gateways), 1)
	assert.Equal(t, vs.Spec.Gateways[0], apiGateway)

	assert.Equal(t, len(vs.Spec.Hosts), 1)
	assert.Equal(t, vs.Spec.Hosts[0], serviceHost)

	assert.Equal(t, len(vs.Spec.HTTP), 1)
	assert.Equal(t, len(vs.Spec.HTTP[0].Route), 1)
	assert.Equal(t, len(vs.Spec.HTTP[0].Match), 1)
	assert.Equal(t, vs.Spec.HTTP[0].Route[0].Destination.Host, serviceName+"."+apiNamespace+".svc.cluster.local")
	assert.Equal(t, vs.Spec.HTTP[0].Route[0].Destination.Port.Number, uint32(servicePort))
	assert.Equal(t, vs.Spec.HTTP[0].Match[0].URI.Regex, "/.*")

	assert.Equal(t, vs.ObjectMeta.Name, apiName+"-"+serviceName)
	assert.Equal(t, vs.ObjectMeta.Namespace, apiNamespace)

	assert.Equal(t, vs.ObjectMeta.OwnerReferences[0].APIVersion, apiAPIVersion)
	assert.Equal(t, vs.ObjectMeta.OwnerReferences[0].Kind, apiKind)
	assert.Equal(t, vs.ObjectMeta.OwnerReferences[0].Name, apiName)
	assert.Equal(t, vs.ObjectMeta.OwnerReferences[0].UID, apiUID)

}
