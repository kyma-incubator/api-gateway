package validation

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helpers Suite")
}

var _ = Describe("ValidateDomainName function", func() {

	It("Should return true for valid simple domain", func() {
		//given
		testDomain := "kyma.local"

		//when
		valid := ValidateDomainName(testDomain)

		//then
		Expect(valid).To(BeTrue())
	})

	It("Should return true for valid complicated domain", func() {
		//given
		testDomain := "gke-upgrade-pr-5776-47nlgu1ch0.a.build.kyma-project.io"

		//when
		valid := ValidateDomainName(testDomain)

		//then
		Expect(valid).To(BeTrue())
	})

	It("Should return false for invalid domain", func() {
		//given
		testDomain := "notdomain"

		//when
		valid := ValidateDomainName(testDomain)

		//then
		Expect(valid).To(BeFalse())
	})
})
