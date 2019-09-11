package builders

import (
	networkingv1alpha1 "knative.dev/pkg/apis/istio/common/v1alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

// VirtualService returns builder for knative.dev/pkg/apis/istio/v1alpha3/VirtualService type
func VirtualService() *virtualService {
	return &virtualService{}
}

type virtualServiceFunc func(value *networkingv1alpha3.VirtualService)

type virtualService struct {
	fns    []virtualServiceFunc
	intVal *networkingv1alpha3.VirtualService
}

func (vs *virtualService) add(fn virtualServiceFunc) {
	vs.fns = append(vs.fns, fn)
}

func (vs *virtualService) Name(val string) *virtualService {
	vs.add(func(value *networkingv1alpha3.VirtualService) {
		value.Name = val
	})
	return vs
}

func (vs *virtualService) Namespace(val string) *virtualService {
	vs.add(func(value *networkingv1alpha3.VirtualService) {
		value.Namespace = val
	})
	return vs
}

func (vs *virtualService) Owner(val *ownerReference) *virtualService {
	vs.add(func(value *networkingv1alpha3.VirtualService) {
		value.OwnerReferences = append(value.OwnerReferences, *val.Get())
	})
	return vs
}

func (vs *virtualService) Spec(val *virtualServiceSpec) *virtualService {
	vs.add(func(value *networkingv1alpha3.VirtualService) {
		value.Spec = *val.Get()
	})
	return vs
}

func (vs *virtualService) get(value *networkingv1alpha3.VirtualService) *networkingv1alpha3.VirtualService {
	for i := 0; i < len(vs.fns); i++ {
		vs.fns[i](value)
	}
	return value
}

func (vs *virtualService) Get() *networkingv1alpha3.VirtualService {
	if vs.intVal == nil {
		return vs.get(&networkingv1alpha3.VirtualService{})
	}
	return vs.get(vs.intVal)
}
func (vs *virtualService) From(val *networkingv1alpha3.VirtualService) *virtualService {
	vs.intVal = val
	return vs
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

func (vss *virtualServiceSpec) From(val *networkingv1alpha3.VirtualServiceSpec) *virtualServiceSpec {
	vss.value = val
	return vss
}

func (vss *virtualServiceSpec) Host(val string) *virtualServiceSpec {
	vss.value.Hosts = append(vss.value.Hosts, val)
	return vss
}

func (vss *virtualServiceSpec) Gateway(val string) *virtualServiceSpec {
	vss.value.Gateways = append(vss.value.Gateways, val)
	return vss
}

func (vss *virtualServiceSpec) HTTP(mr *matchRequest, rd *routeDestination) *virtualServiceSpec {
	var httpMatch []networkingv1alpha3.HTTPMatchRequest
	var routeDest []networkingv1alpha3.HTTPRouteDestination

	if mr != nil {
		httpMatch = append(httpMatch, *mr.Get())
	}

	if rd != nil {
		routeDest = append(routeDest, *rd.Get())
	}

	vss.value.HTTP = []networkingv1alpha3.HTTPRoute{
		{
			Match: httpMatch,
			Route: routeDest,
		},
	}

	return vss
}

func (vss *virtualServiceSpec) Get() *networkingv1alpha3.VirtualServiceSpec {
	return vss.value
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
