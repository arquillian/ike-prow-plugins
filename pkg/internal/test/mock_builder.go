package test

import (
	"encoding/json"
	"fmt"

	"github.com/arquillian/ike-prow-plugins/pkg/command"
	"github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/scm"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
	"github.com/onsi/gomega"
	"gopkg.in/h2non/gock.v1"
)

// MockPluginTemplate keeps plugin name information
type MockPluginTemplate struct {
	pluginName string
}

// NewMockPluginTemplate creates an instance of MockPluginTemplate
func NewMockPluginTemplate(pluginName string) MockPluginTemplate {
	return MockPluginTemplate{pluginName: pluginName}
}

// MockPrBuilder keeps information about pr, plugin and all mock creators to be initialized
type MockPrBuilder struct {
	pluginName   string
	pullRequest  *gogh.PullRequest
	mockCreators []MockCreator
	errors       []error
}

// MockCreator creates a gock mock
type MockCreator func(builder *MockPrBuilder)

// LoadedFromDefaultJSON loads json from a location test_fixtures/github_calls/prs/pr_details.json
func LoadedFromDefaultJSON() string {
	return LoadedFrom("test_fixtures/github_calls/prs/pr_details.json")
}

// LoadedFromDefaultStruct loads a marshaled instance of default pull request
func LoadedFromDefaultStruct() string {
	pr, _ := json.Marshal(&gogh.PullRequest{
		Number: utils.Int(1),
		User:   createGhUser("bartoszmajsak-test"),
		Base: &gogh.PullRequestBranch{
			Repo: &gogh.Repository{
				Owner: createGhUser("bartoszmajsak"),
				Name:  utils.String("wfswarm-booster-pipeline-test"),
			},
		},
		Head: &gogh.PullRequestBranch{
			SHA: utils.String("df8e5cd15f05e1d975e17df322b9babedccf0a1a"),
		},
	})
	return string(pr)
}

// MockPr creates a builder for mocking a pull request calls
func MockPr(jsonContent string) *MockPrBuilder {
	return NewMockPluginTemplate("any").MockPr(jsonContent)
}

// MockPr creates a builder for mocking a pull request calls
func (t MockPluginTemplate) MockPr(jsonContent string) *MockPrBuilder {
	builder := &MockPrBuilder{pluginName: t.pluginName}
	if err := json.Unmarshal([]byte(jsonContent), &builder.pullRequest); err != nil {
		builder.errors = []error{err}
	}
	builder.mockCreators = []MockCreator{
		func(builder *MockPrBuilder) {
			content, err := json.Marshal(builder.pullRequest)
			if err != nil {
				builder.errors = append(builder.errors, err)
			}
			builder.baseGetMock(fmt.Sprintf("%s/pulls/%d", builder.baseRepoPath(), *builder.pullRequest.Number), string(content))
		},
	}
	return builder
}

// WithTitle sets the given title to the mocked pull request
func (b *MockPrBuilder) WithTitle(title string) *MockPrBuilder {
	b.pullRequest.Title = &title
	return b
}

// Create initializes the gock mocks based on the predefined information
func (b *MockPrBuilder) Create() *PrMock {
	for _, mock := range b.mockCreators {
		mock(b)
	}
	gomega.Expect(b.errors).To(gomega.BeEmpty())

	return &PrMock{PullRequest: b.pullRequest}
}

// PrMock keeps the mocked pull request
type PrMock struct {
	PullRequest *gogh.PullRequest
}

// SenderCreator creates a gogh.User
type SenderCreator func(pullRequest *gogh.PullRequest) *gogh.User

// SentByReviewer creates a user that is stored at the first place in the list of requested reviewers of the mocked PR
var SentByReviewer = func(pullRequest *gogh.PullRequest) *gogh.User {
	return pullRequest.RequestedReviewers[0]
}

// SentByRepoOwner creates a user that is stored as a owner of the mocked PR's repository
var SentByRepoOwner = func(pullRequest *gogh.PullRequest) *gogh.User {
	return pullRequest.Base.Repo.Owner
}

