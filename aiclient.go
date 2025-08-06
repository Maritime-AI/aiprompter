package aiprompter

import (
	"bytes"
	"context"
)

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

// PromptOption is a functional option for configuring the prompt options.
type PromptOption func(*PromptOptions)

// WithRunID sets a custom run ID for the prompt operation.
func WithRunID(runID string) PromptOption {
	return func(o *PromptOptions) {
		o.runID = runID
	}
}

// WithLogBuffer sets a buffer to log prompt and response data.
func WithLogBuffer(b *bytes.Buffer) PromptOption {
	return func(o *PromptOptions) {
		o.logBuffer = b
	}
}

// WithSystemPrompt sets a system prompt for the AI model.
func WithSystemPrompt(p *string) PromptOption {
	return func(o *PromptOptions) {
		o.systemPrompt = p
	}
}

// OpenAIOption defines a function signature used to modify OpenAIClient options.
type PromptOptions struct {
	systemPrompt *string
	logBuffer    *bytes.Buffer
	runID        string
}

type Client interface {
	Prompt(ctx context.Context, msgs []Message, opts ...PromptOption) (*PromptResponse, error)
}
