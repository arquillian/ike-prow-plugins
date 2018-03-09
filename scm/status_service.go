package scm

type StatusService interface {
	Failure(reason string) error
	Success(reason string) error
}

type RepositoryChange struct {
	Owner,
	RepoName,
	Hash string
}
