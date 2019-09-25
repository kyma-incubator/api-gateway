package processing

import (
	"fmt"
	//"github.com/kyma-incubator/api-gateway/internal/processing"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"testing"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	//"github.com/kyma-incubator/api-gateway/internal/processing"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	apiName                     = "test-apirule"
	apiUID            types.UID = "eab0f1c8-c417-11e9-bf11-4ac644044351"
	apiNamespace                = "some-namespace"
	apiAPIVersion               = "gateway.kyma-project.io/v1alpha1"
	apiKind                     = "ApiRule"
	apiGateway                  = "some-gateway"
	apiPath                     = "/.*"
	jwtApiPath                  = "/headers"
	apiMethods                  = []string{"GET"}
	serviceName                 = "example-service"
	serviceHost                 = "myService.myDomain.com"
	servicePort       uint32    = 8080
	apiScopes                   = []string{"write", "read"}
	jwtIssuer                   = "https://oauth2.example.com/"
	oathkeeperSvc               = "fake.oathkeeper"
	oathkeeperSvcPort uint32    = 1234
)

func TestProcessing(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Processing Suite")
}

var _ = Describe("Factory", func() {
	Describe("CalculateRequiredState", func() {
		Context("APIRule", func() {
			It("should produce VS for allow authenticator", func() {
				strategies := []*rulev1alpha1.Authenticator{
					{
						Handler: &rulev1alpha1.Handler{
							Name: "allow",
						},
					},
				}

				allowRule := gatewayv1alpha1.Rule{
					Path:             apiPath,
					Methods:          apiMethods,
					Mutators:         []*rulev1alpha1.Mutator{},
					AccessStrategies: strategies,
				}

				rules := []gatewayv1alpha1.Rule{allowRule}

				apiRule := getAPIRuleFor(rules)

				f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json")

				desiredState := f.CalculateRequiredState(apiRule)
				vs := desiredState.virtualService
				accessRules := desiredState.accessRules

				//verify VS
				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(vs.Spec.HTTP)).To(Equal(1))

				Expect(len(vs.Spec.HTTP[0].Route)).To(Equal(1))
				Expect(vs.Spec.HTTP[0].Route[0].Destination.Host).To(Equal(serviceName + "." + apiNamespace + ".svc.cluster.local"))
				Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Number).To(Equal(servicePort))

				Expect(len(vs.Spec.HTTP[0].Match)).To(Equal(1))
				Expect(vs.Spec.HTTP[0].Match[0].URI.Regex).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(vs.ObjectMeta.Name).To(BeEmpty())
				Expect(vs.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(vs.ObjectMeta.Namespace).To(Equal(apiNamespace))

				Expect(vs.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(vs.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(vs.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(vs.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))

				//Verify AR
				Expect(len(accessRules)).To(Equal(0))
			})

			It("should produce VS and ARs for given paths", func() {
				noop := []*rulev1alpha1.Authenticator{
					{
						Handler: &rulev1alpha1.Handler{
							Name: "noop",
						},
					},
				}

				noopRule := gatewayv1alpha1.Rule{
					Path:             apiPath,
					Methods:          apiMethods,
					Mutators:         []*rulev1alpha1.Mutator{},
					AccessStrategies: noop,
				}

				jwtConfigJSON := fmt.Sprintf(`
					{
						"trusted_issuers": ["%s"],
						"jwks": [],
						"required_scope": [%s]
				}`, jwtIssuer, toCSVList(apiScopes))

				jwt := []*rulev1alpha1.Authenticator{
					{
						Handler: &rulev1alpha1.Handler{
							Name: "jwt",
							Config: &runtime.RawExtension{
								Raw: []byte(jwtConfigJSON),
							},
						},
					},
				}

				testMutators := []*rulev1alpha1.Mutator{
					{
						Handler: &rulev1alpha1.Handler{
							Name: "noop",
						},
					},
					{
						Handler: &rulev1alpha1.Handler{
							Name: "idtoken",
						},
					},
				}

				jwtRule := gatewayv1alpha1.Rule{
					Path:             jwtApiPath,
					Methods:          apiMethods,
					Mutators:         testMutators,
					AccessStrategies: jwt,
				}

				rules := []gatewayv1alpha1.Rule{noopRule, jwtRule}

				expectedNoopRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, apiPath)
				expectedJwtRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, jwtApiPath)
				expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, apiNamespace, servicePort)

				apiRule := getAPIRuleFor(rules)

				f := NewFactory(nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json")

				desiredState := f.CalculateRequiredState(apiRule)
				vs := desiredState.virtualService
				accessRules := desiredState.accessRules

				//verify VS
				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(vs.Spec.HTTP)).To(Equal(2))

				Expect(len(vs.Spec.HTTP[0].Route)).To(Equal(1))
				Expect(vs.Spec.HTTP[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))
				Expect(len(vs.Spec.HTTP[0].Match)).To(Equal(1))
				Expect(vs.Spec.HTTP[0].Match[0].URI.Regex).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(len(vs.Spec.HTTP[1].Route)).To(Equal(1))
				Expect(vs.Spec.HTTP[1].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.HTTP[1].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))
				Expect(len(vs.Spec.HTTP[1].Match)).To(Equal(1))
				Expect(vs.Spec.HTTP[1].Match[0].URI.Regex).To(Equal(apiRule.Spec.Rules[1].Path))

				Expect(vs.ObjectMeta.Name).To(BeEmpty())
				Expect(vs.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(vs.ObjectMeta.Namespace).To(Equal(apiNamespace))

				Expect(vs.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(vs.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(vs.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(vs.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))

				//Verify ARs
				Expect(len(accessRules)).To(Equal(2))

				noopAccessRule := accessRules[expectedNoopRuleMatchURL]

				Expect(len(accessRules)).To(Equal(2))
				Expect(len(noopAccessRule.Spec.Authenticators)).To(Equal(1))

				Expect(noopAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
				Expect(noopAccessRule.Spec.Authorizer.Config).To(BeNil())

				Expect(noopAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("noop"))
				Expect(noopAccessRule.Spec.Authenticators[0].Handler.Config).To(BeNil())

				Expect(len(noopAccessRule.Spec.Match.Methods)).To(Equal(len(apiMethods)))
				Expect(noopAccessRule.Spec.Match.Methods).To(Equal(apiMethods))
				Expect(noopAccessRule.Spec.Match.URL).To(Equal(expectedNoopRuleMatchURL))

				Expect(noopAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))

				Expect(noopAccessRule.ObjectMeta.Name).To(BeEmpty())
				Expect(noopAccessRule.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(noopAccessRule.ObjectMeta.Namespace).To(Equal(apiNamespace))

				Expect(noopAccessRule.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(noopAccessRule.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(noopAccessRule.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(noopAccessRule.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))

				jwtAccessRule := accessRules[expectedJwtRuleMatchURL]

				Expect(len(jwtAccessRule.Spec.Authenticators)).To(Equal(1))

				Expect(jwtAccessRule.Spec.Authorizer.Name).To(Equal("allow"))
				Expect(jwtAccessRule.Spec.Authorizer.Config).To(BeNil())

				Expect(jwtAccessRule.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
				Expect(jwtAccessRule.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
				Expect(string(jwtAccessRule.Spec.Authenticators[0].Handler.Config.Raw)).To(Equal(jwtConfigJSON))

				Expect(len(jwtAccessRule.Spec.Match.Methods)).To(Equal(len(apiMethods)))
				Expect(jwtAccessRule.Spec.Match.Methods).To(Equal(apiMethods))
				Expect(jwtAccessRule.Spec.Match.URL).To(Equal(expectedJwtRuleMatchURL))

				Expect(jwtAccessRule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))

				Expect(jwtAccessRule.Spec.Mutators).NotTo(BeNil())
				Expect(len(jwtAccessRule.Spec.Mutators)).To(Equal(len(testMutators)))
				Expect(jwtAccessRule.Spec.Mutators[0].Handler.Name).To(Equal(testMutators[0].Name))
				Expect(jwtAccessRule.Spec.Mutators[1].Handler.Name).To(Equal(testMutators[1].Name))

				Expect(jwtAccessRule.ObjectMeta.Name).To(BeEmpty())
				Expect(jwtAccessRule.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(jwtAccessRule.ObjectMeta.Namespace).To(Equal(apiNamespace))

				Expect(jwtAccessRule.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(jwtAccessRule.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(jwtAccessRule.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(jwtAccessRule.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))
			})

			It("should produce VS & AR for jwt & oauth authenticators for given path", func() {
				oauthConfigJSON := fmt.Sprintf(`{"required_scope": [%s]}`, toCSVList(apiScopes))

				jwtConfigJSON := fmt.Sprintf(`
					{
						"trusted_issuers": ["%s"],
						"jwks": [],
						"required_scope": [%s]
				}`, jwtIssuer, toCSVList(apiScopes))

				jwt := &rulev1alpha1.Authenticator{
					Handler: &rulev1alpha1.Handler{
						Name: "jwt",
						Config: &runtime.RawExtension{
							Raw: []byte(jwtConfigJSON),
						},
					},
				}
				oauth := &rulev1alpha1.Authenticator{
					Handler: &rulev1alpha1.Handler{
						Name: "oauth2_introspection",
						Config: &runtime.RawExtension{
							Raw: []byte(oauthConfigJSON),
						},
					},
				}

				strategies := []*rulev1alpha1.Authenticator{jwt, oauth}

				allowRule := gatewayv1alpha1.Rule{
					Path:             apiPath,
					Methods:          apiMethods,
					Mutators:         []*rulev1alpha1.Mutator{},
					AccessStrategies: strategies,
				}

				rules := []gatewayv1alpha1.Rule{allowRule}

				expectedRuleMatchURL := fmt.Sprintf("<http|https>://%s<%s>", serviceHost, apiPath)
				expectedRuleUpstreamURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", serviceName, apiNamespace, servicePort)

				apiRule := getAPIRuleFor(rules)

				f := NewFactory(  nil, ctrl.Log.WithName("test"), oathkeeperSvc, oathkeeperSvcPort, "https://example.com/.well-known/jwks.json")

				desiredState := f.CalculateRequiredState(apiRule)
				vs := desiredState.virtualService
				accessRules := desiredState.accessRules

				//verify VS
				Expect(vs).NotTo(BeNil())
				Expect(len(vs.Spec.Gateways)).To(Equal(1))
				Expect(len(vs.Spec.Hosts)).To(Equal(1))
				Expect(vs.Spec.Hosts[0]).To(Equal(serviceHost))
				Expect(len(vs.Spec.HTTP)).To(Equal(1))

				Expect(len(vs.Spec.HTTP[0].Route)).To(Equal(1))
				Expect(vs.Spec.HTTP[0].Route[0].Destination.Host).To(Equal(oathkeeperSvc))
				Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Number).To(Equal(oathkeeperSvcPort))

				Expect(len(vs.Spec.HTTP[0].Match)).To(Equal(1))
				Expect(vs.Spec.HTTP[0].Match[0].URI.Regex).To(Equal(apiRule.Spec.Rules[0].Path))

				Expect(vs.ObjectMeta.Name).To(BeEmpty())
				Expect(vs.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(vs.ObjectMeta.Namespace).To(Equal(apiNamespace))

				Expect(vs.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(vs.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(vs.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(vs.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))

				rule := accessRules[expectedRuleMatchURL]

				//Verify AR
				Expect(len(accessRules)).To(Equal(1))
				Expect(len(rule.Spec.Authenticators)).To(Equal(2))

				Expect(rule.Spec.Authorizer.Name).To(Equal("allow"))
				Expect(rule.Spec.Authorizer.Config).To(BeNil())

				Expect(rule.Spec.Authenticators[0].Handler.Name).To(Equal("jwt"))
				Expect(rule.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
				Expect(string(rule.Spec.Authenticators[0].Handler.Config.Raw)).To(Equal(jwtConfigJSON))

				Expect(rule.Spec.Authenticators[1].Handler.Name).To(Equal("oauth2_introspection"))
				Expect(rule.Spec.Authenticators[1].Handler.Config).NotTo(BeNil())
				Expect(string(rule.Spec.Authenticators[1].Handler.Config.Raw)).To(Equal(oauthConfigJSON))

				Expect(len(rule.Spec.Match.Methods)).To(Equal(len(apiMethods)))
				Expect(rule.Spec.Match.Methods).To(Equal(apiMethods))
				Expect(rule.Spec.Match.URL).To(Equal(expectedRuleMatchURL))

				Expect(rule.Spec.Upstream.URL).To(Equal(expectedRuleUpstreamURL))

				Expect(rule.ObjectMeta.Name).To(BeEmpty())
				Expect(rule.ObjectMeta.GenerateName).To(Equal(apiName + "-"))
				Expect(rule.ObjectMeta.Namespace).To(Equal(apiNamespace))

				Expect(rule.ObjectMeta.OwnerReferences[0].APIVersion).To(Equal(apiAPIVersion))
				Expect(rule.ObjectMeta.OwnerReferences[0].Kind).To(Equal(apiKind))
				Expect(rule.ObjectMeta.OwnerReferences[0].Name).To(Equal(apiName))
				Expect(rule.ObjectMeta.OwnerReferences[0].UID).To(Equal(apiUID))

			})
		})
	})
})

func getAPIRuleFor(rules []gatewayv1alpha1.Rule) *gatewayv1alpha1.APIRule {
	return &gatewayv1alpha1.APIRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      apiName,
			UID:       apiUID,
			Namespace: apiNamespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: apiAPIVersion,
			Kind:       apiKind,
		},
		Spec: gatewayv1alpha1.APIRuleSpec{
			Gateway: &apiGateway,
			Service: &gatewayv1alpha1.Service{
				Name: &serviceName,
				Host: &serviceHost,
				Port: &servicePort,
			},
			Rules: rules,
		},
	}
}

func toCSVList(input []string) string {
	if len(input) == 0 {
		return ""
	}

	res := `"` + input[0] + `"`

	for i := 1; i < len(input); i++ {
		res = res + "," + `"` + input[i] + `"`
	}

	return res
}
