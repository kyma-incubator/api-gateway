package processing

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	istioClient "github.com/kyma-incubator/api-gateway/internal/clients/istio"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
)

type factory struct {
	vsClient *istioClient.VirtualService
	apClient *istioClient.AuthenticationPolicy
	Log      logr.Logger
	JWKSURI  string
}

type ProcessingStrategy interface {
	Process(ctx context.Context, api *gatewayv2alpha1.Gate) error
}

func NewFactory(vsClient *istioClient.VirtualService, apClient *istioClient.AuthenticationPolicy, logger logr.Logger, jwksURI string) *factory {
	return &factory{
		vsClient: vsClient,
		apClient: apClient,
		Log:      logger,
		JWKSURI:  jwksURI,
	}
}

func (f *factory) StrategyFor(strategyName string) (ProcessingStrategy, error) {
	switch strategyName {
	case gatewayv2alpha1.PASSTHROUGH:
		f.Log.Info("PASSTHROUGH processing mode detected")
		return &passthrough{vsClient: f.vsClient}, nil
	case gatewayv2alpha1.JWT:
		f.Log.Info("JWT processing mode detected")
		return &jwt{vsClient: f.vsClient, apClient: f.apClient, JWKSURI: f.JWKSURI}, nil
	default:
		return nil, fmt.Errorf("Unsupported mode: %s", strategyName)
	}
}
