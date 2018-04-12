package scm

import gogh "github.com/google/go-github/github"

// RepositoryIssue holds owner name, repository name and an issue number
type RepositoryIssue struct {
	Owner    string
	RepoName string
	Number   int
}

// ChangedFile is a type that contains information about created/modified/removed file within an scm repository
type ChangedFile struct {
	Name      string
	Status    string
	Additions int
	Deletions int
}

// RepositoryChange holds information about owner and repository to which the change indicated by Hash belongs
type RepositoryChange struct {
	Owner,
	RepoName,
	Hash string
}

// NewChangedFile maps the fields and returns the new struct
func NewChangedFile(file *gogh.CommitFile) *ChangedFile {
	return &ChangedFile{
		Name:      *file.Filename,
		Status:    *file.Status,
		Additions: *file.Additions,
		Deletions: *file.Deletions,
	}
}
