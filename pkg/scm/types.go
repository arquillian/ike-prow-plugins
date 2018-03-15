package scm

// RepositoryIssue holds owner name, repository name and an issue number
type RepositoryIssue struct {
	Owner    string
	RepoName string
	Number   int
}
