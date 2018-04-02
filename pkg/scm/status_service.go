package scm

// StatusService encapsulates operation for updating status of the RepositoryChange
type StatusService interface {
	Failure(reason, detailsLink string) error
	Success(reason, detailsLink string) error
	Pending(reason string) error
	Error(reason string) error
}
