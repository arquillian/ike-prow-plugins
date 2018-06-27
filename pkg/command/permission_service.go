package command

import (
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
)

// PermissionService keeps user name and PR loader and provides information about the user's permissions
type PermissionService struct {
	client   ghclient.Client
	user     string
	prLoader *ghservice.PullRequestLazyLoader
}

// NewPermissionService creates a new instance of PermissionService with the given client, user and pr loader
func NewPermissionService(client ghclient.Client, user string, prLoader *ghservice.PullRequestLazyLoader) *PermissionService {
	return &PermissionService{
		client:   client,
		user:     user,
		prLoader: prLoader,
	}
}

func (s *PermissionService) newPermissionStatus(allowedRoles ...string) *PermissionStatus {
	return &PermissionStatus{User: s.user, ApprovedRoles: allowedRoles}
}

// Admin checks if the user is admin
func (s *PermissionService) Admin(evaluate bool) (*PermissionStatus, error) {
	status := s.newPermissionStatus(Admin)
	if !evaluate {
		return status, nil
	}
	permissionLevel, err := s.client.GetPermissionLevel(s.prLoader.RepoOwner, s.prLoader.RepoName, s.user)
	if err != nil {
		return status.reject(), err
	}

	if *permissionLevel.Permission == Admin {
		return status.allow(), nil
	}
	return status.reject(), nil
}

// PRReviewer checks if the user is pull request reviewer
func (s *PermissionService) PRReviewer(evaluate bool) (*PermissionStatus, error) {
	status := s.newPermissionStatus(RequestedReviewer)
	if !evaluate {
		return status, nil
	}
	pr, err := s.prLoader.Load()
	if err != nil {
		return status.reject(), err
	}
	for _, reviewer := range pr.RequestedReviewers {
		if s.user == *reviewer.Login {
			return status.allow(), nil
		}
	}
	return status.reject(), nil
}

// PRCreator checks if the user is pull request creator
func (s *PermissionService) PRCreator(evaluate bool) (*PermissionStatus, error) {
	status := s.newPermissionStatus(PullRequestCreator)
	if !evaluate {
		return status, nil
	}
	pr, err := s.prLoader.Load()
	if err != nil {
		return status.reject(), err
	}
	if s.user == *pr.User.Login {
		return status.allow(), nil
	}
	return status.reject(), nil
}

// PRApprover checks if the user approved the pull request
func (s *PermissionService) PRApprover(evaluate bool) (*PermissionStatus, error) {
	status := s.newPermissionStatus(PullRequestApprover)
	if !evaluate {
		return status, nil
	}
	prReviews, err := s.client.GetPullRequestReviews(s.prLoader.RepoOwner, s.prLoader.RepoName, s.prLoader.Number)
	if err != nil {
		return status.reject(), err
	}
	for _, review := range prReviews {
		if *review.State == "APPROVED" && *review.User.Login == s.user {
			return status.allow(), nil
		}
	}
	return status.reject(), nil
}
