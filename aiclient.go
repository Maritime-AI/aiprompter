package aiprompter

import "context"

// Role defines the role of the message in the chat completion request.
type Role string

// RoleAssistant and others are the possible roles for messages in the chat completion request.
const (
	RoleAssistant Role = "assistant"
	RoleUser      Role = "user"
)

// Message defines a message in the chat completion request.
type Message struct {
	Role    Role
	Message string
}

type Client interface {
	Prompt(ctx context.Context, systemPrompt string, msgs ...Message) (*PromptResponse, error)
}
