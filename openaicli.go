package aiprompter

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/sashabaranov/go-openai"
)

const (
	defaultOpenAIModel = openai.GPT4oMini
)

// OpenAIOption is a functional option for configuring the Client.
// It allows customization of the client properties.
type OpenAIOption func(c *OpenAIClient)

// WithOpenAIModel is an Option to specify a custom model for the OpenAI client.
// Example usage: WithModel("gpt-4").
func WithOpenAIModel(model string) OpenAIOption {
	return func(c *OpenAIClient) {
		c.model = model
	}
}

// OpenAIClient wraps the OpenAI client with additional configuration options.
type OpenAIClient struct {
	client *openai.Client
	model  string
	apiKey string
}

// NewClient initializes a new Client with the provided API key and options.
// Default model is set to `model4oMini` unless overridden by an Option.
func NewOpenAIClient(apiKey string, opts ...OpenAIOption) *OpenAIClient {
	c := &OpenAIClient{
		apiKey: apiKey,
		model:  openai.GPT4oMini,
		client: openai.NewClient(apiKey),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// CloneClientWithModel clones the client with a new model.
func (c *OpenAIClient) CloneClientWithModel(model string) Client {
	return NewOpenAIClient(c.apiKey, WithOpenAIModel(model))
}

type PromptResponse struct {
	Response     string
	TotalTokens  int
	PromptTokens int
	RequestData  map[string]any
}

// Prompt sends a chat completion request to the OpenAI API with the given prompt.
// It returns the generated response as a string or an error if the request fails.
func (c *OpenAIClient) Prompt(ctx context.Context, msgs []Message, opts ...PromptOption) (*PromptResponse, error) {
	now := time.Now()

	pOpts := &PromptOptions{}
	for _, opt := range opts {
		opt(pOpts)
	}

	var cmsgs []openai.ChatCompletionMessage
	if pOpts.systemPrompt != nil {
		cmsgs = append(cmsgs, openai.ChatCompletionMessage{
			Role:    "system",
			Content: *pOpts.systemPrompt,
		})
	}

	// add chat history
	for _, m := range msgs {
		cmsgs = append(cmsgs, openai.ChatCompletionMessage{
			Role:    string(m.Role),
			Content: m.Message,
		})
	}

	fmt.Println("Chat Messages:", cmsgs)

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Temperature: 0.0,
		ResponseFormat: &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		},
		Messages: cmsgs,
	}

	requestData := map[string]interface{}{
		"model": c.model,
		//"messages":    cmsgs,
		"temperature": 0.0,
	}
	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to chat with model with response object: %s", resp.Object)
	}

	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices returned")
	}

	cmsgs = append(cmsgs, openai.ChatCompletionMessage{
		Role:    "assistant",
		Content: resp.Choices[0].Message.Content,
	})

	requestData["messages"] = cmsgs
	requestData["execution_time_in_secs"] = time.Since(now).Seconds()

	return &PromptResponse{
		Response:     resp.Choices[0].Message.Content,
		TotalTokens:  resp.Usage.TotalTokens,
		PromptTokens: resp.Usage.PromptTokens,
		RequestData:  requestData,
	}, nil
}
