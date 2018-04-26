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
	return fmt.Sprintf(rawURL, s.GetRelativePath(path))
}

// GetRelativePath creates repository specific relative path
func (s *RawFileService) GetRelativePath(path string) string {
	return fmt.Sprintf("%s/%s/%s/%s", s.Change.Owner, s.Change.RepoName, s.Change.Hash, path)
}
