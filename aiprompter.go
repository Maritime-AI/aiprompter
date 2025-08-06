package aiprompter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Option defines a function signature used to modify Options.
type Option func(*Options)

// AIPrompter is a wrapper around an AI chat client, used to manage prompts and logging.
type AIPrompter struct {
	chatCli Client
}

// NewAIPrompter creates a new instance of AIPrompter with the provided AI chat client.
func NewAIPrompter(chatCli Client) *AIPrompter {
	return &AIPrompter{
		chatCli: chatCli,
	}
}

// Options defines configuration settings for a single prompt execution.
type Options struct {
	RunID     string
	LogBuffer *bytes.Buffer // Assuming Buffer is a type that implements io.Writer
}

// SinglePrompt prompts the AI model with a single system prompt and optional chat history.
// It returns the AI's response or an error if the operation fails.
func (s *AIPrompter) SinglePrompt(ctx context.Context, systemPrompt string,
	chatHistory []Message, opts ...Option) (*PromptResponse, error) {

	rOpts := &Options{
		RunID: uuid.NewString(),
	}

	// apply options
	for _, opt := range opts {
		opt(rOpts)
	}

	pr, err := s.prompt(ctx, rOpts, systemPrompt, chatHistory...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to process single prompt")
	}

	return pr, nil
}

func (s *AIPrompter) prompt(ctx context.Context, rOpts *Options, systemPrompt string, chatHistory ...Message) (*PromptResponse, error) {
	resp, err := s.chatCli.Prompt(ctx, systemPrompt, chatHistory...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prompt model")
	}

	if err := s.appendToLogFile(rOpts, systemPrompt, resp); err != nil {
		return nil, errors.Wrap(err, "failed to append to log file")
	}

	return resp, nil
}

func (s *AIPrompter) appendToLogFile(rOpts *Options, promptStr string, response *PromptResponse) error {
	if rOpts.LogBuffer == nil {
		return nil
	}

	data, err := json.Marshal(response)
	if err != nil {
		return errors.Wrap(err, "failed to marshal prompt response")
	}

	logData := fmt.Sprintf("RunID: %s\n Prompt: %s\nResponse: %s\n\n",
		rOpts.RunID, promptStr, string(data))

	if _, err := rOpts.LogBuffer.WriteString(logData); err != nil {
		return errors.Wrap(err, "failed to write to log buffer")
	}

	return nil
}
