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

type ParamType string

type ParamsSchema struct {
	Properties map[string]*ParamProperty
	Required []string
}

type ParamProperty struct {
	Name string
	Type ParamType
	Description string
}

type Tool interface {
	Name() string
	Description() string
	Params() ParamsSchema
	Call(ctx context.Context, params map[string]any) (any, error)
}

type Client interface {
	Completion(ctx context.Context, model string, messages []ChatMessage) (ChatCompletion, error)
	Stream(ctx context.Context, model string, messages []ChatMessage) <-chan StreamResult
	RunTools(ctx context.Context, model string, messages []ChatMessage, tools []Tool) (any, error)
}
