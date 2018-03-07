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
type RepoScmService interface {
	GetRepoLanguages() ([]string, error)
}

// RepoScmServiceImpl is a concrete implementation of the interface RepoScmService
type RepoScmServiceImpl struct {
	client *github.Client
	log    *logrus.Entry
	repo   *github.Repository
}

// CommitScmService is a type that gets information about the given commit inside of a repository
type CommitScmService interface {
	GetRepoService() RepoScmService
	GetAffectedFiles() ([]AffectedFile, error)
	Fail(reason string)
	Success(reason string)
	SetStatus(status, reason string)
	GetRawFile(filePath string) []byte
}

// CommitScmServiceImpl is a concrete implementation of the interface CommitScmService
type CommitScmServiceImpl struct {
	client        *github.Client
	log           *logrus.Entry
	RepoService   RepoScmService
	repoOwnerName string
	repoName      string
	sha           string
}

// CreateScmCommitService creates and instance of CommitScmService with an instance of RepoScmService stored in it
func CreateScmCommitService(client *github.Client, log *logrus.Entry, repo *github.Repository, sha string) CommitScmService {
	return &CommitScmServiceImpl{
		client: client,
		log:    log,
		RepoService: &RepoScmServiceImpl{
			client: client,
			log:    log,
			repo:   repo,
		},
		repoOwnerName: *repo.Owner.Login,
		repoName:      *repo.Name,
		sha:           sha,
	}
}

// AffectedFile is a type that contains information about created/modified/removed file within an scm repository
type AffectedFile struct {
	Name   string
	Status string
}

// GetRepoLanguages returns an array of used programing languages in the related repository
func (repo *RepoScmServiceImpl) GetRepoLanguages() ([]string, error) {
	return repo.getLanguages(*repo.repo.LanguagesURL)
}

// getRepoLanguages returns an array of used programing languages retrieved from the given url
func (repo *RepoScmServiceImpl) getLanguages(langURL string) ([]string, error) {
	req, err := repo.client.NewRequest("GET", langURL, nil)
	if err != nil {
		return nil, err
	}

	langsStat := make(map[string]interface{})
	_, err = repo.client.Do(context.Background(), req, &langsStat)
	if err != nil {
		return nil, err
	}

	languages := make([]string, 0, len(langsStat))
	for lang := range langsStat {
		languages = append(languages, lang)
	}

	return languages, nil
}

// GetRepoService returns scm service of the repository this commit belongs to
func (commit *CommitScmServiceImpl) GetRepoService() RepoScmService {
	return commit.RepoService
}

// GetAffectedFiles returns an array of files affected by the related commit
func (commit *CommitScmServiceImpl) GetAffectedFiles() ([]AffectedFile, error) {
	repoCommit, _, err := commit.client.Repositories.
		GetCommit(context.Background(), commit.repoOwnerName, commit.repoName, commit.sha)

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
func (commit *CommitScmServiceImpl) Fail(reason string) {
	commit.SetStatus("failure", reason)
}

// Success sets a status "success" with the given reason to the related commit
func (commit *CommitScmServiceImpl) Success(reason string) {
	commit.SetStatus("success", reason)
}

// SetStatus sets the given status with the given reason to the related commit
func (commit *CommitScmServiceImpl) SetStatus(status, reason string) {
	if _, _, err := commit.client.Repositories.
		CreateStatus(context.Background(), commit.repoOwnerName, commit.repoName, commit.sha,
		&github.RepoStatus{
			State:       &status,
			Context:     utils.String("alien-ike/prow-spike"),
			Description: &reason,
		}); err != nil {
		commit.log.Info("Error handling event.", err)
	}
}

// GetRawFile retrieves and returns content of a file at the given path on branch of the related commit
func (commit *CommitScmServiceImpl) GetRawFile(filePath string) []byte {
	return commit.getRawFile(commit.repoOwnerName, commit.repoName, commit.sha, filePath)
}

// GetRawFile retrieves and returns content of a file at the given location
func (commit *CommitScmServiceImpl) getRawFile(owner, repo, sha, path string) []byte {
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
