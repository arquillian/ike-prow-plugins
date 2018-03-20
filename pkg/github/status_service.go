package github

import (
	"context"
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/google/go-github/github"
)

// StatusService is a struct
type StatusService struct {
	client        *github.Client
	log           log.Logger
	statusContext StatusContext
	change        scm.RepositoryChange
}

// NewStatusService creates an instance of GitHub StatusService
func NewStatusService(client *github.Client, log log.Logger, change scm.RepositoryChange, context StatusContext) scm.StatusService {
	return &StatusService{
		client:        client,
		log:           log,
		statusContext: context,
		change:        change,
	}
}

// Success marks given change as a success.
func (s *StatusService) Success(reason string) error {
	return s.setStatus(StatusSuccess, reason)
}

// Failure marks given change as a failure.
func (s *StatusService) Failure(reason string) error {
	return s.setStatus(StatusFailure, reason)
}

// Pending marks given change as a pending.
func (s *StatusService) Pending(reason string) error {
	return s.setStatus(StatusPending, reason)
}

// Error marks given change as a error.
func (s *StatusService) Error(reason string) error {
	return s.setStatus(StatusError, reason)
}


// setStatus sets the given status with the given reason to the related commit
func (s *StatusService) setStatus(status, reason string) error {
	c := fmt.Sprintf("%q/%q", s.statusContext.BotName, s.statusContext.PluginName)
	repoStatus := github.RepoStatus{
		State:       &status,
		Context:     &c,
		Description: &reason,
	}
	_, _, err := s.client.Repositories.CreateStatus(context.Background(), s.change.Owner, s.change.RepoName, s.change.Hash, &repoStatus)

	if err != nil {
		s.log.Errorf("error trying to send status. %q. cause: %q", repoStatus, err)
	}

	return err
}
