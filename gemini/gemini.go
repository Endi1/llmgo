package gemini

import (
	"context"
	"fmt"
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

func (geminiClient *GeminiClient) RunTools(ctx context.Context, model string, messages []client.ChatMessage, tools []client.Tool) (any, error) {
	var functionDeclarations []*genai.FunctionDeclaration = []*genai.FunctionDeclaration{}
	for _, tool := range tools {
		properties := map[string]*genai.Schema{}
		for _, property := range tool.Params().Properties {
			properties[property.Name] = &genai.Schema{
				Description: property.Description,
				Type:        genai.Type(property.Type),
			}
		}
		functionDeclarations = append(functionDeclarations, &genai.FunctionDeclaration{
			Name:        tool.Name(),
			Description: tool.Description(),
			Parameters: &genai.Schema{
				Required:   tool.Params().Required,
				Properties: properties,
				Type:       genai.TypeObject,
			},
		})
	}

	geminiTool := genai.Tool{FunctionDeclarations: functionDeclarations}
	config := genai.GenerateContentConfig{Tools: []*genai.Tool{&geminiTool}}
	chat, err := geminiClient.c.Chats.Create(ctx, model, &config, nil)

	if err != nil {
		return nil, err
	}

	parts := make([]genai.Part, len(messages))
	for _, msg := range messages {
		parts = append(parts, genai.Part{Text: msg.Content})
	}

	completion, err := chat.SendMessage(ctx, parts...)
	if err != nil {
		return client.ChatCompletion{}, err
	}

	for _, candidate := range completion.Candidates {
		for _, part := range candidate.Content.Parts {
			functionCall := part.FunctionCall
			if functionCall == nil {
				continue
			}
			for _, tool := range tools {
				if tool.Name() == functionCall.Name {
					return tool.Call(ctx, functionCall.Args)
				}
			}

		}
	}

	return nil, fmt.Errorf("No function call found")
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
