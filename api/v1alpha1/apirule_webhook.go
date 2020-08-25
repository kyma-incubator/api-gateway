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

package v1alpha1

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var apirulelog = logf.Log.WithName("apirule-resource")

// SetupWebhookWithManager .
func (r *APIRule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-gateway-kyma-project-io-v1alpha1-apirule,mutating=false,failurePolicy=fail,groups=gateway.kyma-project.io,resources=apirules,versions=v1alpha1,name=vapirule.kb.io

var _ webhook.Validator = &APIRule{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *APIRule) ValidateCreate() error {
	apirulelog.Info("validate create", "name", r.Name)

	var allErrs field.ErrorList
	if err := r.validateAPIRuleSpec(); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "gateway.kyma-project.io", Kind: "APIRule"},
		r.Name, allErrs)

	return nil
}

func (r *APIRule) validateAPIRuleSpec() *field.Error {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	return validateService(
		r.Spec.Service,
		r.Namespace,
		field.NewPath("spec").Child("service"))
}

func validateService(svc *Service, namespace string, fldPath *field.Path) *field.Error {
	k8sSvc, err := getSvc(*svc.Name, namespace)
	if err != nil {
		return field.Invalid(fldPath, svc, err.Error())
	}

	return validateSvcType(k8sSvc, svc, fldPath)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *APIRule) ValidateUpdate(old runtime.Object) error {
	apirulelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *APIRule) ValidateDelete() error {
	apirulelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func getClient() (*kubernetes.Clientset, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, err
	}

	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}
	return k8sClient, nil
}

func getSvc(name, namespace string) (*v1.Service, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	return client.CoreV1().Services(namespace).Get(name, metav1.GetOptions{})
}

func validateSvcType(k8sSvc *v1.Service, svc *Service, fldPath *field.Path) *field.Error {
	if k8sSvc.Spec.Type == v1.ServiceTypeExternalName {
		return field.Invalid(fldPath, svc, fmt.Sprintf("service can't be type of %s", v1.ServiceTypeExternalName))
	}
	return nil
}
