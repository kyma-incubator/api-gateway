package builders

import (
	networkingv1alpha1 "knative.dev/pkg/apis/istio/common/v1alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

// VirtualService returns builder for knative.dev/pkg/apis/istio/v1alpha3/VirtualService type
// This builder is deferred - it records all requested changes and then "replays" them on existing object with `Set()` method or on a new one with `New()` method
func VirtualService() *virtualService {
	return &virtualService{}
}

type virtualService struct {
	fns []virtualServiceFunc
}

type virtualServiceFunc func(target *networkingv1alpha3.VirtualService)

func (vs *virtualService) add(fn virtualServiceFunc) {
	vs.fns = append(vs.fns, fn)
}

func (vs *virtualService) set(target *networkingv1alpha3.VirtualService) *networkingv1alpha3.VirtualService {
	for i := 0; i < len(vs.fns); i++ {
		vs.fns[i](target)
	}
	return target
}

func (vs *virtualService) Name(value string) *virtualService {
	vs.add(func(target *networkingv1alpha3.VirtualService) {
		target.Name = value
	})
	return vs
}

func (vs *virtualService) Namespace(value string) *virtualService {
	vs.add(func(target *networkingv1alpha3.VirtualService) {
		target.Namespace = value
	})
	return vs
}

func (vs *virtualService) Owner(value *ownerReference) *virtualService {
	vs.add(func(target *networkingv1alpha3.VirtualService) {
		target.OwnerReferences = append(target.OwnerReferences, *value.Get())
	})
	return vs
}

func (vs *virtualService) Spec(specBuilder *virtualServiceSpec) *virtualService {
	vs.add(func(target *networkingv1alpha3.VirtualService) {
		target.Spec = *specBuilder.New() //replaces entire Spec with data returned by given builder, merge is not supported.
	})
	return vs
}

//New replays all requested changes on a new networkingv1alpha3.VirtualService instance
func (vs *virtualService) New() *networkingv1alpha3.VirtualService {
	return vs.set(&networkingv1alpha3.VirtualService{})
}

//Set replays all requested changes on an existing networkingv1alpha3.VirtualService instance
func (vs *virtualService) Set(target *networkingv1alpha3.VirtualService) *networkingv1alpha3.VirtualService {
	return vs.set(target)
}

// VirtualServiceSpec returns builder for knative.dev/pkg/apis/istio/v1alpha3/VirtualServiceSpec type
// This builder is deferred - it records all requested changes and then "replays" them on existing object with `Set()` method or on a new one with `New()` method
func VirtualServiceSpec() *virtualServiceSpec {
	return &virtualServiceSpec{}
}

type virtualServiceSpec struct {
	fns []virtualServiceSpecFunc
}

type virtualServiceSpecFunc func(target *networkingv1alpha3.VirtualServiceSpec)

func (vss *virtualServiceSpec) add(fn virtualServiceSpecFunc) {
	vss.fns = append(vss.fns, fn)
}

func (vss *virtualServiceSpec) set(target *networkingv1alpha3.VirtualServiceSpec) *networkingv1alpha3.VirtualServiceSpec {
	for i := 0; i < len(vss.fns); i++ {
		vss.fns[i](target)
	}
	return target
}

func (vss *virtualServiceSpec) Host(value string) *virtualServiceSpec {
	vss.add(func(target *networkingv1alpha3.VirtualServiceSpec) {
		target.Hosts = append(target.Hosts, value)
	})
	return vss
}

func (vss *virtualServiceSpec) Gateway(value string) *virtualServiceSpec {
	vss.add(func(target *networkingv1alpha3.VirtualServiceSpec) {
		target.Gateways = append(target.Gateways, value)
	})
	return vss
}

func (vss *virtualServiceSpec) HTTP(hr *httpRoute) *virtualServiceSpec {
	vss.add(func(target *networkingv1alpha3.VirtualServiceSpec) {
		target.HTTP = append(target.HTTP, *hr.New())
	})
	return vss
}

//New replays all requested changes on a new networkingv1alpha3.VirtualServiceSpec instance
func (vss *virtualServiceSpec) New() *networkingv1alpha3.VirtualServiceSpec {
	return vss.set(&networkingv1alpha3.VirtualServiceSpec{})
}

//Set replays all requested changes on an existing networkingv1alpha3.VirtualServiceSpec instance
func (vss *virtualServiceSpec) Set(val *networkingv1alpha3.VirtualServiceSpec) *networkingv1alpha3.VirtualServiceSpec {
	return vss.set(val)
}

// HTTPRoute returns builder for knative.dev/pkg/apis/istio/v1alpha3/HTTPRoute type
// This builder is deferred - it records all requested changes and then "replays" them on existing object with `Set()` method or on a new one with `New()` method
func HTTPRoute() *httpRoute {
	return &httpRoute{}
}

type httpRoute struct {
	fns []httpRouteFunc
}

type httpRouteFunc func(target *networkingv1alpha3.HTTPRoute)

func (hr *httpRoute) add(fn httpRouteFunc) {
	hr.fns = append(hr.fns, fn)
}

func (hr *httpRoute) set(target *networkingv1alpha3.HTTPRoute) *networkingv1alpha3.HTTPRoute {
	for i := 0; i < len(hr.fns); i++ {
		hr.fns[i](target)
	}
	return target
}

func (hr *httpRoute) Match(mr *matchRequest) *httpRoute {
	hr.add(func(target *networkingv1alpha3.HTTPRoute) {
		target.Match = append(target.Match, *mr.New())
	})
	return hr
}

func (hr *httpRoute) Route(rd *routeDestination) *httpRoute {
	hr.add(func(target *networkingv1alpha3.HTTPRoute) {
		target.Route = append(target.Route, *rd.New())
	})
	return hr
}

//New replays all requested changes on a new networkingv1alpha3.HTTPRoute instance
func (hr *httpRoute) New() *networkingv1alpha3.HTTPRoute {
	return hr.set(&networkingv1alpha3.HTTPRoute{})
}

//Set replays all requested changes on an existing networkingv1alpha3.HTTPRoute instance
func (hr *httpRoute) Set(target *networkingv1alpha3.HTTPRoute) *networkingv1alpha3.HTTPRoute {
	return hr.set(target)
}

// MatchRequest returns builder for knative.dev/pkg/apis/istio/v1alpha3/HTTPMatchRequest type
// This builder is deferred - it records all requested changes and then "replays" them on existing object with `Set()` method or on a new one with `New()` method
func MatchRequest() *matchRequest {
	return &matchRequest{}
}

type matchRequest struct {
	fns []matchRequestFunc
}

type matchRequestFunc func(target *networkingv1alpha3.HTTPMatchRequest)

func (mr *matchRequest) add(fn matchRequestFunc) {
	mr.fns = append(mr.fns, fn)
}

func (mr *matchRequest) set(target *networkingv1alpha3.HTTPMatchRequest) *networkingv1alpha3.HTTPMatchRequest {
	for i := 0; i < len(mr.fns); i++ {
		mr.fns[i](target)
	}
	return target
}

func (mr *matchRequest) URI() *stringMatch {
	res := &networkingv1alpha1.StringMatch{}

	mr.add(func(target *networkingv1alpha3.HTTPMatchRequest) {
		target.URI = res
	})

	return &stringMatch{
		parent: mr,
		value:  res,
	}
}

//New replays all requested changes on a new networkingv1alpha3.HTTPMatchRequest instance
func (mr *matchRequest) New() *networkingv1alpha3.HTTPMatchRequest {
	return mr.set(&networkingv1alpha3.HTTPMatchRequest{})
}

//Set replays all requested changes on an existing networkingv1alpha3.HTTPMatchRequest instance
func (mr *matchRequest) Set(target *networkingv1alpha3.HTTPMatchRequest) *networkingv1alpha3.HTTPMatchRequest {
	return mr.set(target)
}

type stringMatch struct {
	parent *matchRequest
	value  *networkingv1alpha1.StringMatch
}

func (st *stringMatch) Regex(val string) *matchRequest {
	st.value.Regex = val
	return st.parent
}

// RouteDestination returns builder for knative.dev/pkg/apis/istio/v1alpha3/HTTPRouteDestination type
// This builder is deferred - it records all requested changes and then "replays" them on existing object with `Set()` method or on a new one with `New()` method
func RouteDestination() *routeDestination {
	return &routeDestination{}
}

type routeDestination struct {
	fns []routeDestinationFunc
}

type routeDestinationFunc func(target *networkingv1alpha3.HTTPRouteDestination)

func (rd *routeDestination) add(fn routeDestinationFunc) {
	rd.fns = append(rd.fns, fn)
}

func (rd *routeDestination) set(target *networkingv1alpha3.HTTPRouteDestination) *networkingv1alpha3.HTTPRouteDestination {
	for i := 0; i < len(rd.fns); i++ {
		rd.fns[i](target)
	}
	return target
}

func (rd *routeDestination) Host(value string) *routeDestination {
	rd.add(func(target *networkingv1alpha3.HTTPRouteDestination) {
		target.Destination.Host = value
	})
	return rd
}

func (rd *routeDestination) Port(value uint32) *routeDestination {
	rd.add(func(target *networkingv1alpha3.HTTPRouteDestination) {
		target.Destination.Port.Number = value
	})
	return rd
}

//New replays all requested changes on a new networkingv1alpha3.HTTPRouteDestination instance
func (rd *routeDestination) New() *networkingv1alpha3.HTTPRouteDestination {
	return rd.set(&networkingv1alpha3.HTTPRouteDestination{})
}

//Set replays all requested changes on an existing networkingv1alpha3.HTTPRouteDestination instance
func (rd *routeDestination) Set(target *networkingv1alpha3.HTTPRouteDestination) *networkingv1alpha3.HTTPRouteDestination {
	return rd.set(target)
}
