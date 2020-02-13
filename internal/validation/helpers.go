package validation

import (
	"net/url"
	"regexp"

	gatewayv1alpha1 "github.com/kyma-incubator/api-gateway/api/v1alpha1"
)

func hasDuplicates(rules []gatewayv1alpha1.Rule) bool {
	duplicates := false

	if len(rules) > 1 {
		for i, ruleToCheck := range rules {
			for j, rule := range rules {
				if i != j && ruleToCheck.Path == rule.Path {
					if len(ruleToCheck.Methods) > 0 || len(rule.Methods) > 0 {
						for _, methodToCheck := range ruleToCheck.Methods {
							for _, method := range rule.Methods {
								if methodToCheck == method {
									duplicates = true
								}
							}
						}
					} else {
						duplicates = true
					}
				}
			}
		}
	}

	return duplicates
}

func isValidURL(toTest string) bool {
	if len(toTest) == 0 {
		return false
	}
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	}
	return true
}

//ValidateDomainName ?
func ValidateDomainName(domain string) bool {
	RegExp := regexp.MustCompile(`^([a-zA-Z0-9][a-zA-Z0-9-_]*\.)*[a-zA-Z0-9]*[a-zA-Z0-9-_]*[[a-zA-Z0-9]+$`)
	return RegExp.MatchString(domain)
}

func ValidateServiceName(service string) bool {
	regExp := regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?\.[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	return regExp.MatchString(service)
}
