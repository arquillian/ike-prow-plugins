package test

import (
	"encoding/json"
	"fmt"

	ghservice "github.com/arquillian/ike-prow-plugins/pkg/github/service"
	"github.com/arquillian/ike-prow-plugins/pkg/utils"
	gogh "github.com/google/go-github/github"
	gock "gopkg.in/h2non/gock.v1"
)

// WithFiles sets the given payload containing changed files to the mocked PR
func (b *MockPrBuilder) WithFiles(jsonContent string, options ...RequestOption) *MockPrBuilder {
	b.mockFiles(jsonContent, options...)
	return b
}

// WithoutFiles sets an empty array as a payload representing changed files
func (b *MockPrBuilder) WithoutFiles(options ...RequestOption) *MockPrBuilder {
	b.mockFiles("[]", options...)
	return b
}

func (b *MockPrBuilder) mockFiles(content string, options ...RequestOption) {
	if len(options) == 0 {
		options = []RequestOption{perPage100, page1}
	}
	b.addMockCreator(b.mockGetForPR("pulls", "/files", content, options...))
}

// WithComments sets the given payload containing comments to the mocked PR
func (b *MockPrBuilder) WithComments(jsonContent string, options ...RequestOption) *MockPrBuilder {
	b.mockComments(jsonContent, options...)
	return b
}

// WithoutComments sets an empty array as a payload representing pr/issue comments
func (b *MockPrBuilder) WithoutComments(options ...RequestOption) *MockPrBuilder {
	b.mockComments("[]", options...)
	return b
}

func (b *MockPrBuilder) mockComments(content string, options ...RequestOption) {
	if len(options) == 0 {
		options = []RequestOption{perPage100, page1}
	}
	b.addMockCreator(b.mockGetForPR("issues", "/comments", content, options...))
}

// WithReviews sets the given payload containing list of reviews to the mocked PR
func (b *MockPrBuilder) WithReviews(jsonContent string, options ...RequestOption) *MockPrBuilder {
	b.mockReviews(jsonContent, options...)
	return b
}

// WithoutReviews sets an empty array as a payload representing pr reviews
func (b *MockPrBuilder) WithoutReviews(options ...RequestOption) *MockPrBuilder {
	b.mockReviews("[]", options...)
	return b
}

func (b *MockPrBuilder) mockReviews(content string, options ...RequestOption) {
	if len(options) == 0 {
		options = []RequestOption{perPage100, page1}
	}
	b.addMockCreator(b.mockGetForPR("pulls", "/reviews", content, options...))
}

// WithLabels sets the given payload containing list of labels to the mocked PR
func (b *MockPrBuilder) WithLabels(labelNames ...string) *MockPrBuilder {
	for _, labelName := range labelNames {
		b.pullRequest.Labels = append(b.pullRequest.Labels, &gogh.Label{Name: utils.String(labelName)})
	}
	return b
}

// WithoutLabels sets an empty array as a payload representing pr/issue labels
func (b *MockPrBuilder) WithoutLabels() *MockPrBuilder {
	b.pullRequest.Labels = make([]*gogh.Label, 0)
	return b
}

func (b *MockPrBuilder) mockGetForPR(targetType, suffix string, body string, options ...RequestOption) MockCreator {
	return func(builder *MockPrBuilder) {
		b.baseGetMock(fmt.Sprintf("%s/%s/%d", b.baseRepoPath(), targetType, *b.pullRequest.Number)+suffix, body, options...)
	}
}

// GhUser keeps name and user's permission
type GhUser struct {
	name       string
	permission string
}

// GhUserCreator creates an instance of GhUser
type GhUserCreator func(pullRequest *gogh.PullRequest) *GhUser

// ExternalUser creates an instance of GhUser with read permission
func ExternalUser(name string) func(pr *gogh.PullRequest) *GhUser {
	return func(pr *gogh.PullRequest) *GhUser {
		return &GhUser{name, "read"}
	}
}

// Admin creates an instance of GhUser with read admin
func Admin(name string) func(pr *gogh.PullRequest) *GhUser {
	return func(pr *gogh.PullRequest) *GhUser {
		return &GhUser{name, "admin"}
	}
}

// PrCreator creates an instance of GhUser with the given name and the name sets as PR's creator login
func PrCreator(name string) func(pr *gogh.PullRequest) *GhUser {
	return func(pr *gogh.PullRequest) *GhUser {
		*pr.User.Login = name
		return &GhUser{name, ""}
	}
}

// RequestedReviewer creates an instance of GhUser with the given name and appends the user to the list of PR's reviewers
func RequestedReviewer(name string) func(pr *gogh.PullRequest) *GhUser {
	return func(pr *gogh.PullRequest) *GhUser {
		pr.RequestedReviewers = append(pr.RequestedReviewers, createGhUser(name))
		return &GhUser{name, ""}
	}
}

// WithUsers sets the users to be mock as part of the MockPrBuilder
func (b *MockPrBuilder) WithUsers(userCreators ...GhUserCreator) *MockPrBuilder {
	for _, userCreator := range userCreators {
		b.mockUser(userCreator(b.pullRequest))
	}
	return b
}

