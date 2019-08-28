package processing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"
	oryClient "github.com/kyma-incubator/api-gateway/internal/clients/ory"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

type factory struct {
	vsClient      *istioClient.VirtualService
	apClient      *istioClient.AuthenticationPolicy
	arClient      *oryClient.AccessRule
	Log           logr.Logger
	oathkeeperSvc string
}

type ProcessingStrategy interface {
	Process(ctx context.Context, api *gatewayv2alpha1.Gate) error
}

func NewFactory(vsClient *istioClient.VirtualService, apClient *istioClient.AuthenticationPolicy, arClient *oryClient.AccessRule, logger logr.Logger, oathkeeperSvc string) *factory {
	return &factory{
		vsClient:      vsClient,
		apClient:      apClient,
		arClient:      arClient,
		Log:           logger,
		oathkeeperSvc: oathkeeperSvc,
	}
}

func (f *factory) StrategyFor(strategyName string) (ProcessingStrategy, error) {
	switch strategyName {
	case gatewayv2alpha1.PASSTHROUGH:
		f.Log.Info("PASSTHROUGH processing mode detected")
		return &passthrough{vsClient: f.vsClient}, nil
	case gatewayv2alpha1.JWT:
		f.Log.Info("JWT processing mode detected")
		return &jwt{vsClient: f.vsClient, apClient: f.apClient}, nil
	case gatewayv2alpha1.OAUTH:
		f.Log.Info("OAUTH processing mode detected")
		return &oauth{vsClient: f.vsClient, arClient: f.arClient, oathkeeperSvc: f.oathkeeperSvc}, nil
	default:
		return nil, fmt.Errorf("Unsupported mode: %s", strategyName)
	}
}
