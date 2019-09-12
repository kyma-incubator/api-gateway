package builders

import (
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"

	internalTypes "github.com/kyma-incubator/api-gateway/internal/types/ory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8sTypes "k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Builder for", func() {

	host := "oauthkeeper.cluster.local"
	hostPath := "/.*"
	destHost := "somehost.somenamespace.svc.cluster.local"
	var destPort uint32 = 4321
	methods := []string{"GET", "POST", "PUT"}
	testScopes := []string{"read", "write"}

	Describe("AccessRule", func() {
		It("should build the object", func() {
			name := "testName"
			namespace := "testNs"

			refName := "refName"
			refVersion := "v2alpha1"
			refKind := "Gate"
			var refUID k8sTypes.UID = "123"

			testUpstreamURL := fmt.Sprintf("http://%s:%d", destHost, destPort)
			testMatchURL := fmt.Sprintf("<http|https>://%s<%s>", host, hostPath)

			requiredScopes := &internalTypes.OauthIntrospectionConfig{
				RequiredScope: testScopes,
			}
			requiredScopesJSON, _ := json.Marshal(requiredScopes)

			rawConfig := &runtime.RawExtension{
				Raw: requiredScopesJSON,
			}

			ar := AccessRule().Name(name).Namespace(namespace).
				Owner(OwnerReference().Name(refName).APIVersion(refVersion).Kind(refKind).UID(refUID).Controller(true)).
				Spec(AccessRuleSpec().
					Upstream(Upstream().
						URL(testUpstreamURL)).
					Match(Match().
						URL(testMatchURL).
						Methods(methods)).
					Authorizer(Authorizer().
						Handler(Handler().
							Name("allow"))).
					Authenticators(Authenticators().
						Handler(Handler().
							Name("oauth2_introspection").
							Config(rawConfig)).
						Handler(Handler().
							Name("jwt"))).
					Mutators(Mutators().
						Handler(Handler().
							Name("hydrator")))).
				Get()
			fmt.Printf("%#v", ar)
			Expect(ar.Name).To(Equal(name))
			Expect(ar.Namespace).To(Equal(namespace))
			Expect(ar.OwnerReferences).To(HaveLen(1))
			Expect(ar.OwnerReferences[0].Name).To(Equal(refName))
			Expect(ar.OwnerReferences[0].APIVersion).To(Equal(refVersion))
			Expect(ar.OwnerReferences[0].Kind).To(Equal(refKind))
			Expect(ar.OwnerReferences[0].UID).To(BeEquivalentTo(refUID))
			Expect(ar.Spec.Upstream.URL).To(Equal(testUpstreamURL))
			Expect(ar.Spec.Match.URL).To(Equal(testMatchURL))
			Expect(ar.Spec.Match.Methods).To(BeEquivalentTo(methods))
			Expect(ar.Spec.Authorizer.Handler.Name).To(Equal("allow"))
			Expect(len(ar.Spec.Authenticators)).To(Equal(2))
			Expect(ar.Spec.Authenticators[0].Name).To(Equal("oauth2_introspection"))
			Expect(ar.Spec.Authenticators[0].Config).To(Equal(rawConfig))
			Expect(ar.Spec.Authenticators[1].Name).To(Equal("jwt"))
			Expect(len(ar.Spec.Mutators)).To(Equal(1))
			Expect(ar.Spec.Mutators[0].Name).To(Equal("hydrator"))
		})
	})
})