func (b *MockPrBuilder) mockUser(user *GhUser) {
	if user.permission == "" {
		return
	}
	permission := gogh.RepositoryPermissionLevel{Permission: &user.permission, User: createGhUser(user.name)}
	content, err := json.Marshal(permission)
	if err != nil {
		b.errors = append(b.errors, err)
	}
	b.mockCreators = append(b.mockCreators,
		func(builder *MockPrBuilder) {
			b.mockGetForCollaborators(user.name, "/permission", string(content))
		})
}

func (b *MockPrBuilder) mockGetForCollaborators(user, suffix string, body string, options ...RequestOption) {
	b.baseGetMock(fmt.Sprintf("%s/collaborators/%s", b.baseRepoPath(), user)+suffix, body, options...)
}

// RequestOption add a option to a associated request
type RequestOption = func(request *gock.Request)

var perPage100 = func(request *gock.Request) {
	request.MatchParam("per_page", "100")
}
var page1 = func(request *gock.Request) {
	request.MatchParam("page", "1")
}

// WithoutConfigFiles sets that the associated mocked PR shouldn't contain any configuration file for the before-set plugin
func (b *MockPrBuilder) WithoutConfigFiles() *MockPrBuilder {
	return b.WithoutConfigFilesForPlugin(b.pluginName)
}

// WithoutConfigFilesForPlugin sets that the associated mocked PR shouldn't contain any configuration file for the given plugin
func (b *MockPrBuilder) WithoutConfigFilesForPlugin(pluginName string) *MockPrBuilder {
	configsToMock := []string{"%s.yaml", "%s.yml"}

	for _, config := range configsToMock {
		b.WithoutRawFiles(ghservice.ConfigHome + fmt.Sprintf(config, pluginName))
	}
	return b
}

// WithoutMessageFiles sets that the associated mocked PR shouldn't contain status messages with the given names
func (b *MockPrBuilder) WithoutMessageFiles(fileNames ...string) *MockPrBuilder {
	for _, fileName := range fileNames {
		b.WithoutRawFiles(ghservice.ConfigHome + fileName)
	}
	return b
}

// WithConfigFile sets that the associated mocked PR should contain the given configuration file
func (b *MockPrBuilder) WithConfigFile(configMock func(builder *MockPrBuilder)) *MockPrBuilder {
	configMock(b)
	return b
}

// KeyValue is a function returning key and value
type KeyValue func() (string, string)

// Param represents a key-value pair
func Param(key, value string) KeyValue {
	return func() (string, string) {
		return key, value
	}
}

// Containing creates a json content of a config file from the given params
func Containing(params ...KeyValue) string {
	var content string
	for _, param := range params {
		key, value := param()
		content += fmt.Sprintf("%s\n%s : %s", content, key, value)
	}
	return content
}

// ConfigYaml creates a representation of a config file with yaml suffix
func ConfigYaml(content string) func(builder *MockPrBuilder) {
	return func(builder *MockPrBuilder) {
		builder.WithRawFile(ghservice.ConfigHome+builder.pluginName+".yaml", content)
	}
}

// ConfigYml creates a representation of a config file with yml suffix
func ConfigYml(content string) func(builder *MockPrBuilder) {
	return func(builder *MockPrBuilder) {
		builder.WithRawFile(ghservice.ConfigHome+builder.pluginName+".yml", content)
	}
}

// WithoutRawFiles sets that the associated mocked PR should not contain the given files
func (b *MockPrBuilder) WithoutRawFiles(fileNames ...string) *MockPrBuilder {
	for _, path := range fileNames {
		path := path
		b.addMockCreator(func(builder *MockPrBuilder) {
			builder.getBaseRawFilesMock(path).
				Reply(404)
		})
	}
	return b
}

// WithRawFile sets that the associated mocked PR should contain the given files
func (b *MockPrBuilder) WithRawFile(fileName, content string) *MockPrBuilder {
	b.addMockCreator(func(builder *MockPrBuilder) {
		builder.getBaseRawFilesMock(fileName).
			Reply(200).
			BodyString(content)
	})
	return b
}

func (b *MockPrBuilder) getBaseRawFilesMock(path string) *gock.Request {
	pr := b.pullRequest
	return gock.New("https://raw.githubusercontent.com").
		Path(fmt.Sprintf("%s/%s/%s/%s", *pr.Base.Repo.Owner.Login, *pr.Base.Repo.Name, *pr.Head.SHA, path))
}

func (b *MockPrBuilder) baseGetMock(path, body string, options ...RequestOption) {
	baseGockMock(func(request *gock.Request) { request.Get(path + "$") }, options...).
		Reply(200).
		BodyString(body)
}

func (b *MockPrBuilder) baseRepoPath() string {
	repository := b.pullRequest.Base.Repo
	return fmt.Sprintf("/repos/%s/%s", *repository.Owner.Login, *repository.Name)
}
