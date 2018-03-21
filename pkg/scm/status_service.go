package scm

// StatusService encapsulates operation for updating status of the RepositoryChange
type StatusService interface {
	Failure(reason string) error
	Success(reason string) error
	Pending(reason string) error
	Error(reason string) error
}

// RepositoryChange holds information about owner and repository to which the change indicated by Hash belongs
type RepositoryChange struct {
	Owner,
	RepoName,
	Hash string
}
