package command

import (
	"bytes"
	"fmt"
	"strings"
)

var (
	// Admin is a name of the admin role
	Admin = "admin"
	// RequestReviewer is a name of the requested reviewer role
	RequestReviewer = "requested reviewer"
	// PullRequestCreator is a name of the pull request creator role
	PullRequestCreator = "pull request creator"
	// PullRequestApprover is a name of a person who gave an approval to the PR
	PullRequestApprover = "pull request approver"
	// Unknown represents an unknown user
	Unknown = "unknown"
	// Anyone represents any user/role
	Anyone = "anyone"
)

// PermissionStatus keeps information about the user, his permissions and which roles are approved and which rejected
type PermissionStatus struct {
	User           string
	UserIsApproved bool
	ApprovedRoles  []string
	RejectedRoles  []string
}

// NewPermissionStatus creates a new instance of PermissionStatus with the given values
func NewPermissionStatus(user string, userIsApproved bool, approvedRoles []string, rejectedRoles []string) *PermissionStatus {
	return &PermissionStatus{
		User:           user,
		UserIsApproved: userIsApproved,
		ApprovedRoles:  approvedRoles,
		RejectedRoles:  rejectedRoles,
	}
}

func (s *PermissionStatus) reject() *PermissionStatus {
	s.UserIsApproved = false
	return s
}

func (s *PermissionStatus) allow() *PermissionStatus {
	s.UserIsApproved = true
	return s
}

func (s *PermissionStatus) constructMessage(operation, command string) string {
	var msg bytes.Buffer

	msg.WriteString(fmt.Sprintf(
		"Hey @%s! It seems you tried to %s `%s` command, but this will not have any effect due to insufficient permission. "+
			"You have to be ",
		s.User, operation, command))

	if len(s.ApprovedRoles) > 0 {
		msg.WriteString(strings.Join(s.ApprovedRoles, " or "))
		if len(s.RejectedRoles) > 0 {
			msg.WriteString(", but ")
		}
	}

	if len(s.RejectedRoles) > 0 {
		msg.WriteString("not " + strings.Join(s.RejectedRoles, " nor "))
	}

	msg.WriteString(" for this command to take an effect. ")
	return msg.String()
}
