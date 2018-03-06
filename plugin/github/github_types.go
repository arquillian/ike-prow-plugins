package githubevents

// EventType encapsulates all event types
type EventType string

const (
	IssueComment = EventType("issue_comment") // nolint
	PullRequest  = EventType("pull_request") // nolint
)
