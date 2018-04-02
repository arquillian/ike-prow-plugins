package scm

// StatusService encapsulates operation for updating status of the RepositoryChange
type StatusService interface {
	Failure(reason, targetURL string) error
	Success(reason, targetURL string) error
	Pending(reason string) error
	Error(reason string) error
}
