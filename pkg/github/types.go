package github

// EventType encapsulates all event types
type EventType string

// StatusContext enc
type StatusContext struct {
	BotName    string
	PluginName string
}

const (
	// EventGUID is sent by Github in a header of every webhook request.
	EventGUID = "event-GUID" // aligned with Prow fields
	// Event is sent by GitHub in a header of every webhook request.
	Event = "event-type" // aligned with Prow fields
	// RepoLogField is the repository from where the event came.
	RepoLogField = "github-repo"
	// SenderLogField is the username who caused the event to be sent.
	SenderLogField = "github-event-sender"
)

// These are possible State entries for a Status.
const (
	StatusPending = "pending"
	StatusSuccess = "success"
	StatusError   = "error"
	StatusFailure = "failure"
)

const (
	IssueComment = EventType("issue_comment") // nolint
	PullRequest  = EventType("pull_request")  // nolint
)
