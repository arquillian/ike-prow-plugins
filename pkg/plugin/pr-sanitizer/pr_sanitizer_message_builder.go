package prsanitizer

import "strings"

type failureMessageBuilder struct {
	description []string
}

// FailureMessage is message containing failure reasons of pr-sanitizer.
type FailureMessage string

// FailureMessageBuilder is builder to build failure message for description.
type FailureMessageBuilder interface {
	Title(isValid bool) FailureMessageBuilder
	Description(desc string) FailureMessageBuilder
	IssueLink(isIssueLinked bool) FailureMessageBuilder
	Build() FailureMessage
}

const (
	// TitleFailure is a message used in GH Status as description when the PR title does not follow semantic message style
	TitleFailure = "PR title does not conform with semantic commit message style."

	// IssueLinkMissing is a message used in GH Status as description when no tests shipped with the PR
	IssueLinkMissing = "Issue link is missing in this PR description."

	// DescriptionLengthShort is a message used in GH Status as description when the PR description does not have minimum required characters.
	DescriptionLengthShort = "PR description is too short, expecting more than 50 characters."
)

func (mb *failureMessageBuilder) Title(isValid bool) FailureMessageBuilder {
	if !isValid {
		mb.description = append(mb.description, TitleFailure)
	}
	return mb
}

func (mb *failureMessageBuilder) Description(desc string) FailureMessageBuilder {
	if len(desc) <= 50 {
		mb.description = append(mb.description, DescriptionLengthShort)
	}
	return mb
}

func (mb *failureMessageBuilder) IssueLink(isIssueLinked bool) FailureMessageBuilder {
	if !isIssueLinked {
		mb.description = append(mb.description, IssueLinkMissing)
	}
	return mb
}

func (mb *failureMessageBuilder) Build() FailureMessage {
	return FailureMessage(strings.Join(mb.description, " "))
}

// NewFailureMessageBuilder creates failureMessageBuilder with empty message.
func NewFailureMessageBuilder() FailureMessageBuilder {
	return &failureMessageBuilder{}
}
