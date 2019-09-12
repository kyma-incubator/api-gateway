package processing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	oryClient "github.com/kyma-incubator/api-gateway/internal/clients/ory"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
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

//Strategy .
type Strategy interface {
	Process(ctx context.Context, api *gatewayv2alpha1.Gate) error
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

//StrategyFor .
func (f *Factory) StrategyFor(strategyName string) (Strategy, error) {
	switch strategyName {
	case gatewayv2alpha1.Allow:
		f.Log.Info("Allow processing mode detected")
		return &allow{vsClient: f.vsClient, oathkeeperSvc: f.oathkeeperSvc, oathkeeperSvcPort: f.oathkeeperSvcPort}, nil
	case gatewayv2alpha1.Jwt:
		f.Log.Info("JWT processing mode detected")
		return &jwt{vsClient: f.vsClient, arClient: f.arClient, JWKSURI: f.JWKSURI, oathkeeperSvc: f.oathkeeperSvc, oathkeeperSvcPort: f.oathkeeperSvcPort}, nil
	case gatewayv2alpha1.Oauth:
		f.Log.Info("OAUTH processing mode detected")
		return &oauth{vsClient: f.vsClient, arClient: f.arClient, oathkeeperSvc: f.oathkeeperSvc, oathkeeperSvcPort: f.oathkeeperSvcPort}, nil
	default:
		return nil, fmt.Errorf("unsupported mode: %s", strategyName)
	}
}

// Run ?
func (f *Factory) Run(ctx context.Context, api *gatewayv2alpha1.Gate) error {
	var destinationHost string
	var destinationPort uint32
	var err error
	var accessStrategies []*rulev1alpha1.Authenticator
	// Get gate
	// Get list of Paths
	// Check Paths - validate
	//Process path one by one
	for i := range api.Spec.Rules {
		// Single Path level
		// Validate
		// Process

		if isSecured(api.Spec.Rules[i]) {
			destinationHost = f.oathkeeperSvc
			destinationPort = f.oathkeeperSvcPort
		} else {
			destinationHost = fmt.Sprintf("%s.%s.svc.cluster.local", *api.Spec.Service.Name, api.ObjectMeta.Namespace)
			destinationPort = *api.Spec.Service.Port
		}

		for j := range api.Spec.Rules[i].AccessStrategy {
			// Single Strategy for single Path
			// Compile Oathkeeper config from this
			accessStrategies = append(accessStrategies, api.Spec.Rules[i].AccessStrategy[j])
		}
		err = f.processAR(ctx, api, accessStrategies)
		if err != nil {
			return err
		}
		err = f.processVS(ctx, api, destinationHost, destinationPort)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Factory) getVirtualService(ctx context.Context, api *gatewayv2alpha1.Gate) (*networkingv1alpha3.VirtualService, error) {
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

func (f *Factory) getAccessRule(ctx context.Context, api *gatewayv2alpha1.Gate) (*rulev1alpha1.Rule, error) {
	ar, err := f.arClient.GetForAPI(ctx, api)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return ar, nil
}

func (f *Factory) processVS(ctx context.Context, api *gatewayv2alpha1.Gate, destinationHost string, destinationPort uint32) error {
	oldVS, err := f.getVirtualService(ctx, api)
	if err != nil {
		return err
	}

	if oldVS != nil {
		newVS := prepareVirtualService(api, oldVS, destinationHost, destinationPort, api.Spec.Rules[0].Path)
		return f.updateVirtualService(ctx, newVS)
	}
	vs := generateVirtualService(api, destinationHost, destinationPort, api.Spec.Rules[0].Path)
	fmt.Printf("---\n%+v\n", vs)
	return f.createVirtualService(ctx, vs)
}

func (f *Factory) processAR(ctx context.Context, api *gatewayv2alpha1.Gate, config []*rulev1alpha1.Authenticator) error {
	ar := &rulev1alpha1.Rule{}
	oldAR, err := f.getAccessRule(ctx, api)
	if err != nil {
		return err
	}

	if oldAR != nil {
		ar = prepareAccessRule(api, oldAR, api.Spec.Rules[0], config)
		err = f.updateAccessRule(ctx, ar)
		if err != nil {
			return err
		}
	} else {
		ar = generateAccessRule(api, api.Spec.Rules[0], config)
		err = f.createAccessRule(ctx, ar)
		if err != nil {
			return err
		}
	}
	fmt.Printf("+++\n%+v\n", ar)
	return nil
}
