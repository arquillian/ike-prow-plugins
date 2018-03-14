package github

import (
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
)

// RawFileService encapsulates retrieval of files in the given GitHub repository change
type RawFileService struct {
	Change scm.RepositoryChange
}

// GetRawFile retrieves raw file content on the given path from the related GitHub repository change
func (s *RawFileService) GetRawFile(path string) ([]byte, bool, error) {
	return utils.GetFileFromURL(s.GetRawFileURL(path))
}

// GetRawFileURL creates a url to the given path related to the GitHub repository change
func (s *RawFileService) GetRawFileURL(path string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s",
		s.Change.Owner, s.Change.RepoName, s.Change.Hash, path)
}
