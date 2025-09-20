package gemini

import (
	"context"
	"flag"
	"fmt"
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

type WeatherTool struct {
}

func (w *WeatherTool) Name() string {
	return "get_weather"
}

func (w *WeatherTool) Description() string {
	return "Get the weather for a location and a particular day. The inputs are the location name and the day of the week"
}

func (w *WeatherTool) Params() client.ParamsSchema {
	return client.ParamsSchema{
		Properties: map[string]*client.ParamProperty{
			"location": &client.ParamProperty{Name: "location", Type: "string", Description: "The location for which to get the weather"},
			"day":      &client.ParamProperty{Name: "day", Type: "string", Description: "The day of the week for which to get the weather"},
		},
		Required: []string{"location", "day"},
	}
}

func (w *WeatherTool) Call(ctx context.Context, params map[string]any) (any, error) {
	return fmt.Sprintf("The weather in %s on %s is fine", params["location"], params["day"]), nil
}

func TestGeminiRunTools(t *testing.T) {
	if !*integration {
		t.Skip("skipping integration tests")
	}

	ctx := context.Background()
	config := GeminiConfig{Region: os.Getenv("GEMINI_REGION"), Project: os.Getenv("GEMINI_PROJECT")}
	var geminiClient client.Client
	geminiClient = New(config)

	model := "gemini-2.5-flash"
	response, err := geminiClient.RunTools(ctx, model, []client.ChatMessage{{Content: "what is the weather in Rome on Sunday?"}}, []client.Tool{&WeatherTool{}})

	if err != nil {
		t.Errorf("geminiClient.RunTools failed with error %s", err)
	}

	if response != "The weather in Rome on Sunday is fine" {
		t.Errorf("geminiClients.RunTools responded with the wrong response %s", response)
	}
}