// SentByPrCreator creates a user that is stored as a creator of the mocked PR
var SentByPrCreator = func(pullRequest *gogh.PullRequest) *gogh.User {
	return pullRequest.User
}

// SentBy creates a user with the given name
func SentBy(name string) SenderCreator {
	return func(pullRequest *gogh.PullRequest) *gogh.User {
		return createGhUser(name)
	}
}

// CreateCommentEvent based on the mocked PR information creates a marshaled payload of the IssueCommentEvent struct
func (pr *PrMock) CreateCommentEvent(userCreator SenderCreator, content, action string) []byte {
	event := gogh.IssueCommentEvent{
		Action: utils.String(action),
		Issue: &gogh.Issue{
			Number: pr.PullRequest.Number,
		},
		Comment: &gogh.IssueComment{
			Body: utils.String(content),
		},
		Repo:   pr.PullRequest.Base.Repo,
		Sender: userCreator(pr.PullRequest),
	}
	return marshalObject(event)
}

func createGhUser(name string) *gogh.User {
	return &gogh.User{Login: utils.String(name)}
}

func marshalObject(toMarshal interface{}) []byte {
	json, err := json.Marshal(toMarshal)
	gomega.Ω(err).ShouldNot(gomega.HaveOccurred())
	return json
}

// CreatePullRequestEvent based on the mocked PR information creates a marshaled payload of the PullRequestEvent struct
func (pr *PrMock) CreatePullRequestEvent(action string) []byte {
	event := gogh.PullRequestEvent{
		Action:      utils.String(action),
		Number:      pr.PullRequest.Number,
		PullRequest: pr.PullRequest,
		Repo:        pr.PullRequest.Base.Repo,
		Sender:      pr.PullRequest.User,
	}
	return marshalObject(event)
}

// CreateUserPermissionService based on the mocked PR information creates an instance of PermissionService
func (pr *PrMock) CreateUserPermissionService(userName string) *command.PermissionService {
	return &command.PermissionService{
		Client: NewDefaultGitHubClient(),
		User:   userName,
		PRLoader: &ghservice.PullRequestLazyLoader{
			Client:    NewDefaultGitHubClient(),
			RepoOwner: *pr.PullRequest.Base.Repo.Owner.Login,
			RepoName:  *pr.PullRequest.Base.Repo.Name,
			Number:    *pr.PullRequest.Number,
		},
	}
}

// SubMockBuilder represents a builder to be used for mocking of a particular PR parameter
type SubMockBuilder struct {
	prBuilder *MockPrBuilder
}

// MockConfig creates an instance of SubMockBuilder for the given configuration
func (t MockPluginTemplate) MockConfig(configMock func(builder *MockPrBuilder)) *SubMockBuilder {
	builder := &SubMockBuilder{prBuilder: &MockPrBuilder{pluginName: t.pluginName}}
	builder.prBuilder.WithConfigFile(configMock)
	return builder
}

// ForChange defines that the predefined information should be applicable for the given scm.RepositoryChange
func (b *SubMockBuilder) ForChange(change scm.RepositoryChange) {
	b.prBuilder.pullRequest = changeToPr(change)
	b.prBuilder.Create()
}

func baseGockMock(method RequestOption, options ...RequestOption) *gock.Request {
	request := gock.New("https://api.github.com")
	method(request)

	for _, opt := range options {
		opt(request)
	}
	return request
}

func changeToPr(change scm.RepositoryChange) *gogh.PullRequest {
	return &gogh.PullRequest{
		Base: &gogh.PullRequestBranch{
			Repo: &gogh.Repository{
				Owner: createGhUser(change.Owner),
				Name:  &change.RepoName,
			},
		},
		Head: &gogh.PullRequestBranch{
			SHA: &change.Hash,
		},
	}
}

func (b *MockPrBuilder) addMockCreator(creator MockCreator) {
	b.mockCreators = append(b.mockCreators, creator)
}
