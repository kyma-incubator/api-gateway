/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	"github.com/kyma-incubator/api-gateway/controllers"
	crClients "github.com/kyma-incubator/api-gateway/internal/clients"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = gatewayv1alpha1.AddToScheme(scheme)
	_ = networkingv1alpha3.AddToScheme(scheme)
	_ = rulev1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var jwksURI string
	var oathkeeperSvcAddr string
	var oathkeeperSvcPort uint

	flag.StringVar(&oathkeeperSvcAddr, "oathkeeper-svc-address", "", "Oathkeeper proxy service")
	flag.UintVar(&oathkeeperSvcPort, "oathkeeper-svc-port", 0, "Oathkeeper proxy service port")
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&jwksURI, "jwks-uri", "", "URL of the provider's public key set to validate signature of the JWT")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	if jwksURI == "" {
		setupLog.Error(fmt.Errorf("jwksURI required, but not supplied"), "unable to create controller", "controller", "Api")
		os.Exit(1)
	}
	if oathkeeperSvcAddr == "" {
		setupLog.Error(fmt.Errorf("oathkeeper service address can't be empty"), "unable to create controller", "controller", "Api")
		os.Exit(1)
	}
	if oathkeeperSvcPort == 0 {
		setupLog.Error(fmt.Errorf("oathkeeper service port can't be empty"), "unable to create controller", "controller", "Api")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.APIReconciler{
		Client:            mgr.GetClient(),
		ExtCRClients:      crClients.New(mgr.GetClient()),
		Log:               ctrl.Log.WithName("controllers").WithName("Api"),
		OathkeeperSvc:     oathkeeperSvcAddr,
		OathkeeperSvcPort: uint32(oathkeeperSvcPort),
		JWKSURI:           jwksURI,
		Validator:         &validation.APIRule{},
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Api")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
