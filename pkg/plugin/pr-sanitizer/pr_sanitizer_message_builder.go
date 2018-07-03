package prsanitizer

import (
	"strings"
	"fmt"
)

type failureHintMessageBuilder struct {
	hint []string
}

// HintMessage is message containing failure reasons of pr-sanitizer and why is it important.
type HintMessage string

// FailureHintMessageBuilder is builder to build failure message for description.
type FailureHintMessageBuilder interface {
	Title(isValid bool) FailureHintMessageBuilder
	Description(desc string, contentLength int) FailureHintMessageBuilder
	IssueLink(isIssueLinked bool) FailureHintMessageBuilder
	Build() HintMessage
}

const (
	// TitleFailureMessage is a message used in GH Status as description when the PR title does not follow semantic message style
	TitleFailureMessage = "PR title does not conform with semantic commit message style. Conformance with the semantic commit message style makes your changelog and git history clean."

	// DescriptionLengthShortMessage is message notification for contributor about short PR description content.
	DescriptionLengthShortMessage = "PR description is too short, expecting more than %d characters. More elaborated description will be helpful to understand changes proposed in this PR."

	// IssueLinkMissingMessage is message notification for contributor about missing issue link.
	IssueLinkMissingMessage = "Issue link is missing in this PR description. Issue link with keywords in the PR description is helpful to close issues automatically after merging PR."
)

func (mb *failureHintMessageBuilder) Title(isValid bool) FailureHintMessageBuilder {
	if !isValid {
		mb.hint = append(mb.hint, TitleFailureMessage)
	}
	return mb
}

func (mb *failureHintMessageBuilder) Description(desc string, descriptionContentLength int) FailureHintMessageBuilder {
	contentLength := 50
	if descriptionContentLength != 0 {
		contentLength = descriptionContentLength
	}
	if len(desc) <= contentLength {
		mb.hint = append(mb.hint, fmt.Sprintf(DescriptionLengthShortMessage, contentLength))
	}
	return mb
}

func (mb *failureHintMessageBuilder) IssueLink(isIssueLinked bool) FailureHintMessageBuilder {
	if !isIssueLinked {
		mb.hint = append(mb.hint, IssueLinkMissingMessage)
	}
	return mb
}

func (mb *failureHintMessageBuilder) Build() (HintMessage) {
	return HintMessage(strings.Join(mb.hint, "\n\n"))
}

// NewFailureMessageBuilder creates failureHintMessageBuilder with empty message.
func NewFailureMessageBuilder() FailureHintMessageBuilder {
	return &failureHintMessageBuilder{}
}
