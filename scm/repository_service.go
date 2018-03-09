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

//func ToChangedFiles()

//// GetAffectedFiles returns an array of files affected by the related commit
//func (commit *CommitScmServiceImpl) GetAffectedFiles() ([]ChangedFile, error) {
//	repoCommit, _, err := commit.client.Repositories.
//		GetCommit(context.Background(), commit.repoOwnerName, commit.repoName, commit.sha)
//
//	if err != nil {
//		return nil, err
//	}
//
//	fileNames := make([]ChangedFile, len(repoCommit.Files))
//	for _, file := range repoCommit.Files {
//		fileNames = append(fileNames, ChangedFile{*file.Filename, *file.Status})
//	}
//	return fileNames, nil
//}
