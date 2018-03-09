package github

// EventType encapsulates all event types
type EventType string

// StatusContext enc
type StatusContext struct {
	BotName    string
	PluginName string
}

const (
	IssueComment = EventType("issue_comment") // nolint
	PullRequest  = EventType("pull_request") // nolint
)
