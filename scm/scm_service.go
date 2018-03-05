package scm

import (
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"context"
	"fmt"
	"net/http"
	"io/ioutil"
	"github.com/arquillian/ike-prow-plugins/plugin/utils"
)

// RepoScmService is a service for getting information about the given repository
type RepoScmService struct {
	Client *github.Client
	Log    *logrus.Entry
	Repo   *github.Repository
}

// CommitScmService is a service for getting information about the given commit inside of a repository
type CommitScmService struct {
	RepoService   *RepoScmService
	repoOwnerName string
	repoName      string
	SHA           string
}

// CreateScmCommitService creates and instance of CommitScmService with an instance of RepoScmService stored in it
func CreateScmCommitService(client *github.Client, log *logrus.Entry, repo *github.Repository, sha string) *CommitScmService {
	return &CommitScmService{
		RepoService: &RepoScmService{
			Client: client,
			Log:    log,
			Repo:   repo,
		},
		repoOwnerName: *repo.Owner.Login,
		repoName:      *repo.Name,
		SHA:           sha,
	}
}

// AffectedFile is a type that contains information about created/modified/removed file within an scm repository
type AffectedFile struct {
	Name   string
	Status string
}

// GetRepoLanguages returns an array of used programing languages in the related repository
func (repo *RepoScmService) GetRepoLanguages() ([]string, error) {
	return repo.getLanguages(*repo.Repo.LanguagesURL)
}

// getRepoLanguages returns an array of used programing languages retrieved from the given url
func (repo *RepoScmService) getLanguages(langURL string) ([]string, error) {
	req, err := repo.Client.NewRequest("GET", langURL, nil)
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

// GetAffectedFiles returns an array of files affected by the related commit
func (commit *CommitScmService) GetAffectedFiles() ([]AffectedFile, error) {
	repoCommit, _, err := commit.RepoService.Client.Repositories.
		GetCommit(context.Background(), commit.repoOwnerName, commit.repoName, commit.SHA)

	if err != nil {
		return nil, err
	}

	fileNames := make([]AffectedFile, len(repoCommit.Files))
	for _, file := range repoCommit.Files {
		fileNames = append(fileNames, AffectedFile{*file.Filename, *file.Status})
	}
	return fileNames, nil
}

// Fail sets a status "failure" with the given reason to the related commit
func (commit *CommitScmService) Fail(reason string) {
	commit.SetStatus("failure", reason)
}

// Success sets a status "success" with the given reason to the related commit
func (commit *CommitScmService) Success(reason string) {
	commit.SetStatus("success", reason)
}

// SetStatus sets the given status with the given reason to the related commit
func (commit *CommitScmService) SetStatus(status, reason string) {
	if _, _, err := commit.RepoService.Client.Repositories.
		CreateStatus(context.Background(), commit.repoOwnerName, commit.repoName, commit.SHA,
		&github.RepoStatus{
			State:       &status,
			Context:     utils.String("alien-ike/prow-spike"),
			Description: &reason,
		}); err != nil {
		commit.RepoService.Log.Info("Error handling event.", err)
	}
}

// GetRawFile retrieves and returns content of a file at the given path on branch of the related commit
func (commit *CommitScmService) GetRawFile(filePath string) []byte {
	return commit.getRawFile(commit.repoOwnerName, commit.repoName, commit.SHA, filePath)
}

// GetRawFile retrieves and returns content of a file at the given location
func (commit *CommitScmService) getRawFile(owner, repo, sha, path string) []byte {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, sha, path)

	resp, err := http.Get(url)
	if err != nil {
		return nil
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	return body
}
