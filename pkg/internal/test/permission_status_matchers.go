package test

import (
	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/onsi/gomega"
)

// ExpectPermissionStatus wraps the given PermissionStatus allowing assertions to be made on it
func ExpectPermissionStatus(status *command.PermissionStatus) *SoftAssertion {
	return NewSoftAssertion(gomega.Expect(*status))
}

// HaveRejectedUser asserts if the PermissionStatus contains a rejected user with the given name
func HaveRejectedUser(expUserName string) SoftMatcher {
	return createMatcher(expUserName, false)
}

// HaveApprovedUser asserts if the PermissionStatus contains an approved user with the given name
func HaveApprovedUser(expUserName string) SoftMatcher {
	return createMatcher(expUserName, true)
}

func createMatcher(user string, shouldBeApproved bool) SoftMatcher {
	return SoftlySatisfyAll(
		TransformWithName(
			func(s command.PermissionStatus) interface{} { return s.User },
			gomega.Equal(user),
			"User"),
		TransformWithName(
			func(s command.PermissionStatus) interface{} { return s.UserIsApproved },
			gomega.Equal(shouldBeApproved),
			"UserIsApproved"))
}

// HaveApprovedRoles asserts if the PermissionStatus contains the given approved roles
func HaveApprovedRoles(roles ...string) SoftMatcher {
	return TransformWithName(
		func(s command.PermissionStatus) interface{} { return s.ApprovedRoles },
		gomega.ConsistOf(roles),
		"ApprovedRoles")
}

// HaveNoApprovedRoles asserts if the PermissionStatus doesn't contain any role that would be approved
func HaveNoApprovedRoles() SoftMatcher {
	return TransformWithName(
		func(s command.PermissionStatus) interface{} { return s.ApprovedRoles },
		gomega.BeEmpty(),
		"ApprovedRoles")
}

// HaveRejectedRoles asserts if the PermissionStatus contains the given rejected roles
func HaveRejectedRoles(roles ...string) SoftMatcher {
	return TransformWithName(
		func(s command.PermissionStatus) interface{} { return s.RejectedRoles },
		gomega.ConsistOf(roles),
		"RejectedRoles")
}

// HaveNoRejectedRoles asserts if the PermissionStatus doesn't contain any role that would be rejected
func HaveNoRejectedRoles() SoftMatcher {
	return TransformWithName(
		func(s command.PermissionStatus) interface{} { return s.RejectedRoles },
		gomega.BeEmpty(),
		"RejectedRoles")
}
