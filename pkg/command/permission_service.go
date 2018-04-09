package command

import (
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github"
)

// PermissionService keeps user name and PR loader and provides information about the user's permissions
type PermissionService struct {
	Client   *github.Client
	User     string
	PRLoader *github.PullRequestLazyLoader
}

// NewPermissionService creates a new instance of PermissionService with the given client, user and pr loader
func NewPermissionService(client *github.Client, user string, prLoader *github.PullRequestLazyLoader) *PermissionService {
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

func (s *PermissionService) newPermissionStatus(allowedRoles ...string) *PermissionStatus {
	return &PermissionStatus{User: s.User, ApprovedRoles: allowedRoles}
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
	msg := fmt.Sprintf(
		"@%s has %s a command `%s` but this user is not allowed to do that. "+
			"Users with the necessary permissions are anybody who is %s, but not %s",
		s.User, operation, command, strings.Join(s.ApprovedRoles, " or "), strings.Join(s.RejectedRoles, " nor "))
	return msg
}

// PermissionCheck represents any check of the user's permissions and returns PermissionStatus that contains the result
type PermissionCheck func() (*PermissionStatus, error)

// Anybody allows to any user
var Anybody PermissionCheck = func() (*PermissionStatus, error) {
	return &PermissionStatus{UserIsApproved: true, ApprovedRoles: []string{"anyone"}}, nil
}

// Admin checks if the user is admin
func (s *PermissionService) Admin() (*PermissionStatus, error) {
	status := s.newPermissionStatus("admin")
	permissionLevel, err := s.Client.GetPermissionLevel(s.PRLoader.RepoOwner, s.PRLoader.RepoName, s.User)
	if err != nil {
		return status.reject(), err
	}

	if *permissionLevel.Permission == "admin" {
		return status.allow(), nil
	}
	return status.reject(), nil
}

// PRReviewer checks if the user is pull request reviewer
func (s *PermissionService) PRReviewer() (*PermissionStatus, error) {
	status := s.newPermissionStatus("requested reviewer")
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
func (s *PermissionService) PRCreator() (*PermissionStatus, error) {
	status := s.newPermissionStatus("pull request creator")
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

// AnyOf checks if any of the given permission checks is fulfilled
var AnyOf = func(permissionChecks ...PermissionCheck) PermissionCheck {
	return of(true, permissionChecks...)
}

// AllOf checks if all of the given permission checks are fulfilled
var AllOf = func(permissionChecks ...PermissionCheck) PermissionCheck {
	return of(false, permissionChecks...)
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
		return &PermissionStatus{User: "unknown", UserIsApproved: true, ApprovedRoles: []string{"anyone"}}
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
