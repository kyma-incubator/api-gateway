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
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"

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
var supportedHandlers = []string{"allow", "noop", "oauth2_introspection", "jwt"}

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

	return r.validateAPIRule()
}

func (r *APIRule) validateAPIRule() error {
	errs := r.validateAPIRuleSpec()

	if len(errs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(
		schema.GroupKind{Group: "gateway.kyma-project.io", Kind: "APIRule"},
		r.Name, errs)

	return nil
}

func (r *APIRule) validateAPIRuleSpec() field.ErrorList {
	// The field helpers from the kubernetes API machinery help us return nicely
	// structured validation errors.
	var allErrs field.ErrorList

	err := validateService(
		r.Spec.Service,
		r.Namespace,
		field.NewPath("spec").Child("service"))
	if err != nil {
		allErrs = append(allErrs, err)
	}

	if err := validateRules(r.Spec.Rules, field.NewPath("spec").Child("rules")); err != nil {
		allErrs = append(allErrs, err)
	}

	if len(allErrs) == 0 {
		return nil
	}
	return allErrs
}

func validateService(svc *Service, namespace string, fldPath *field.Path) *field.Error {
	k8sSvc, err := getSvc(*svc.Name, namespace)
	if err != nil {
		return field.Invalid(fldPath, svc, err.Error())
	}

	return validateSvcType(k8sSvc, svc, fldPath)
}

func validateRules(rules []Rule, fldPath *field.Path) *field.Error {

	for _, rule := range rules {
		err := validateAccessStrategies(fldPath, rule.AccessStrategies)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateAccessStrategies(fldPath *field.Path, accStrategies []*rulev1alpha1.Authenticator) *field.Error {
	for _, accStrategy := range accStrategies {
		if !includedIn(accStrategy.Handler.Name, supportedHandlers) {
			return field.NotSupported(fldPath.Child("handler"), accStrategy.Handler, supportedHandlers)
		}
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *APIRule) ValidateUpdate(old runtime.Object) error {
	apirulelog.Info("validate update", "name", r.Name)

	return r.validateAPIRule()
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
		return field.Invalid(fldPath, svc.Name, fmt.Sprintf("service can't be type of %s", v1.ServiceTypeExternalName))
	}
	return nil
}

func includedIn(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
