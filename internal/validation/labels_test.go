package validation

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

func TestLabels(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Labels Suite")
}

func entry(value string, valid bool) TableEntry {
	return Entry(value, value, valid)
}

var _ = Describe("Label validation function: valideLabelKey() should correctly", func() {
	DescribeTable("recognize generally invalid labels",
		func(labelKey string, shouldBeValid bool) {
			err := validateLabelKey(labelKey)
			Expect(err == nil).To(Equal(shouldBeValid))
		},
		entry("", false),
		entry("  ", false),
		entry("/a", false),
		entry("a/", false),
		entry("a/1/b", false),
		entry("//", false),
		entry("a /b", false),
		entry("a/ b", false),
		entry("a / b", false),
		entry("label-with-318-characters-is-to-long-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-e", false),
		entry("label-with-317-characters-is-ok-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-x/s_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa-e", true),
	)

	DescribeTable("validate label keys with only \"name\" part",
		func(labelKey string, shouldBeValid bool) {
			err := validateLabelKey(labelKey)
			Expect(err == nil).To(Equal(shouldBeValid))
		},
		entry("a1b", true),
		entry("aaa", true),
		entry("1aa", true),
		entry("2.a-a", true),
		entry("a.3_a", true),
		entry("s-CAPITAL-L3TT3RS-ARE-0K", true),
		entry("2.a=a", false), //Invalid character: "="
		entry("LEN_64_IS_TOO_LONG-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", false),
		entry("len_63_is_ok-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", true),
		entry("ENDS-WITH-DASH-------------------------------------------------", false),
		entry("many-dashes-are-allowed---------------------------------------b", true),
		entry("many-dots-are-allowed-also---..............-------------------b", true),
	)

	DescribeTable("validate label keys with \"prefix\" part",
		func(labelKey string, shouldBeValid bool) {
			err := validateLabelKey(labelKey)
			Expect(err == nil).To(Equal(shouldBeValid))
		},
		entry("a/b", true),
		entry(" a/b ", true),
		entry("a1/a1b", true),
		entry("abc.def/aaa", true),
		entry("a.b/a", true),
		entry("a.b.c/a", true),
		entry("a-b.c/a", true),
		entry("a-b-.c/a-DASH-DOT-SEQUENCE-IN-PREFIX", false),
		entry("a.b.c-/a-DASH-AS-LAST-CHARACTER-IN-PREFIX", false),
		entry("a_b.c/a-UNDERSCORE-IN-PREFIX", false),
		entry("a.B.c-/a-CAPITAL-LETTER-B-IN-PREFIX", false),
	)
})
