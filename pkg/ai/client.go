package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type GenerateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (c *Client) Generate(ctx context.Context, model, prompt string) (string, error) {
	reqBody := GenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return genResp.Response, nil
}

func (c *Client) GenerateRequestBody(ctx context.Context, schema string) (string, error) {
	prompt := fmt.Sprintf(`Generate a valid JSON request body based on this schema:

%s

Return only the JSON, no explanation.`, schema)

	return c.Generate(ctx, "llama2", prompt)
}

// CreateMessage implements AIClient.CreateMessage for local LLM adapters.
func (c *Client) CreateMessage(ctx context.Context, prompt string, maxTokens int) (string, error) {
	// The local adapter uses a simple generate endpoint; ignore maxTokens for now.
	return c.Generate(ctx, "llama2", prompt)
}

func (c *Client) GenerateTests(ctx context.Context, apiSpec string) ([]string, error) {
	prompt := fmt.Sprintf(`Generate test assertions for this API endpoint:

%s

Return test assertions in this format:
- status == 200
- body.id != null
- response.time < 1000

Return only the assertions, one per line.`, apiSpec)

	response, err := c.Generate(ctx, "llama2", prompt)
	if err != nil {
		return nil, err
	}

	tests := []string{}
	for _, line := range splitLines(response) {
		if line != "" && (containsAny(line, "status", "body", "response")) {
			tests = append(tests, line)
		}
	}

	return tests, nil
}

func (c *Client) SuggestOptimizations(ctx context.Context, requestInfo string) ([]string, error) {
	prompt := fmt.Sprintf(`Analyze this HTTP request and suggest optimizations:

%s

Suggest optimizations for:
- Caching headers
- Compression
- Connection reuse
- Request size reduction

Return suggestions as a bulleted list.`, requestInfo)

	response, err := c.Generate(ctx, "llama2", prompt)
	if err != nil {
		return nil, err
	}

	suggestions := []string{}
	for _, line := range splitLines(response) {
		if line != "" && (line[0] == '-' || line[0] == '*') {
			suggestions = append(suggestions, line)
		}
	}

	return suggestions, nil
}

func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func containsAny(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstr(s, substr))
}

func findSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
