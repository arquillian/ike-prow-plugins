package scm

// StatusService encapsulates operation for updating status of the RepositoryChange
type StatusService interface {
	Failure(reason string) error
	Success(reason string) error
	Pending(reason string) error
	Error(reason string) error
}
