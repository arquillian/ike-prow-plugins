package test

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

// SoftMatcher is a types.GomegaMatcher that can be used in SoftlySatisfyAll(...) matching
type SoftMatcher interface {
	types.GomegaMatcher
	createMsgAndIncrement(sequenceNumber *int, actual interface{}, isNegated bool) string
}

// SoftAssertion wraps GomegaAssertion instance
type SoftAssertion struct {
	assertion gomega.GomegaAssertion
}

// NewSoftAssertion creates a new instance of SoftAssertion with the given GomegaAssertion
func NewSoftAssertion(assertion gomega.GomegaAssertion) *SoftAssertion {
	return &SoftAssertion{assertion: assertion}
}

// To expects that all of the given matchers should pass
func (m *SoftAssertion) To(matchers ...SoftMatcher) bool {
	return m.assertion.To(SoftlySatisfyAll(matchers...))
}

// NotTo expects that none of the given matchers should pass
func (m *SoftAssertion) NotTo(matchers ...SoftMatcher) bool {
	return m.assertion.NotTo(SoftlySatisfyAll(matchers...))
}

// NamedGomegaMatcher keeps a GomegaMatcher instance along with an element name the matchers matches
type NamedGomegaMatcher struct {
	types.GomegaMatcher
	ElementName string
}

func (m *NamedGomegaMatcher) createMsgAndIncrement(sequenceNumber *int, actual interface{}, isNegated bool) string {
	if isNegated {
		return createMsgAndIncrement(sequenceNumber, m.ElementName, m.NegatedFailureMessage(actual))
	}
	return createMsgAndIncrement(sequenceNumber, m.ElementName, m.FailureMessage(actual))
}

func createMsgAndIncrement(sequenceNumber *int, elementName, failureMessage string) string {
	msg := fmt.Sprintf("\n%d) [%s]\n%s\n", *sequenceNumber, elementName, failureMessage)
	*sequenceNumber++
	return msg
}

// TransformWithName creates a new instance of NamedGomegaMatcher with GomegaMatcher retrieved by transforming of the given
// transform and matcher values
func TransformWithName(transform interface{}, matcher types.GomegaMatcher, elementName string) SoftMatcher {
	return &NamedGomegaMatcher{
		GomegaMatcher: gomega.WithTransform(transform, matcher),
		ElementName:   elementName,
	}
}

// SoftlySatisfyAll succeeds only if all of the given matchers succeed.
// The matchers are tried in order, and all errors are collected instead of stopping at the first one.
func SoftlySatisfyAll(matchers ...SoftMatcher) SoftMatcher {
	return &SoftlyAllMatcher{Matchers: matchers}
}

// SoftlyAllMatcher as a GomegaMatcher implementation for soft assertions
type SoftlyAllMatcher struct {
	Matchers       []SoftMatcher
	failedMatchers []SoftMatcher
	passedMatchers []SoftMatcher
}

// Match matches all available matchers and divides them into two arrays based on the information if the match was successful or not
func (m *SoftlyAllMatcher) Match(actual interface{}) (success bool, err error) {
	m.failedMatchers = make([]SoftMatcher, 0, len(m.Matchers))
	m.passedMatchers = make([]SoftMatcher, 0, len(m.Matchers))

	for _, matcher := range m.Matchers {
		success, err := matcher.Match(actual)
		if !success {
			m.failedMatchers = append(m.failedMatchers, matcher)
		} else {
			m.passedMatchers = append(m.passedMatchers, matcher)
		}
		if err != nil {
			return false, err
		}
	}

	return len(m.failedMatchers) == 0, nil
}

// FailureMessage creates a failure message
func (m *SoftlyAllMatcher) FailureMessage(actual interface{}) (message string) {
	return m.createMsg(actual, false)
}

func (m *SoftlyAllMatcher) createMsg(actual interface{}, isNegated bool) string {
	sequenceNumber := utils.Int(1)
	return fmt.Sprintf("For the given input:\n\n%s\n\nthe following matchers failed:\n%s",
		format.Object(actual, 1), m.createMsgAndIncrement(sequenceNumber, actual, isNegated))
}

func (m *SoftlyAllMatcher) createMsgAndIncrement(sequenceNumber *int, actual interface{}, isNegated bool) string {
	var msg string

	matchers := m.failedMatchers
	if isNegated {
		matchers = m.passedMatchers
	}

	for _, matcher := range matchers {
		msg += matcher.createMsgAndIncrement(sequenceNumber, actual, isNegated)
	}

	return msg
}

// NegatedFailureMessage creates a negated failure message
func (m *SoftlyAllMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return m.createMsg(actual, true)
}
