package processing

import (
	"context"
	"fmt"
	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
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

	oldVS, err := p.getVirtualService(api)
	if err != nil {
		return err
	}

	vs := p.generateVirtualService(api)

	if oldVS != nil {
		newVS := p.prepareVirtualService(api, oldVS)
		return p.updateVirtualService(newVS)
	} else {
		return p.createVirtualService(vs)
	}

	return nil
}

func (p *passthrough) getVirtualService(api *gatewayv2alpha1.Api) (*networkingv1alpha3.VirtualService, error) {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	namespacedName := client.ObjectKey{Namespace: api.GetNamespace(), Name: virtualServiceName}
	vs := &networkingv1alpha3.VirtualService{}

	err := p.Client.Get(context.TODO(), namespacedName, vs)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return vs, nil
}

func (p *passthrough) createVirtualService(vs *networkingv1alpha3.VirtualService) error {
	return p.Client.Create(context.TODO(), vs)
}

func (p * passthrough) prepareVirtualService (api *gatewayv2alpha1.Api, vs *networkingv1alpha3.VirtualService) *networkingv1alpha3.VirtualService{
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

	vs.ObjectMeta.OwnerReferences = []k8sMeta.OwnerReference{*ownerRef}
	vs.ObjectMeta.Name = virtualServiceName
	vs.ObjectMeta.Namespace = api.ObjectMeta.Namespace

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

	vs.Spec = *spec

	return vs

}

func (p *passthrough) updateVirtualService(vs *networkingv1alpha3.VirtualService) error {
	return p.Client.Update(context.TODO(), vs)
}

func (p *passthrough) generateVirtualService(api *gatewayv2alpha1.Api) *networkingv1alpha3.VirtualService {
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
