package github_events

type EventType string

const (
	IssueComment = EventType("issue_comment")
	PullRequest  = EventType("pull_request")
)
