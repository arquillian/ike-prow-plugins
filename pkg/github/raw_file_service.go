package github

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

// RawFileService encapsulates retrieval of files in the given GitHub repository change
type RawFileService struct {
	Change scm.RepositoryChange
}

// GetRawFileURL creates a url to the given path related to the GitHub repository change
func (s *RawFileService) GetRawFileURL(path string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
		s.Change.Owner, s.Change.RepoName, s.Change.Hash, path)
}
