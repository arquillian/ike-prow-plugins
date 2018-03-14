package github

import (
	"context"
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
)

// RepositoryService is a concrete implementation of the interface RepositoryService
type RepositoryService struct {
	Client *github.Client
	Log    *logrus.Entry
	Change scm.RepositoryChange
}

// UsedLanguages returns an array of used programing languages in the related repository
func (repo *RepositoryService) UsedLanguages() ([]string, error) {
	url := fmt.Sprintf("/repos/%s/%s/languages", repo.Change.Owner, repo.Change.RepoName)
	req, err := repo.Client.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	langsStat := make(map[string]interface{})
	_, err = repo.Client.Do(context.Background(), req, &langsStat)
	if err != nil {
		return nil, err
	}

	languages := make([]string, 0, len(langsStat))
	for lang := range langsStat {
		languages = append(languages, lang)
	}

	return languages, nil
}
