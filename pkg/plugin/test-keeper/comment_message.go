package testkeeper

import (
	"github.com/arquillian/ike-prow-plugins/pkg/comment"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
)

const (
	paragraph = "\n\n"

	beginning = "It appears that no tests have been added or updated in this PR." +
		paragraph +
		"Automated tests give us confidence in shipping reliable software. Please add some as part of this change." +
		paragraph +
		"If you are an admin or the reviewer of this PR and you are sure that no test is needed then you can use the command `" + BypassCheckComment + "` " +
		"as a comment to make the status green.\n"

	documentationSection = "#_test_keeper_plugin"

	sadIke = `<img align="left" src="https://cdn.rawgit.com/bartoszmajsak/ike-prow-plugins/2025328b70bd1879520411b3cacadee61a49641a/docs/images/arquillian_ui_failure_128px.png">`
)

// CreateCommentMessage creates a comment message for the test-keeper plugin. If the comment message is set in config then it takes that one, the default otherwise.
func CreateCommentMessage(configuration PluginConfiguration, change scm.RepositoryChange) string {
	message := comment.Message{
		Thumbnail:     sadIke,
		Description:   beginning,
		ConfigFile:    configuration.LocationURL,
		Documentation: documentationSection,
	}
	return comment.LoadMessageTemplate(message, configuration.PluginConfiguration, change)
}
