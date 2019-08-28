package controllers_test

import (
	"context"
	"fmt"
	"time"

	gatewayv2alpha1 "github.com/kyma-incubator/api-gateway/api/v2alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	networkingv1alpha3 "knative.dev/pkg/apis/istio/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: "test", Namespace: "default"}}

const timeout = time.Second * 5

var _ = Describe("Gate Controller", func() {
	var tstNamespace = "default"

	Context("in a happy-path scenario", func() {
		It("should create a VirtualService and an AccessRule", func() {

			s := runtime.NewScheme()
			err := gatewayv2alpha1.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())

			err = networkingv1alpha3.AddToScheme(s)
			Expect(err).NotTo(HaveOccurred())
			// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
			// channel when it is finished.
			mgr, err := manager.New(cfg, manager.Options{Scheme: s})
			Expect(err).NotTo(HaveOccurred())
			c := mgr.GetClient()

			recFn, requests := SetupTestReconcile(getAPIReconciler(mgr))

			Expect(add(mgr, recFn)).To(Succeed())

			//Start the manager and the controller
			stopMgr, mgrStopped := StartTestManager(mgr)

			//Ensure manager is stopped properly
			defer func() {
				close(stopMgr)
				mgrStopped.Wait()
			}()

			instance := testInstance()
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
			vs := networkingv1alpha3.VirtualService{}
			err = c.Get(context.TODO(), client.ObjectKey{Name: "test-test", Namespace: "default"}, &vs)
			Expect(err).NotTo(HaveOccurred())
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

func testInstance(name, namespace string) *gatewayv2alpha1.Gate {
	serviceName = "test"
	servicePort = 8000
	host = "foo.bar"
	isExernal = false
	authStrategy = gatewayv2alpha1.PASSTHROUGH
	gateway = "some-gateway.some-namespace.foo"

	return &gatewayv2alpha1.Gate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: gatewayv2alpha1.GateSpec{
			Service: &gatewayv2alpha1.Service{
				Name: "httpbin",
				Port: 8000,
				Host: "httpbin.kyma.local",
			},
			Auth: &gatewayv2alpha1.AuthStrategy{
				Name:   &authStrategy,
				Config: nil,
			},
			Gateway: &gateway,
		},
	}
}
