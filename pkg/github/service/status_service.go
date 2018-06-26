package ghservice

import (
	"fmt"

	"strings"

	github_type "github.com/arquillian/ike-prow-plugins/pkg/github"
	"github.com/arquillian/ike-prow-plugins/pkg/github/client"
	"github.com/arquillian/ike-prow-plugins/pkg/log"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	"github.com/google/go-github/github"
)

// StatusService is a struct
type StatusService struct {
	client        ghclient.Client
	log           log.Logger
	statusContext github_type.StatusContext
	change        scm.RepositoryChange
}

// NewStatusService creates an instance of GitHub StatusService
func NewStatusService(client ghclient.Client, log log.Logger, change scm.RepositoryChange, context github_type.StatusContext) scm.StatusService {
	return &StatusService{
		client:        client,
		log:           log,
		statusContext: context,
		change:        change,
	}
}

// Success marks given change as a success.
func (s *StatusService) Success(reason, detailsPageName string) error {
	return s.setStatus(github_type.StatusSuccess, reason, s.generateDetailsLink(detailsPageName, github_type.StatusSuccess))
}

// Failure marks given change as a failure.
func (s *StatusService) Failure(reason, detailsPageName string) error {
	return s.setStatus(github_type.StatusFailure, reason, s.generateDetailsLink(detailsPageName, github_type.StatusFailure))
}

// Pending marks given change as a pending.
func (s *StatusService) Pending(reason string) error {
	return s.setStatus(github_type.StatusPending, reason, "")
}

// Error marks given change as a error.
func (s *StatusService) Error(reason string) error {
	return s.setStatus(github_type.StatusError, reason, "")
}

// setStatus sets the given status with the given reason to the related commit
func (s *StatusService) setStatus(status, reason, detailsLink string) error {
	c := fmt.Sprintf("%s/%s", s.statusContext.BotName, s.statusContext.PluginName)
	repoStatus := github.RepoStatus{
		State:       &status,
		Context:     &c,
		Description: &reason,
		TargetURL:   utils.String(detailsLink),
	}

	err := s.client.CreateStatus(s.change, &repoStatus)

	if err != nil {
		s.log.Errorf("error trying to send status. %q. cause: %q", repoStatus, err)
	}

	return err
}

func (s StatusService) generateDetailsLink(filename, status string) string {
	return fmt.Sprintf("%s/status/%s/%s/%s.html", plugin.DocumentationURL, s.statusContext.PluginName, strings.ToLower(status), filename)
}
