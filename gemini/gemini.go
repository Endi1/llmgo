package gemini

import (
	"context"
	"github.com/endi1/llmgo/client"
	"google.golang.org/genai"
	"log"
)

type GeminiClient struct {
	c *genai.Client
}

type GeminiConfig struct {
	Region  string
	Project string
}

func (geminiClient *GeminiClient) Completion(ctx context.Context, model string, messages []client.ChatMessage) (client.ChatCompletion, error) {
	chat, _ := geminiClient.c.Chats.Create(ctx, model, nil, nil)

	parts := make([]genai.Part, len(messages))
	for _, msg := range messages {
		parts = append(parts, genai.Part{Text: msg.Content})
	}

	completion, err := chat.SendMessage(ctx, parts...)
	if err != nil {
		return client.ChatCompletion{}, err
	}
	return client.ChatCompletion{
		Text:         completion.Candidates[0].Content.Parts[0].Text,
		InputTokens:  0,
		OutputTokens: 0,
	}, nil
}

func (geminiClient *GeminiClient) Stream(ctx context.Context, model string, messages []client.ChatMessage) <-chan client.StreamResult {
	resultCh := make(chan client.StreamResult)

	go func() {
		defer close(resultCh)

		chat, err := geminiClient.c.Chats.Create(ctx, model, nil, nil)
		if err != nil {
			select {
			case resultCh <- client.StreamResult{Error: err}:
			case <-ctx.Done():
			}
			return
		}

		parts := make([]genai.Part, len(messages))
		for _, msg := range messages {
			parts = append(parts, genai.Part{Text: msg.Content})
		}

		stream := chat.SendMessageStream(ctx, parts...)
		for token, err := range stream {
			if err != nil {
				select {
				case resultCh <- client.StreamResult{Error: err}:
				case <-ctx.Done():
				}
				return
			}

			completion := client.ChatCompletion{
				Text:         token.Candidates[0].Content.Parts[0].Text,
				InputTokens:  token.UsageMetadata.PromptTokenCount,
				OutputTokens: token.UsageMetadata.CandidatesTokenCount,
			}

			select {
			case resultCh <- client.StreamResult{Completion: completion}:
			case <-ctx.Done():
				return
			}
		}

		// Signal completion
		select {
		case resultCh <- client.StreamResult{Done: true}:
		case <-ctx.Done():
		}
	}()

	return resultCh
}

func New(config GeminiConfig) *GeminiClient {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  config.Project,
		Location: config.Region,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	return &GeminiClient{c: client}

}
