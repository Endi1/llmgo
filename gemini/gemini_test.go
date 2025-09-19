package gemini

import (
	"context"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/endi1/llmgo/client"
)

var integration = flag.Bool("integration", false, "run integration tests")

func TestGeminiSimpleCompletion(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	ctx := context.Background()
	config := GeminiConfig{Region: os.Getenv("GEMINI_REGION"), Project: os.Getenv("GEMINI_PROJECT")}
	var geminiClient client.Client
	geminiClient = New(config)

	model := "gemini-2.5-flash"
	completion, err := geminiClient.Completion(ctx, model, []client.ChatMessage{{Content: "hello, how are you?"}})
	if err != nil {
		t.Errorf("geminiClient.Completion returned an error: %s", err.Error())
	}

	if completion.Text == "" {
		t.Error("geminiClient.Completion returned an empty string")
	}

}

func TestGeminiStreaming(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	ctx := context.Background()
	config := GeminiConfig{Region: os.Getenv("GEMINI_REGION"), Project: os.Getenv("GEMINI_PROJECT")}
	var geminiClient client.Client
	geminiClient = New(config)

	model := "gemini-2.5-flash"
	results := geminiClient.Stream(ctx, model, []client.ChatMessage{{Content: "hello, tell me a short story"}})
	var fullCompletionBuffer strings.Builder
	for result := range results {
		if result.Error != nil {
			t.Errorf("geminiClient.Stream failed with error %s", result.Error.Error())
			break
		}
		fullCompletionBuffer.WriteString(result.Completion.Text)
	}

	if fullCompletionBuffer.String() == "" {
		t.Error("geminiClient.Stream streamed an empty string")
	}
}
