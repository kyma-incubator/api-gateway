package processing

import (
	"context"
	"fmt"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	k8sMeta "k8s.io/apimachinery/pkg/apis/meta/v1"
	authenticationv1alpha1 "knative.dev/pkg/apis/istio/authentication/v1alpha1"
	"knative.dev/pkg/apis/istio/common/v1alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

type jwt struct {
	vsClient *istioClient.VirtualService
	apClient *istioClient.AuthenticationPolicy
}

func (j *jwt) Process(ctx context.Context, api *gatewayv2alpha1.Gate) error {
	fmt.Println("Processing API for JWT")

	oldVS, err := j.getVirtualService(ctx, api)
	if err != nil {
		return err
	}
	if oldVS != nil {
		return j.updateVirtualService(ctx, j.prepareVirtualService(api, oldVS))
	}
	err = j.createVirtualService(ctx, j.generateVirtualService(api))
	if err != nil {
		return err
	}

	oldAP, err := j.getAuthenticationPolicy(ctx, api)
	if err != nil {
		return err
	}
	if oldAP != nil {
		return j.updateAuthenticationPolicy(ctx, j.prepareAuthenticationPolicy(api, oldAP))
	}
	err = j.createAuthenticationPolicy(ctx, j.generateAuthenticationPolicy(api))
	if err != nil {
		return err
	}

	return nil
}

func (j *jwt) getVirtualService(ctx context.Context, api *gatewayv2alpha1.Gate) (*networkingv1alpha3.VirtualService, error) {
	vs, err := j.vsClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return vs, nil
}

func (j *jwt) createVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return j.vsClient.Create(ctx, vs)
}

func (j *jwt) prepareVirtualService(api *gatewayv2alpha1.Gate, vs *networkingv1alpha3.VirtualService) *networkingv1alpha3.VirtualService {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	controller := true

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

func (j *jwt) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return j.vsClient.Update(ctx, vs)
}

func (j *jwt) generateVirtualService(api *gatewayv2alpha1.Gate) *networkingv1alpha3.VirtualService {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	controller := true

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

// ---
// apiVersion: networking.istio.io/v1alpha3
// kind: VirtualService
// metadata:
//   name: httpbin-httpbin-proxy
//   namespace: default
// spec:
//   gateways:
//   - kyma-gateway.kyma-system.svc.cluster.local
//   hosts:
//   - httpbin-proxy.kyma.local
//   http:
//     - match:
//       - uri:
//         regex: /.*
//     route:
//     - destination:
//         host: httpbin.default.svc.cluster.local
//         port:
//           number: 8000
// ---
// apiVersion: authentication.istio.io/v1alpha1
// kind: Policy
// metadata:
//   name: httpbin-httpbin-proxy
//   namespace: default
// spec:
//   origins:
//   - jwt:
//       issuer: https://dex.kyma.local
//       jwksUri: http://dex-service.kyma-system.svc.cluster.local:5556/keys
//   peers:
//   - mtls: {}
//   principalBinding: USE_ORIGIN
//   targets:
//   - name: httpbin

func (j *jwt) getAuthenticationPolicy(ctx context.Context, api *gatewayv2alpha1.Gate) (*authenticationv1alpha1.Policy, error) {
	ap, err := j.apClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return ap, nil
}

func (j *jwt) createAuthenticationPolicy(ctx context.Context, ap *authenticationv1alpha1.Policy) error {
	return j.apClient.Create(ctx, ap)
}

func (j *jwt) updateAuthenticationPolicy(ctx context.Context, ap *authenticationv1alpha1.Policy) error {
	return j.apClient.Update(ctx, ap)
}

func (j *jwt) prepareAuthenticationPolicy(api *gatewayv2alpha1.Gate, ap *authenticationv1alpha1.Policy) *authenticationv1alpha1.Policy {
	authenticationPolicyName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	controller := true

	ownerRef := &k8sMeta.OwnerReference{
		Name:       api.ObjectMeta.Name,
		APIVersion: api.TypeMeta.APIVersion,
		Kind:       api.TypeMeta.Kind,
		UID:        api.ObjectMeta.UID,
		Controller: &controller,
	}

	ap.ObjectMeta.OwnerReferences = []k8sMeta.OwnerReference{*ownerRef}
	ap.ObjectMeta.Name = authenticationPolicyName
	ap.ObjectMeta.Namespace = api.ObjectMeta.Namespace

	targets := []authenticationv1alpha1.TargetSelector{
		{
			Name: fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace),
		},
	}
	peers := []authenticationv1alpha1.PeerAuthenticationMethod{
		{
			Mtls: &authenticationv1alpha1.MutualTLS{},
		},
	}

	spec := &authenticationv1alpha1.PolicySpec{
		Targets:          targets,
		PrincipalBinding: authenticationv1alpha1.PrincipalBindingUserOrigin,
		Peers:            peers,
	}

	ap.Spec = *spec

	return ap

}

func (j *jwt) generateAuthenticationPolicy(api *gatewayv2alpha1.Gate) *authenticationv1alpha1.Policy {
	authenticationPolicyName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
	controller := true

	ownerRef := &k8sMeta.OwnerReference{
		Name:       api.ObjectMeta.Name,
		APIVersion: api.TypeMeta.APIVersion,
		Kind:       api.TypeMeta.Kind,
		UID:        api.ObjectMeta.UID,
		Controller: &controller,
	}

	objectMeta := k8sMeta.ObjectMeta{
		Name:            authenticationPolicyName,
		Namespace:       api.ObjectMeta.Namespace,
		OwnerReferences: []k8sMeta.OwnerReference{*ownerRef},
	}
	targets := []authenticationv1alpha1.TargetSelector{
		{
			Name: fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace),
		},
	}
	peers := []authenticationv1alpha1.PeerAuthenticationMethod{
		{
			Mtls: &authenticationv1alpha1.MutualTLS{},
		},
	}

	spec := &authenticationv1alpha1.PolicySpec{
		Targets:          targets,
		PrincipalBinding: authenticationv1alpha1.PrincipalBindingUserOrigin,
		Peers:            peers,
	}

	ap := &authenticationv1alpha1.Policy{
		ObjectMeta: objectMeta,
		Spec:       *spec,
	}
	return ap
}
