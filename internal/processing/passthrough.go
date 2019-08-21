package processing

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis/istio/common/v1alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type passthrough struct {
	client.Client
}

func (p *passthrough) Process(api *gatewayv2alpha1.Api) error {
	fmt.Println("Processing API")

	//1. get VS

	//2. create VS if needed

	return p.createVirtualService(api)
}

func (p *passthrough) createVirtualService(api *gatewayv2alpha1.Api) error {
	var virtualServiceName string
	virtualServiceName = fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)

	objectMeta := k8sMeta.ObjectMeta{
		Name:      virtualServiceName,
		Namespace: api.ObjectMeta.Namespace,
	}

	match := &networkingv1alpha3.HTTPMatchRequest{
		URI: &v1alpha1.StringMatch{
			Regex: "/.*",
		},
	}
	route := &networkingv1alpha3.HTTPRouteDestination{
		Destination: networkingv1alpha3.Destination{
			Host: fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace),
			Port: networkingv1alpha3.PortSelector{
				Number: uint32(*api.Spec.Service.Port),
			},
		},
	}

	spec := &networkingv1alpha3.VirtualServiceSpec{
		Hosts:    []string{*api.Spec.Service.HostURL},
		Gateways: []string{"kyma-gateway"},
		HTTP: []networkingv1alpha3.HTTPRoute{
			networkingv1alpha3.HTTPRoute{
				Match: []networkingv1alpha3.HTTPMatchRequest{*match},
				Route: []networkingv1alpha3.HTTPRouteDestination{*route},
			},
		},
	}

	vs := &networkingv1alpha3.VirtualService{
		ObjectMeta: objectMeta,
		Spec:       *spec,
	}

	return p.Client.Create(context.TODO(), vs)
}
