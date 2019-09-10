package builders

import (
	networkingv1alpha1 "knative.dev/pkg/apis/istio/common/v1alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

// VirtualService returns builder for knative.dev/pkg/apis/istio/v1alpha3/VirtualService type
func VirtualService() *virtualService {
	return &virtualService{
		value: &networkingv1alpha3.VirtualService{},
	}
}

type virtualService struct {
	value *networkingv1alpha3.VirtualService
}

func (vs *virtualService) From(val *networkingv1alpha3.VirtualService) *virtualService {
	vs.value = val
	return vs
}

func (vs *virtualService) Name(val string) *virtualService {
	vs.value.Name = val
	return vs
}

func (vs *virtualService) Namespace(val string) *virtualService {
	vs.value.Namespace = val
	return vs
}

func (vs *virtualService) Owner(val *ownerReference) *virtualService {
	vs.value.OwnerReferences = append(vs.value.OwnerReferences, *val.Get())
	return vs
}

func (vs *virtualService) Spec(val *virtualServiceSpec) *virtualService {
	vs.value.Spec = *val.Get()
	return vs
}

func (vs *virtualService) Get() *networkingv1alpha3.VirtualService {
	return vs.value
}

// VirtualServiceSpec returns builder for knative.dev/pkg/apis/istio/v1alpha3/VirtualServiceSpec type
func VirtualServiceSpec() *virtualServiceSpec {
	return &virtualServiceSpec{
		value: &networkingv1alpha3.VirtualServiceSpec{},
	}
}

type virtualServiceSpec struct {
	value *networkingv1alpha3.VirtualServiceSpec
}

func (b *virtualServiceSpec) From(val *networkingv1alpha3.VirtualServiceSpec) *virtualServiceSpec {
	b.value = val
	return b
}

func (b *virtualServiceSpec) Host(val string) *virtualServiceSpec {
	b.value.Hosts = append(b.value.Hosts, val)
	return b
}

func (b *virtualServiceSpec) Gateway(val string) *virtualServiceSpec {
	b.value.Gateways = append(b.value.Gateways, val)
	return b
}

func (b *virtualServiceSpec) HTTP(mr *matchRequest, rd *routeDestination) *virtualServiceSpec {
	var httpMatch []networkingv1alpha3.HTTPMatchRequest
	var routeDest []networkingv1alpha3.HTTPRouteDestination

	if mr != nil {
		httpMatch = append(httpMatch, *mr.Get())
	}

	if rd != nil {
		routeDest = append(routeDest, *rd.Get())
	}

	b.value.HTTP = []networkingv1alpha3.HTTPRoute{
		{
			Match: httpMatch,
			Route: routeDest,
		},
	}
	//b.routeDest = rd

	return b
}

func (b *virtualServiceSpec) Get() *networkingv1alpha3.VirtualServiceSpec {
	return b.value
}

// MatchRequest returns builder for knative.dev/pkg/apis/istio/v1alpha3/HTTPMatchRequest type
func MatchRequest() *matchRequest {
	return &matchRequest{}
}

type matchRequest struct {
	data *networkingv1alpha3.HTTPMatchRequest
}

func (mr *matchRequest) Get() *networkingv1alpha3.HTTPMatchRequest {
	return mr.data
}

func (mr *matchRequest) URI() *stringMatch {
	mr.data = &networkingv1alpha3.HTTPMatchRequest{}
	mr.data.URI = &networkingv1alpha1.StringMatch{}
	return &stringMatch{mr.data.URI, func() *matchRequest { return mr }}
}

type stringMatch struct {
	value  *networkingv1alpha1.StringMatch
	parent func() *matchRequest
}

func (st *stringMatch) Regex(val string) *matchRequest {
	st.value.Regex = val
	return st.parent()
}

// RouteDestination returns builder for knative.dev/pkg/apis/istio/v1alpha3/HTTPRouteDestination type
func RouteDestination() *routeDestination {
	return &routeDestination{&networkingv1alpha3.HTTPRouteDestination{}}
}

type routeDestination struct {
	value *networkingv1alpha3.HTTPRouteDestination
}

func (rd *routeDestination) Host(val string) *routeDestination {
	rd.value.Destination.Host = val
	return rd
}
func (rd *routeDestination) Port(val uint32) *routeDestination {
	rd.value.Destination.Port.Number = val
	return rd
}
func (rd *routeDestination) Get() *networkingv1alpha3.HTTPRouteDestination {
	return rd.value
}
