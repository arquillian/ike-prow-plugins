package scm

// RepositoryService is a service for getting information about the given repository
type RepositoryService interface {
	UsedLanguages() ([]string, error)
}

// ChangedFile is a type that contains information about created/modified/removed file within an scm repository
type ChangedFile struct {
	Name   string
	Status string
}


