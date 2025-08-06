package aiprompter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

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

// SinglePrompt prompts the AI model with a single system prompt and optional chat history.
// It returns the AI's response or an error if the operation fails.
func (s *AIPrompter) SinglePrompt(ctx context.Context,
	chatHistory []Message, opts ...PromptOption) (*PromptResponse, error) {

	rOpts := &PromptOptions{
		runID: uuid.NewString(),
	}

	// apply options
	for _, opt := range opts {
		opt(rOpts)
	}

	pr, err := s.prompt(ctx, rOpts, chatHistory...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to process single prompt")
	}

	return pr, nil
}

// StreamPromptChunksResponse defines the response structure for streaming prompt chunks.
type StreamPromptChunksResponse struct {
	Response string
	Error    error
}

// StreamPromptChunks takes already chunked user input (as []string),
// and streams each response's text via responseChan. Any error is sent through errorChan and halts further processing.
func (s *AIPrompter) StreamPromptChunks(ctx context.Context, chunks []string, opts ...PromptOption) <-chan StreamPromptChunksResponse {

	responseChan := make(chan StreamPromptChunksResponse)

	go func() {
		defer close(responseChan)

		if len(chunks) == 0 {
			responseChan <- StreamPromptChunksResponse{Error: errors.New("no chunks provided")}
			return
		}

		for i, chunk := range chunks {
			select {
			case <-ctx.Done():
				responseChan <- StreamPromptChunksResponse{Error: ctx.Err()}
				return
			default:
				cm := []Message{
					{
						Role:    RoleUser,
						Message: chunk,
					},
				}

				resp, err := s.SinglePrompt(ctx, cm, opts...)
				if err != nil {
					responseChan <- StreamPromptChunksResponse{Error: errors.Wrapf(err, "failed to process chunk %d", i+1)}
					return
				}

				if resp != nil && strings.TrimSpace(resp.Response) != "" {
					responseChan <- StreamPromptChunksResponse{Response: resp.Response}
				}
			}
		}
	}()

	return responseChan
}

func (s *AIPrompter) prompt(ctx context.Context, rOpts *PromptOptions, chatHistory ...Message) (*PromptResponse, error) {

	var pOpts []PromptOption
	if rOpts.systemPrompt != nil {
		pOpts = append(pOpts, WithSystemPrompt(rOpts.systemPrompt))
	}

	resp, err := s.chatCli.Prompt(ctx, chatHistory, pOpts...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prompt model")
	}

	if err := s.appendToLogFile(rOpts, resp); err != nil {
		return nil, errors.Wrap(err, "failed to append to log file")
	}

	return resp, nil
}

func (s *AIPrompter) appendToLogFile(pOpts *PromptOptions, response *PromptResponse) error {
	if pOpts.logBuffer == nil {
		return nil
	}

	data, err := json.Marshal(response)
	if err != nil {
		return errors.Wrap(err, "failed to marshal prompt response")
	}

	var logDataStr string
	if pOpts.runID != "" {
		logDataStr += fmt.Sprintf("RunID: %s\n", pOpts.runID)
	}
	if pOpts.systemPrompt != nil {
		logDataStr += fmt.Sprintf("System Prompt: %s\n", *pOpts.systemPrompt)
	}
	logDataStr += fmt.Sprintf("Response: %s\n\n", string(data))
	if _, err := pOpts.logBuffer.WriteString(logDataStr); err != nil {
		return errors.Wrap(err, "failed to write to log buffer")
	}

	return nil
}
