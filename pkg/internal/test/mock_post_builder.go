package test

import (
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"github.com/arquillian/ike-prow-plugins/pkg/status/message"
	"gopkg.in/h2non/gock.v1"
)

var (
	// ToBe as a wrapper for a status matcher
	ToBe = func(status, description, detailsLink string) BuilderMatcher {
		return func(builder *MockPrBuilder) SoftMatcher {
			docStatusRoot := fmt.Sprintf("%s/status/%s", plugin.DocumentationURL, builder.pluginName)

			return SoftlySatisfyAll(
				HaveState(status),
				HaveDescription(description),
				HaveContext(strings.Join([]string{"alien-ike", builder.pluginName}, "/")),
				HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, strings.ToLower(status), detailsLink)),
			)
		}
	}
)

// BuilderMatcher creates a SoftMatcher for the given builder
type BuilderMatcher func(builder *MockPrBuilder) SoftMatcher

// ContainingStatusMessage verifies that the comment should contain status message with the given description
func ContainingStatusMessage(description string) BuilderMatcher {
	return func(builder *MockPrBuilder) SoftMatcher {
		return SoftlySatisfyAll(
			HaveBodyThatContains(fmt.Sprintf(message.PluginTitleTemplate, builder.pluginName)),
			HaveBodyThatContains("@"+*builder.pullRequest.User.Login),
			HaveBodyThatContains(description))
	}
}

// To softly satisfies all the given matchers and creates a BuilderMatcher from it
func To(matchers ...SoftMatcher) BuilderMatcher {
	return func(builder *MockPrBuilder) SoftMatcher {
		return SoftlySatisfyAll(matchers...)
	}
}

// Expecting creates mocks for the given matchers
func (b *MockPrBuilder) Expecting(mockCreators ...MockCreator) *MockPrBuilder {
	b.mockCreators = append(b.mockCreators, mockCreators...)
	return b
}

// Comment creates a gock matcher to check that there is a Post with a comment that complies with the given restrictions
func Comment(matherForPlugin BuilderMatcher) MockCreator {
	return func(builder *MockPrBuilder) {
		basePostCommentMock(builder)(matherForPlugin(builder))
	}
}

// ChangedComment creates a gock matcher to check that there is a Patch request for the given comment id and containing a comment that complies with the given restrictions
func ChangedComment(commendID int, matherForPlugin BuilderMatcher) MockCreator {
	return func(builder *MockPrBuilder) {
		path := fmt.Sprintf("%s/issues/comments/%d", builder.baseRepoPath(), commendID)
		basePatchMock(path)(matherForPlugin(builder))
	}
}

// NoComment creates a gock matcher to check that there is no Post comment request sent
func NoComment() MockCreator {
	return func(builder *MockPrBuilder) {
		basePostCommentMock(builder)(nil)
	}
}
func basePostCommentMock(builder *MockPrBuilder) func(mather SoftMatcher) {
	path := fmt.Sprintf("%s/issues/%d/comments", builder.baseRepoPath(), *builder.pullRequest.Number)
	return basePostMock(path)
}

func basePostMock(path string) func(mather SoftMatcher) {
	return func(mather SoftMatcher) {
		post := baseGockMock(func(request *gock.Request) { request.Post(path) })
		if mather != nil {
			post.SetMatcher(ExpectPayload(mather)).
				Reply(201)
		} else {
			post.Times(0)
		}
	}
}

func baseDeleteMock(path, responseBody string) {
	baseGockMock(func(request *gock.Request) { request.Delete(path) }).
		Reply(200).
		BodyString(responseBody)
}

func basePatchMock(path string) func(mather SoftMatcher) {
	return func(mather SoftMatcher) {
		baseGockMock(func(request *gock.Request) { request.Patch(path) }).
			SetMatcher(ExpectPayload(mather)).
			Reply(200)
	}
}

// Status creates a gock matcher to check that there is a Post with a status that complies with the given restrictions
func Status(matherForPlugin BuilderMatcher) MockCreator {
	return func(builder *MockPrBuilder) {
		basePostStatusMock(builder)(matherForPlugin(builder))
	}
}

// NoStatus creates a gock matcher to check that there is no Post status request sent
func NoStatus() MockCreator {
	return func(builder *MockPrBuilder) {
		basePostStatusMock(builder)(nil)
	}
}

func basePostStatusMock(builder *MockPrBuilder) func(mather SoftMatcher) {
	return basePostMock(fmt.Sprintf("%s/statuses", builder.baseRepoPath()))
}

// RemovedLabel creates a gock matcher to check that there is a Delete request for the given label sent
func RemovedLabel(labelName string, response string) MockCreator {
	return func(builder *MockPrBuilder) {
		path := fmt.Sprintf("%s/issues/%d/labels/%s", builder.baseRepoPath(), *builder.pullRequest.Number, labelName)
		baseDeleteMock(path, response)
	}
}

// AddedLabel creates a gock matcher to check that there is a Post request for the given label sent
func AddedLabel(labelContent string) MockCreator {
	return func(builder *MockPrBuilder) {
		path := fmt.Sprintf("%s/issues/%d/labels", builder.baseRepoPath(), *builder.pullRequest.Number)
		basePostMock(path)(SoftlySatisfyAll(HaveBodyThatContains(labelContent)))
	}
}

// ChangedTitle creates a gock matcher to check that there is a Patch request containing a new changed title
func ChangedTitle(newTitleContent string) MockCreator {
	return func(builder *MockPrBuilder) {
		path := fmt.Sprintf("%s/pulls/%d", builder.baseRepoPath(), *builder.pullRequest.Number)
		basePatchMock(path)(SoftlySatisfyAll(HaveTitle(newTitleContent)))
	}
}
