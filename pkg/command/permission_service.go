package command

import (
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

func (s *PermissionService) newPermissionStatus(allowedRoles ...string) *PermissionStatus {
	return &PermissionStatus{User: s.User, ApprovedRoles: allowedRoles}
}

// Admin checks if the user is admin
func (s *PermissionService) Admin() (*PermissionStatus, error) {
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
func (s *PermissionService) PRReviewer() (*PermissionStatus, error) {
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
func (s *PermissionService) PRCreator() (*PermissionStatus, error) {
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

