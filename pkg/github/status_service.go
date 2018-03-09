package github

import (
	"github.com/sirupsen/logrus"
	"fmt"
	"github.com/google/go-github/github"
	"context"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// StatusService is a struct
type StatusService struct {
	client        *github.Client
	log           *logrus.Entry
	statusContext StatusContext
	change        scm.RepositoryChange
}

// NewStatusService creates an instance of GitHub StatusService
func NewStatusService(client *github.Client, log *logrus.Entry, change scm.RepositoryChange, context StatusContext) scm.StatusService {
	return &StatusService{
		client:        client,
		log:           log,
		statusContext: context,
		change:        change,
	}
}

// Success marks given change as a success.
func (s *StatusService) Success(reason string) error {
	return s.setStatus("success", reason)
}

// Failure marks given change as a failure.
func (s *StatusService) Failure(reason string) error {
	return s.setStatus("failure", reason)
}

// Pending marks given change as a pending.
func (s *StatusService) Pending(reason string) error {
	return s.setStatus("pending", reason)
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
		s.log.Info("Error handling event.", err)
	}

	return err
}
