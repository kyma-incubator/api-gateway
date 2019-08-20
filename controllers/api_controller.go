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

package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/internal/validation"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ApiReconciler reconciles a Api object
type ApiReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apis,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=gateway.kyma-project.io,resources=apis/status,verbs=get;update;patch

func (r *ApiReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("api", req.NamespacedName)

	api := &gatewayv2alpha1.Api{}

	err := r.Get(context.TODO(), req.NamespacedName, api)
	if err != nil {
		if !apierrs.IsNotFound(err) {
			return reconcile.Result{}, err
		}
	}

	APIStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code: gatewayv2alpha1.STATUS_OK,
	}

	virtualServiceStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code:        gatewayv2alpha1.STATUS_SKIPPED,
		Description: "Skipped setting Istio Virtual Service",
	}
	policyStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code:        gatewayv2alpha1.STATUS_SKIPPED,
		Description: "Skipped setting Istio Policy",
	}

	accessRuleStatus := &gatewayv2alpha1.GatewayResourceStatus{
		Code:        gatewayv2alpha1.STATUS_SKIPPED,
		Description: "Skipped setting Oathkeeper Access Rule",
	}

	if api.Generation != api.Status.ObservedGeneration {
		r.Log.Info("Api processing")
		switch *api.Spec.Auth.Name {
		case gatewayv2alpha1.PASSTHROUGH:
			r.Log.Info("PASSTHROUGH mode detected")
			err := validation.ValidatePassthroughMode(api.Spec.Auth.Config)
			if err != nil {
				r.updateStatus(api, &gatewayv2alpha1.GatewayResourceStatus{Code: gatewayv2alpha1.STATUS_ERROR}, virtualServiceStatus, policyStatus, accessRuleStatus)
				return ctrl.Result{}, err
			}
		case gatewayv2alpha1.JWT:
			r.Log.Info("JWT mode detected")
		case gatewayv2alpha1.OAUTH:
			r.Log.Info("OAUTH mode detected")
			err := validation.ValidateOauthMode(api.Spec.Auth.Config)
			if err != nil {
				r.updateStatus(api, &gatewayv2alpha1.GatewayResourceStatus{Code: gatewayv2alpha1.STATUS_ERROR}, virtualServiceStatus, policyStatus, accessRuleStatus)
				return ctrl.Result{}, err
			}
		default:
			err := fmt.Errorf("Unsupported mode: %s", *api.Spec.Auth.Name)
			return ctrl.Result{}, err
		}
		_, err = r.updateStatus(api, APIStatus, virtualServiceStatus, policyStatus, accessRuleStatus)

		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}
	}

	// demo sample fetching virtualservices

	//list := networkingv1alpha3.VirtualServiceList{}
	//err = r.Client.List(context.TODO(), &list, client.InNamespace(req.Namespace))
	//if err != nil {
	//	fmt.Printf("ooops, error occured when fetching vs " + err.Error())
	//	os.Exit(1)
	//}
	//
	//fmt.Println(list)

	return ctrl.Result{}, nil
}

func (r *ApiReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv2alpha1.Api{}).
		Complete(r)
}

func (r *ApiReconciler) updateStatus(api *gatewayv2alpha1.Api, APIStatus, virtualServiceStatus, policyStatus, accessRuleStatus *gatewayv2alpha1.GatewayResourceStatus) (*gatewayv2alpha1.Api, error) {
	copy := api.DeepCopy()

	copy.Status.ObservedGeneration = api.Generation
	copy.Status.LastProcessedTime = &v1.Time{Time: time.Now()}
	copy.Status.APIStatus = APIStatus
	copy.Status.VirtualServiceStatus = virtualServiceStatus
	copy.Status.PolicyServiceStatus = policyStatus
	copy.Status.AccessRuleStatus = accessRuleStatus

	err := r.Status().Update(context.TODO(), copy)
	if err != nil {
		return nil, err
	}
	return copy, nil
}
