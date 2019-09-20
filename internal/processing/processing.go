package processing

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/kyma-incubator/api-gateway/internal/builders"

	"github.com/go-logr/logr"
	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	oryClient "github.com/kyma-incubator/api-gateway/internal/clients/ory"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
)

//Factory .
type Factory struct {
	vsClient          *istioClient.VirtualService
	arClient          *oryClient.AccessRule
	Log               logr.Logger
	oathkeeperSvc     string
	oathkeeperSvcPort uint32
	JWKSURI           string
}

//NewFactory .
func NewFactory(vsClient *istioClient.VirtualService, arClient *oryClient.AccessRule, logger logr.Logger, oathkeeperSvc string, oathkeeperSvcPort uint32, jwksURI string) *Factory {
	return &Factory{
		vsClient:          vsClient,
		arClient:          arClient,
		Log:               logger,
		oathkeeperSvc:     oathkeeperSvc,
		oathkeeperSvcPort: oathkeeperSvcPort,
		JWKSURI:           jwksURI,
	}
}

//// RequiredObjects carries required state of the cluster after reconciliation
//type RequiredObjects struct {
//	virtualServices []*networkingv1alpha3.VirtualService
//	accessRules     []*rulev1alpha1.Rule
//}

// CalculateRequiredState returns required state of all objects related to given api
func (f *Factory) CalculateRequiredState(api *gatewayv1alpha1.APIRule) State {
	var res State

	res.accessRules = make(map[string]*rulev1alpha1.Rule)

	for i, rule := range api.Spec.Rules {
		if isSecured(rule) {
			ar := generateAccessRule(api, api.Spec.Rules[i], i, rule.AccessStrategies)
			res.accessRules[rule.Path] = ar
		}
	}

	//Only one vs
	vs := f.generateVirtualService(api)
	res.virtualService = vs

	return res
}

type State struct {
	virtualService *networkingv1alpha3.VirtualService
	accessRules    map[string]*rulev1alpha1.Rule
}

func (f *Factory) GetActualState(ctx context.Context, api *gatewayv1alpha1.APIRule) (error, State) {
	labels := make(map[string]string)
	labels["owner"] = fmt.Sprintf("%s.%s", api.ObjectMeta.Name, api.ObjectMeta.Namespace)
	var obj State

	vsList, err := f.vsClient.GetForLabels(ctx, labels)
	if err != nil {
		return err, obj
	}

	//what to do if len(vsList) > 1?

	obj.virtualService = vsList[0]

	arList, err := f.arClient.GetForLabels(ctx, labels)
	if err != nil {
		return err, obj
	}

	obj.accessRules = make(map[string]*rulev1alpha1.Rule)

	for _, ar := range arList {
		obj.accessRules[ar.Spec.Match.URL] = ar
	}

	return nil, obj
}

type Patch struct {
	virtualService *idk
	accessRule     map[string]*idk
}

type idk struct {
	action string
	obj    runtime.Object
}

func (f *Factory) CalculateDiff(requiredState State, actualState State) *Patch {
	arPatch := make(map[string]*idk)

	for path, rule := range requiredState.accessRules {
		if actualState.accessRules[path] != nil {
			arPatch[path].action = "update"
			modifyAccessRule(actualState.accessRules[path], rule)
		} else {
			arPatch[path].action = "create"
		}
		arPatch[path].obj = rule
	}

	for path, rule := range actualState.accessRules {
		if arPatch[path] == nil {
			arPatch[path].action = "delete"
			arPatch[path].obj = rule
		}
	}

	vsPatch := &idk{}

	if actualState.virtualService != nil {
		vsPatch.action = "update"
		f.modifyVirtualService(actualState.virtualService, requiredState.virtualService)
	} else {
		vsPatch.action = "create"
	}

	vsPatch.obj = requiredState.virtualService

	objPatch := &Patch{}
	objPatch.accessRule = arPatch
	objPatch.virtualService = vsPatch

	return objPatch
}

func (f *Factory) ApplyDiff(ctx context.Context, patch *Patch) error{
	if patch.virtualService.action == "create" {
		f.vsClient.Create(ctx, networkingv1alpha3.VirtualService(patch.virtualService.obj))
	}

	if patch.virtualService.action == "update" {
		f.vsClient.Update(ctx, patch.virtualService.obj)
	}

	for _, rule := range patch.accessRule{
		if rule.action == "create"{
			//create
		}

		if rule.action == "update"{
			//update
		}

		if rule.action == "delete"{
			//delete
		}
	}

}

// ApplyRequiredState applies required state to the cluster
//TODO: It should be possible to get rid of api parameter
func (f *Factory) ApplyRequiredState(ctx context.Context, requiredState State, api *gatewayv1alpha1.APIRule) error {
	i := 0

	for _, rule := range requiredState.accessRules {
		err := f.processAR(ctx, rule, api, i)
		i++
		if err != nil {
			return err
		}
	}

	//for _, vs := range requiredState.virtualServices {
	err := f.processVS(ctx, requiredState.virtualService, api)
	if err != nil {
		return err
	}
	//}
	return nil
}

func (f *Factory) getVirtualService(ctx context.Context, api *gatewayv1alpha1.APIRule) (*networkingv1alpha3.VirtualService, error) {
	vs, err := f.vsClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return vs, nil
}

func (f *Factory) createVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return f.vsClient.Create(ctx, vs)
}

func (f *Factory) updateVirtualService(ctx context.Context, vs *networkingv1alpha3.VirtualService) error {
	return f.vsClient.Update(ctx, vs)
}

func (f *Factory) createAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return f.arClient.Create(ctx, ar)
}

func (f *Factory) updateAccessRule(ctx context.Context, ar *rulev1alpha1.Rule) error {
	return f.arClient.Update(ctx, ar)
}

func (f *Factory) getAccessRule(ctx context.Context, api *gatewayv1alpha1.APIRule, ruleInd int) (*rulev1alpha1.Rule, error) {
	ar, err := f.arClient.GetForAPI(ctx, api, ruleInd)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return ar, nil
}

func (f *Factory) processVS(ctx context.Context, required *networkingv1alpha3.VirtualService, api *gatewayv1alpha1.APIRule) error {
	existingVS, err := f.getVirtualService(ctx, api)
	if err != nil {
		return err
	}
	if existingVS != nil {
		f.modifyVirtualService(existingVS, required)
		return f.updateVirtualService(ctx, existingVS)
	}
	return f.createVirtualService(ctx, required)
}

func (f *Factory) processAR(ctx context.Context, required *rulev1alpha1.Rule, api *gatewayv1alpha1.APIRule, ruleInd int) error {
	existingAR, err := f.getAccessRule(ctx, api, ruleInd)
	if err != nil {
		return err
	}

	if existingAR != nil {
		modifyAccessRule(existingAR, required)
		return f.updateAccessRule(ctx, existingAR)
	}
	return f.createAccessRule(ctx, required)
}

//TODO: Find better name (updateVirtualService is already taken)
func (f *Factory) modifyVirtualService(existing, required *networkingv1alpha3.VirtualService) {
	existing.Spec = required.Spec
}

func (f *Factory) generateVirtualService(api *gatewayv1alpha1.APIRule) *networkingv1alpha3.VirtualService {
	virtualServiceName := fmt.Sprintf("%s-%s", api.ObjectMeta.Name, *api.Spec.Service.Name)
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
