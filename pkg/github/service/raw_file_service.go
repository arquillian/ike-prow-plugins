package ghservice

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

const rawURL = "https://raw.githubusercontent.com/%s"

// RawFileService encapsulates retrieval of files in the given GitHub repository change
type RawFileService struct {
	Change scm.RepositoryChange
}

// GetRawFileURL creates a url to the given path related to the GitHub repository change
func (s *RawFileService) GetRawFileURL(path string) string {
	return fmt.Sprintf(rawURL, s.GetRelativePath(path, false))
}

// GetRelativePath creates repository specific relative path
func (s *RawFileService) GetRelativePath(path string, useBlob bool) string {
	repoPath := s.Change.Owner +"/"+ s.Change.RepoName
	if useBlob {
		repoPath += "/blob"
	}
	return fmt.Sprintf("%s/%s/%s", repoPath, s.Change.Hash, path)
}
