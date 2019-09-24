package processing

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/api-gateway/internal/builders"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	oryClient "github.com/kyma-incubator/api-gateway/internal/clients/ory"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

//Factory .
type Factory struct {
	vsClient          *istioClient.VirtualService
	arClient          *oryClient.AccessRule
	client            client.Client
	Log               logr.Logger
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	JWKSURI           string
}

//NewFactory .
func NewFactory(vsClient *istioClient.VirtualService, arClient *oryClient.AccessRule, client client.Client, logger logr.Logger, oathkeeperSvc string, oathkeeperSvcPort uint32, jwksURI string) *Factory {
	return &Factory{
		vsClient:          vsClient,
		arClient:          arClient,
		client:            client,
		Log:               logger,
		oathkeeperSvc:     oathkeeperSvc,
		oathkeeperSvcPort: oathkeeperSvcPort,
		JWKSURI:           jwksURI,
	}
}

// CalculateRequiredState returns required state of all objects related to given api
func (f *Factory) CalculateRequiredState(api *gatewayv1alpha1.APIRule) *State {
	var res State

	res.accessRules = make(map[string]*rulev1alpha1.Rule)

	for _, rule := range api.Spec.Rules {
		if isSecured(rule) {
			ar := generateAccessRule(api, rule, rule.AccessStrategies)
			res.accessRules[ar.Spec.Match.URL] = ar
		}
	}

	//Only one vs
	vs := f.generateVirtualService(api)
	res.virtualService = vs

	return &res
}

//State represents desired or actual state of Istio Virtual Services and Oathkeeper Rules
type State struct {
	virtualService *networkingv1alpha3.VirtualService
	accessRules    map[string]*rulev1alpha1.Rule
}

//GetActualState methods gets actual state of Istio Virtual Services and Oathkeeper Rules
func (f *Factory) GetActualState(ctx context.Context, api *gatewayv1alpha1.APIRule) (*State, error) {
	labels := make(map[string]string)
	labels["owner"] = fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)
	var state State

	vsList, err := f.vsClient.GetForLabels(ctx, labels)
	if err != nil {
		return nil, err
	}

	//what to do if len(vsList) > 1?

	if len(vsList) == 1 {
		state.virtualService = &vsList[0]
	} else {
		state.virtualService = nil
	}

	arList, err := f.arClient.GetForLabels(ctx, labels)
	if err != nil {
		return nil, err
	}

	state.accessRules = make(map[string]*rulev1alpha1.Rule)
	for _, ar := range arList {
		state.accessRules[ar.Spec.Match.URL] = &ar
	}

	return &state, nil
}

//Patch represents diff between desired and actual state
type Patch struct {
	virtualService *objToPatch
	accessRule     map[string]*objToPatch
}

type objToPatch struct {
	action string
	obj    runtime.Object
}

//CalculateDiff methods compute diff between desired & actual state
func (f *Factory) CalculateDiff(requiredState *State, actualState *State) *Patch {
	arPatch := make(map[string]*objToPatch)

	for path, rule := range requiredState.accessRules {
		rulePatch := &objToPatch{}

		if actualState.accessRules[path] != nil {
			rulePatch.action = "update"
			modifyAccessRule(actualState.accessRules[path], rule)
			rulePatch.obj = actualState.accessRules[path]
		} else {
			rulePatch.action = "create"
			rulePatch.obj = rule
		}

		arPatch[path] = rulePatch
	}

	for path, rule := range actualState.accessRules {
		if requiredState.accessRules[path] == nil {
			objToDelete := &objToPatch{action: "delete", obj: rule}
			arPatch[path] = objToDelete
		}
	}

	vsPatch := &objToPatch{}
	if actualState.virtualService != nil {
		vsPatch.action = "update"
		f.updateVirtualService(actualState.virtualService, requiredState.virtualService)
		vsPatch.obj = actualState.virtualService
	} else {
		vsPatch.action = "create"
		vsPatch.obj = requiredState.virtualService
	}

	return &Patch{virtualService: vsPatch, accessRule: arPatch}
}

//ApplyDiff method applies computed diff
func (f *Factory) ApplyDiff(ctx context.Context, patch *Patch) error {
	err := f.applyObjDiff(ctx, patch.virtualService)
	if err != nil {
		return err
	}

	for _, rule := range patch.accessRule {
		err := f.applyObjDiff(ctx, rule)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *Factory) applyObjDiff(ctx context.Context, objToPatch *objToPatch) error {
	var err error

	switch objToPatch.action {
	case "create":
		err = f.client.Create(ctx, objToPatch.obj)
	case "update":
		err = f.client.Update(ctx, objToPatch.obj)
	case "delete":
		err = f.client.Delete(ctx, objToPatch.obj)
	}

	if err != nil {
		return err
	}

	return nil
}

func (f *Factory) updateVirtualService(existing, required *networkingv1alpha3.VirtualService) {
	existing.Spec = required.Spec
}

func (f *Factory) generateVirtualService(api *gatewayv1alpha1.APIRule) *networkingv1alpha3.VirtualService {
	virtualServiceName := fmt.Sprintf("%s", api.ObjectMeta.Name)
	ownerRef := generateOwnerRef(api)

	vsSpecBuilder := builders.VirtualServiceSpec()
	vsSpecBuilder.Host(*api.Spec.Service.Host)
	vsSpecBuilder.Gateway(*api.Spec.Gateway)

	for _, rule := range api.Spec.Rules {
		httpRouteBuilder := builders.HTTPRoute()

		if isSecured(rule) {
			httpRouteBuilder.Route(builders.RouteDestination().Host(f.oathkeeperSvc).Port(f.oathkeeperSvcPort))
		} else {
			destinationHost := fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace)
			httpRouteBuilder.Route(builders.RouteDestination().Host(destinationHost).Port(*api.Spec.Service.Port))
		}

		httpRouteBuilder.Match(builders.MatchRequest().URI().Regex(rule.Path))
		vsSpecBuilder.HTTP(httpRouteBuilder)
	}

	vsBuilder := builders.VirtualService().
		Name(virtualServiceName).
		Namespace(api.ObjectMeta.Namespace).
		Owner(builders.OwnerReference().From(&ownerRef)).
		Label("owner", fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace))

	vsBuilder.Spec(vsSpecBuilder)

	return vsBuilder.Get()
}
