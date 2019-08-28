package controllers_test

import (
	"context"
	"fmt"
	"time"

	"encoding/json"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	"github.com/kyma-incubator/api-gateway/controllers"
	crClients "github.com/kyma-incubator/api-gateway/internal/clients"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	rulev1alpha1 "github.com/ory/oathkeeper-maester/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const timeout = time.Second * 5

const tstGateway = "kyma-gateway.kyma-system.svc.cluster.local"
const tstOathkeeperSvc = "oathkeeper.kyma-system.svc.cluster.local:1234"
const tstNamespace = "padu-system"
const tstName = "test"

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: tstName, Namespace: tstNamespace}}

var _ = Describe("Gate Controller", func() {
	const tstServiceName = "httpbin"
	const tstServiceHost = "httpbin.kyma.local"
	const tstServicePort uint32 = 443
	const tstPath = "/.*"
	var tstMethods = []string{"GET", "PUT"}
	var tstScopes = []string{"foo", "bar"}

	Context("in a happy-path scenario", func() {
		It("should create a VirtualService and an AccessRule", func() {

			s := runtime.NewScheme()

			err := rulev1alpha1.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())

			err = gatewayv2alpha1.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())

			err = networkingv1alpha3.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())

			// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
			// channel when it is finished.
			mgr, err := manager.New(cfg, manager.Options{Scheme: s})
			Expect(err).NotTo(HaveOccurred())
			c := mgr.GetClient()

			reconciler := &controllers.ApiReconciler{
				Client:        mgr.GetClient(),
				ExtCRClients:  crClients.New(mgr.GetClient()),
				Log:           ctrl.Log.WithName("controllers").WithName("Gate"),
				OathkeeperSvc: tstOathkeeperSvc,
			}

			recFn, requests := SetupTestReconcile(reconciler)

			Expect(add(mgr, recFn)).To(Succeed())

			//Start the manager and the controller
			stopMgr, mgrStopped := StartTestManager(mgr)

			//Ensure manager is stopped properly
			defer func() {
				close(stopMgr)
				mgrStopped.Wait()
			}()

			instance := testInstance(tstName, tstNamespace, tstServiceName, tstServiceHost, int32(tstServicePort), tstPath, tstMethods, tstScopes)

			err = c.Create(context.TODO(), instance)
			// The instance object may not be a valid object because it might be missing some required fields.
			// Please modify the instance object by adding required fields and then remove the following if statement.
			if apierrors.IsInvalid(err) {
				Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
				return
			}
			Expect(err).NotTo(HaveOccurred())
			defer c.Delete(context.TODO(), instance)

			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			//Verify VirtualService
			expectedVSName := tstName + "-" + tstServiceName
			expectedVSNamespace := tstNamespace
			vs := networkingv1alpha3.VirtualService{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: expectedVSName, Namespace: expectedVSNamespace}, &vs)
			Expect(err).NotTo(HaveOccurred())

			/*
				jsn, err := json.MarshalIndent(vs, "", "  ")
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("\n----------------------------------------\n%s", jsn)
			*/

			//Meta
			verifyOwnerReference(vs.ObjectMeta, tstName, gatewayv2alpha1.GroupVersion.String(), "Gate")
			//Spec.Hosts
			Expect(vs.Spec.Hosts).To(HaveLen(1))
			Expect(vs.Spec.Hosts[0]).To(Equal(tstServiceHost))
			//Spec.Gateways
			Expect(vs.Spec.Gateways).To(HaveLen(1))
			Expect(vs.Spec.Gateways[0]).To(Equal(tstGateway))
			//Spec.HTTP
			Expect(vs.Spec.HTTP).To(HaveLen(1))
			////// HTTP.Match[]
			Expect(vs.Spec.HTTP[0].Match).To(HaveLen(1))
			/////////// Match[].URI
			Expect(vs.Spec.HTTP[0].Match[0].URI).NotTo(BeNil())
			Expect(vs.Spec.HTTP[0].Match[0].URI.Exact).To(BeEmpty())
			Expect(vs.Spec.HTTP[0].Match[0].URI.Prefix).To(BeEmpty())
			Expect(vs.Spec.HTTP[0].Match[0].URI.Suffix).To(BeEmpty())
			Expect(vs.Spec.HTTP[0].Match[0].URI.Regex).To(Equal(tstPath))
			Expect(vs.Spec.HTTP[0].Match[0].Scheme).To(BeNil())
			Expect(vs.Spec.HTTP[0].Match[0].Method).To(BeNil())
			Expect(vs.Spec.HTTP[0].Match[0].Authority).To(BeNil())
			Expect(vs.Spec.HTTP[0].Match[0].Headers).To(BeNil())
			Expect(vs.Spec.HTTP[0].Match[0].Port).To(BeZero())
			Expect(vs.Spec.HTTP[0].Match[0].SourceLabels).To(BeNil())
			Expect(vs.Spec.HTTP[0].Match[0].Gateways).To(BeNil())
			////// HTTP.Route[]
			Expect(vs.Spec.HTTP[0].Route).To(HaveLen(1))
			Expect(vs.Spec.HTTP[0].Route[0].Destination.Host).To(Equal(tstOathkeeperSvc))
			Expect(vs.Spec.HTTP[0].Route[0].Destination.Subset).To(Equal(""))
			Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Name).To(Equal(""))
			Expect(vs.Spec.HTTP[0].Route[0].Destination.Port.Number).To(Equal(tstServicePort))
			Expect(vs.Spec.HTTP[0].Route[0].Weight).To(BeZero())
			Expect(vs.Spec.HTTP[0].Route[0].Headers).To(BeNil())
			//Others
			Expect(vs.Spec.HTTP[0].Rewrite).To(BeNil())
			Expect(vs.Spec.HTTP[0].WebsocketUpgrade).To(BeFalse())
			Expect(vs.Spec.HTTP[0].Timeout).To(BeEmpty())
			Expect(vs.Spec.HTTP[0].Retries).To(BeNil())
			Expect(vs.Spec.HTTP[0].Fault).To(BeNil())
			Expect(vs.Spec.HTTP[0].Mirror).To(BeNil())
			Expect(vs.Spec.HTTP[0].DeprecatedAppendHeaders).To(BeNil())
			Expect(vs.Spec.HTTP[0].Headers).To(BeNil())
			Expect(vs.Spec.HTTP[0].RemoveResponseHeaders).To(BeNil())
			Expect(vs.Spec.HTTP[0].CorsPolicy).To(BeNil())
			//Spec.TCP
			Expect(vs.Spec.TCP).To(BeNil())
			//Spec.TLS
			Expect(vs.Spec.TLS).To(BeNil())

			//Verify Rule
			expectedRuleName := tstName + "-" + tstServiceName
			expectedRuleNamespace := tstNamespace
			rl := rulev1alpha1.Rule{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: expectedRuleName, Namespace: expectedRuleNamespace}, &rl)
			Expect(err).NotTo(HaveOccurred())

			/*
				jsn, err = json.MarshalIndent(rl, "", "  ")
				Expect(err).NotTo(HaveOccurred())
				fmt.Printf("\n----------------------------------------\n%s", jsn)
			*/

			//Meta
			verifyOwnerReference(rl.ObjectMeta, tstName, gatewayv2alpha1.GroupVersion.String(), "Gate")

			//Spec.Upstream
			Expect(rl.Spec.Upstream).NotTo(BeNil())
			Expect(rl.Spec.Upstream.URL).To(Equal(fmt.Sprintf("http://%s.%s.svc.cluster.local:%d", tstServiceName, tstNamespace, tstServicePort)))
			Expect(rl.Spec.Upstream.StripPath).To(BeNil())
			Expect(rl.Spec.Upstream.PreserveHost).To(BeNil())
			//Spec.Match
			Expect(rl.Spec.Match).NotTo(BeNil())
			Expect(rl.Spec.Match.URL).To(Equal(fmt.Sprintf("<http|https>://%s%s", tstServiceHost, tstPath)))
			Expect(rl.Spec.Match.Methods).To(Equal(tstMethods))
			//Spec.Authenticators
			Expect(rl.Spec.Authenticators).To(HaveLen(1))
			Expect(rl.Spec.Authenticators[0].Handler).NotTo(BeNil())
			Expect(rl.Spec.Authenticators[0].Handler.Name).To(Equal("oauth2_introspection"))
			Expect(rl.Spec.Authenticators[0].Handler.Config).NotTo(BeNil())
			//             Authenticators[0].Handler.Config validation
			handlerConfig := map[string]interface{}{}
			err = json.Unmarshal(rl.Spec.Authenticators[0].Config.Raw, &handlerConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(handlerConfig).To(HaveLen(1))
			Expect(asStringSlice(handlerConfig["required_scope"])).To(BeEquivalentTo(tstScopes))
			//Spec.Authorizer
			Expect(rl.Spec.Authorizer).NotTo(BeNil())
			Expect(rl.Spec.Authorizer.Handler).NotTo(BeNil())
			Expect(rl.Spec.Authorizer.Handler.Name).To(Equal("allow"))
			Expect(rl.Spec.Authorizer.Handler.Config).To(BeNil())

			//Spec.Mutators
			Expect(rl.Spec.Mutators).To(BeNil())
			//
		})
	})
})

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("api-gateway-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Api
	err = c.Watch(&source.Kind{Type: &gatewayv2alpha1.Gate{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Guestbook - change this for objects you create
	//err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &webappv1.Guestbook{},
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}

func testInstance(name, namespace, serviceName, serviceHost string, servicePort int32, path string, methods, scopes []string) *gatewayv2alpha1.Gate {

	toJSONList := func(input []string) string {
		if len(input) == 0 {
			return ""
		}

		res := `"` + input[0] + `"`

		for i := 1; i < len(input); i++ {
			res = res + "," + `"` + input[i] + `"`
		}

		return res
	}

	configJSON := fmt.Sprintf(`
{
	"paths":[
		{
			"path": "%s",
			"scopes": [%s],
			"methods": [%s]
		}
	]
}`, path, toJSONList(scopes), toJSONList(methods))

	rawCfg := &runtime.RawExtension{
		Raw: []byte(configJSON),
	}

	var gateway = tstGateway
	var authStrategyName = gatewayv2alpha1.OAUTH

	return &gatewayv2alpha1.Gate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv2alpha1.GateSpec{
			Gateway: &gateway,
			Service: &gatewayv2alpha1.Service{
				Host: &serviceHost,
				Name: &serviceName,
				Port: &servicePort,
			},
			Auth: &gatewayv2alpha1.AuthStrategy{
				Name:   &authStrategyName,
				Config: rawCfg,
			},
		},
	}
}

func verifyOwnerReference(m metav1.ObjectMeta, name, version, kind string) {
	Expect(m.OwnerReferences).To(HaveLen(1))
	Expect(m.OwnerReferences[0].APIVersion).To(Equal(version))
	Expect(m.OwnerReferences[0].Kind).To(Equal(kind))
	Expect(m.OwnerReferences[0].Name).To(Equal(name))
	Expect(m.OwnerReferences[0].UID).NotTo(BeEmpty())
	Expect(*m.OwnerReferences[0].Controller).To(BeTrue())
}

//Converts a []interface{} to a string slice. Panics if given object is of other type.
func asStringSlice(in interface{}) []string {

	inSlice := in.([]interface{})

	if inSlice == nil {
		return nil
	}

	res := []string{}

	for _, v := range inSlice {
		res = append(res, v.(string))
	}

	return res
}
