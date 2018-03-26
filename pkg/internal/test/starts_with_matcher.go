package test

import (
	"fmt"
	"strings"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

// StartWith is a GomegaMatcher verifying if given string starts with prefix
func StartWith(expected string) types.GomegaMatcher {
	return &startsWith{expected}
}

type startsWith struct {
	prefix string
}

func (matcher *startsWith) Match(actual interface{}) (success bool, err error) {
	actualString, ok := toString(actual)
	if !ok {
		return false, fmt.Errorf("StartsWith matcher requires a string or stringer.  Got:\n%s", format.Object(actual, 1))
	}

	return strings.HasPrefix(actualString, matcher.prefix), nil
}

func (matcher *startsWith) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to start with", matcher.prefix)
}

func (matcher *startsWith) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to start with", matcher.prefix)
}

func toString(a interface{}) (string, bool) {
	aString, isString := a.(string)
	if isString {
		return aString, true
	}

	aBytes, isBytes := a.([]byte)
	if isBytes {
		return string(aBytes), true
	}

	aStringer, isStringer := a.(fmt.Stringer)
	if isStringer {
		return aStringer.String(), true
	}

	return "", false
}
