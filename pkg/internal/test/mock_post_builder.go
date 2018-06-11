package test

import (
	"fmt"
	"strings"

	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/plugin"
	"gopkg.in/h2non/gock.v1"
)

var (
	// ToBe as a wrapper for a status matcher
	ToBe = func(status, description, detailsLink string) func(pluginName string) SoftMatcher {
		return func(pluginName string) SoftMatcher {
			docStatusRoot := fmt.Sprintf("%s/status/%s", plugin.DocumentationURL, pluginName)

			return SoftlySatisfyAll(
				HaveState(status),
				HaveDescription(description),
				HaveContext(strings.Join([]string{"alien-ike", pluginName}, "/")),
				HaveTargetURL(fmt.Sprintf("%s/%s/%s.html", docStatusRoot, strings.ToLower(status), detailsLink)),
			)
		}
	}
	// ToHaveBodyWithWholePluginsComment verifies that the comment should contain the fixed part of plugins hint comment
	ToHaveBodyWithWholePluginsComment = func(pluginName string) SoftMatcher {
		return SoftlySatisfyAll(
			HaveBodyThatContains(fmt.Sprintf(ghservice.PluginTitleTemplate, pluginName)),
			HaveBodyThatContains("@bartoszmajsak"),
		)
	}
)

// Expecting creates mocks for the given matchers
func (b *MockPrBuilder) Expecting(mockCreators ...MockCreator) *MockPrBuilder {
	b.mockCreators = append(b.mockCreators, mockCreators...)
	return b
}

// Comment creates a gock matcher to check that there is a Post with a comment that complies with the given restrictions
func Comment(matherForPlugin func(pluginName string) SoftMatcher) MockCreator {
	return func(builder *MockPrBuilder) {
		basePostCommentMock(builder)(matherForPlugin(builder.pluginName))
	}
}

// CommentTo creates a gock matcher to check that there is a Post with a comment that complies with the given restrictions
func CommentTo(matchers ...SoftMatcher) MockCreator {
	return func(builder *MockPrBuilder) {
		basePostCommentMock(builder)(SoftlySatisfyAll(matchers...))
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

// Status creates a gock matcher to check that there is a Post with a status that complies with the given restrictions
func Status(matherForPlugin func(pluginName string) SoftMatcher) MockCreator {
	return func(builder *MockPrBuilder) {
		basePostStatusMock(builder)(matherForPlugin(builder.pluginName))
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
		basePostMock(path)(To(HaveBodyThatContains(labelContent)))
	}
}
