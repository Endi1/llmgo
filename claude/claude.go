package claude

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/endi1/llmgo/client"
)

type ClaudeClient struct {
	c *anthropic.Client
}

type ClaudeConfig struct {
	Region string
	APIKey string
}

func (claudeClient *ClaudeClient) Completion(ctx context.Context, model string, messages []client.ChatMessage) client.ChatCompletion {
	anthropicMessages := make([]anthropic.MessageParam, len(messages))
	for _, msg := range messages {
		anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
	}
	response, err := claudeClient.c.Messages.New(ctx, anthropic.MessageNewParams{
		Messages: anthropicMessages,
	})

	if err != nil {
		panic(err)
	}

	return client.ChatCompletion{
		Text:         response.Content[0].Text,
		InputTokens:  int32(response.Usage.InputTokens),
		OutputTokens: int32(response.Usage.OutputTokens),
	}
}

func (c *ClaudeClient) Stream(ctx context.Context, model string, messages []client.ChatMessage) <-chan client.StreamResult {

	resultCh := make(chan client.StreamResult)
	go func() {
		defer close(resultCh)

		anthropicMessages := make([]anthropic.MessageParam, len(messages))
		for _, msg := range messages {
			anthropicMessages = append(anthropicMessages, anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content)))
		}

		stream := c.c.Messages.NewStreaming(ctx, anthropic.MessageNewParams{Messages: anthropicMessages})
		for stream.Next() {
			event := stream.Current()
			switch eventVariant := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				switch deltaVariant := eventVariant.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					completion := client.ChatCompletion{
						Text:         deltaVariant.Text,
						InputTokens:  int32(event.Usage.InputTokens),
						OutputTokens: int32(event.Usage.OutputTokens),
					}

					select {
					case resultCh <- client.StreamResult{Completion: completion}:
					case <-ctx.Done():
						return
					}
				}

			}

		}

	}()
	return resultCh
}

func New(config ClaudeConfig) *ClaudeClient {
	client := anthropic.NewClient(option.WithAPIKey(config.APIKey))
	return &ClaudeClient{c: &client}
}
