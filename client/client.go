package client

import (
	"context"
)

type ChatCompletion struct {
	Text         string
	InputTokens  int32
	OutputTokens int32
}

type ChatMessage struct {
	Role    string
	Content string
}

type StreamResult struct {
	Completion ChatCompletion
	Error      error
	Done       bool
}

type Client interface {
	Completion(ctx context.Context, model string, messages []ChatMessage) (ChatCompletion, error)
	Stream(ctx context.Context, model string, messages []ChatMessage) <-chan StreamResult
}
