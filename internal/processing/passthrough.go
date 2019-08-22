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
	vs := p.createVirtualService(api)

	return p.saveVirtualService(vs)
}

func (p *passthrough) getVirtualService(api *gatewayv2alpha1.Api) error {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	namespacedName := client.ObjectKey{Namespace: api.GetNamespace(), Name: virtualServiceName}

	vs := &networkingv1alpha3.VirtualService{}

	p.Client.Get(context.TODO(), namespacedName, vs)

	fmt.Printf("\nVS: %v\n", *vs)

	return nil
}

func (p *passthrough) saveVirtualService(vs *networkingv1alpha3.VirtualService) error {
	return p.Client.Create(context.TODO(), vs)
}

func (p *passthrough) createVirtualService(api *gatewayv2alpha1.Api) *networkingv1alpha3.VirtualService {
	var virtualServiceName string
	virtualServiceName = fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	var controller bool
	controller = true

	ownerRef := &k8sMeta.OwnerReference{
		Name:       api.ObjectMeta.Name,
		APIVersion: api.TypeMeta.APIVersion,
		Kind:       api.TypeMeta.Kind,
		UID:        api.ObjectMeta.UID,
		Controller: &controller,
	}

	objectMeta := k8sMeta.ObjectMeta{
		Name:            virtualServiceName,
		Namespace:       api.ObjectMeta.Namespace,
		OwnerReferences: []k8sMeta.OwnerReference{*ownerRef},
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
		Hosts:    []string{*api.Spec.Service.Host},
		Gateways: []string{*api.Spec.Gateway},
		HTTP: []networkingv1alpha3.HTTPRoute{
			{
				Match: []networkingv1alpha3.HTTPMatchRequest{*match},
				Route: []networkingv1alpha3.HTTPRouteDestination{*route},
			},
		},
	}

	vs := &networkingv1alpha3.VirtualService{
		ObjectMeta: objectMeta,
		Spec:       *spec,
	}

	return vs
}
