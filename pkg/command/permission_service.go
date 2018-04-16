package command

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
)

// PermissionService keeps user name and PR loader and provides information about the user's permissions
type PermissionService struct {
	Client   ghclient.Client
	User     string
	PRLoader *ghservice.PullRequestLazyLoader
}

// NewPermissionService creates a new instance of PermissionService with the given client, user and pr loader
func NewPermissionService(client ghclient.Client, user string, prLoader *ghservice.PullRequestLazyLoader) *PermissionService {
	return &PermissionService{
		Client:   client,
		User:     user,
		PRLoader: prLoader,
	}
}

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

func (s PermissionService) newPermissionStatus(allowedRoles ...string) *PermissionStatus {
	return &PermissionStatus{User: s.User, ApprovedRoles: allowedRoles}
}

func (s PermissionStatus) reject() *PermissionStatus {
	s.UserIsApproved = false
	return &s
}

func (s PermissionStatus) allow() *PermissionStatus {
	s.UserIsApproved = true
	return &s
}

func (s PermissionStatus) constructMessage(operation, command string) string {
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

// PermissionCheck represents any check of the user's permissions and returns PermissionStatus that contains the result
type PermissionCheck func() (*PermissionStatus, error)

var (
	// Admin is a name of the admin role
	Admin = "admin"
	// RequestReviewer is a name of the requested reviewer role
	RequestReviewer = "requested reviewer"
	// PullRequestCreator is a name of the pull request creator role
	PullRequestCreator = "pull request creator"
	// Unknown represents an unknown user
	Unknown = "unknown"
	// Anyone represents any user/role
	Anyone = "anyone"

	// Anybody allows to any user
	Anybody = func() (*PermissionStatus, error) {
		return &PermissionStatus{UserIsApproved: true, ApprovedRoles: []string{Anyone}}, nil
	}

	// AnyOf checks if any of the given permission checks is fulfilled
	AnyOf = func(permissionChecks ...PermissionCheck) PermissionCheck {
		return of(true, permissionChecks...)
	}

	// AllOf checks if all of the given permission checks are fulfilled
	AllOf = func(permissionChecks ...PermissionCheck) PermissionCheck {
		return of(false, permissionChecks...)
	}
)

// Admin checks if the user is admin
func (s PermissionService) Admin() (*PermissionStatus, error) {
	status := s.newPermissionStatus(Admin)
	permissionLevel, err := s.Client.GetPermissionLevel(s.PRLoader.RepoOwner, s.PRLoader.RepoName, s.User)
	if err != nil {
		return status.reject(), err
	}

	if *permissionLevel.Permission == Admin {
		return status.allow(), nil
	}
	return status.reject(), nil
}

// PRReviewer checks if the user is pull request reviewer
func (s PermissionService) PRReviewer() (*PermissionStatus, error) {
	status := s.newPermissionStatus(RequestReviewer)
	pr, err := s.PRLoader.Load()
	if err != nil {
		return status.reject(), err
	}
	for _, reviewer := range pr.RequestedReviewers {
		if s.User == *reviewer.Login {
			return status.allow(), nil
		}
	}
	return status.reject(), nil
}

// PRCreator checks if the user is pull request creator
func (s PermissionService) PRCreator() (*PermissionStatus, error) {
	status := s.newPermissionStatus(PullRequestCreator)
	pr, err := s.PRLoader.Load()
	if err != nil {
		return status.reject(), err
	}
	if s.User == *pr.User.Login {
		return status.allow(), nil
	}
	return status.reject(), nil
}

// Not reverses the given permission check
var Not = func(restriction PermissionCheck) PermissionCheck {
	return func() (*PermissionStatus, error) {
		status, err := restriction()
		reversedStatus := &PermissionStatus{User: status.User}
		reversedStatus.UserIsApproved = !status.UserIsApproved
		reversedStatus.RejectedRoles = status.ApprovedRoles
		reversedStatus.ApprovedRoles = status.RejectedRoles
		return reversedStatus, err
	}
}

func of(any bool, permissionChecks ...PermissionCheck) PermissionCheck {
	return func() (*PermissionStatus, error) {
		statuses := make([]*PermissionStatus, 0, len(permissionChecks))
		for _, checkPermission := range permissionChecks {

			status, err := checkPermission()
			statuses = append(statuses, status)
			if err != nil {
				return Flatten(statuses, any), err
			}
		}
		return Flatten(statuses, any), nil
	}
}

// Flatten takes a slice of permission statuses and returns one with the flattened values.
// anyOff parameter sets if the user should be approved in any of the given permission statuses or in all of them
func Flatten(statuses []*PermissionStatus, anyOff bool) *PermissionStatus {
	if len(statuses) == 0 {
		return &PermissionStatus{User: Unknown, UserIsApproved: true, ApprovedRoles: []string{Anyone}}
	}
	flattenedStatus := statuses[0]
	if len(statuses) == 1 {
		return flattenedStatus
	}
	for i := 1; i < len(statuses); i++ {
		status := statuses[i]
		flattenedStatus.User = status.User
		if anyOff {
			flattenedStatus.UserIsApproved = flattenedStatus.UserIsApproved || status.UserIsApproved
		} else {
			flattenedStatus.UserIsApproved = flattenedStatus.UserIsApproved && status.UserIsApproved
		}
		flattenedStatus.ApprovedRoles = append(flattenedStatus.ApprovedRoles, status.ApprovedRoles...)
		flattenedStatus.RejectedRoles = append(flattenedStatus.RejectedRoles, status.RejectedRoles...)
	}
	return flattenedStatus
}
