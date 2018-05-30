package scm

// StatusService encapsulates operation for updating status of the RepositoryChange
type StatusService interface {
	Failure(reason, detailsPageName string) error
	Success(reason, detailsPageName string) error
	Pending(reason string) error
	Error(reason string) error
}
