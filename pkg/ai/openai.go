package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type OpenAIClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	return &OpenAIClient{
		baseURL: "https://api.openai.com/v1",
		apiKey:  apiKey,
		client: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *OpenAIClient) IsConfigured() bool {
	return c.apiKey != ""
}

func (c *OpenAIClient) GetSetupInstructions() string {
	return `AI features require an OpenAI API key.

Get an API key at: https://platform.openai.com/

Set your API key:
  export OPENAI_API_KEY="your-api-key"

Or pass it via CLI:
  nexus ai --api-key "your-api-key" <command>`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (c *OpenAIClient) CreateMessage(ctx context.Context, prompt string, maxTokens int) (string, error) {
	if !c.IsConfigured() {
		return "", fmt.Errorf("API key not configured\n\n%s", c.GetSetupInstructions())
	}

	if maxTokens == 0 {
		maxTokens = 1024
	}

	reqBody := chatRequest{
		Model: "gpt-4o-mini",
		Messages: []chatMessage{
			{Role: "user", Content: prompt},
		},
		MaxTokens: maxTokens,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var msgResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&msgResp); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(msgResp.Choices) == 0 {
		return "", fmt.Errorf("no content in response")
	}

	return msgResp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) GenerateRequestBody(ctx context.Context, schema string) (string, error) {
	prompt := fmt.Sprintf(`Generate a valid JSON request body based on this schema:

%s

Return only the JSON, no explanation or markdown formatting.`, schema)

	return c.CreateMessage(ctx, prompt, 2048)
}

func (c *OpenAIClient) GenerateTests(ctx context.Context, apiSpec string) ([]string, error) {
	prompt := fmt.Sprintf(`Generate test assertions for this API endpoint:

%s

Return test assertions in this format, one per line:
- status == 200
- body.id != null
- response.time < 1000

Return ONLY the assertions, no explanation.`, apiSpec)

	response, err := c.CreateMessage(ctx, prompt, 1024)
	if err != nil {
		return nil, err
	}

	tests := []string{}
	for _, line := range splitLines(response) {
		trimmed := trim(line)
		if trimmed != "" && (startsWith(trimmed, "-") || startsWith(trimmed, "*")) {
			tests = append(tests, trimAfter(trimmed, 1))
		} else if trimmed != "" && containsAny(trimmed, "status", "body", "response") {
			tests = append(tests, trimmed)
		}
	}

	return tests, nil
}

func (c *OpenAIClient) SuggestOptimizations(ctx context.Context, requestInfo string) ([]string, error) {
	prompt := fmt.Sprintf(`Analyze this HTTP request and suggest optimizations:

%s

Suggest optimizations for:
- Caching headers (Cache-Control, ETag, etc.)
- Compression (Accept-Encoding, Content-Encoding)
- Connection reuse (Connection, Keep-Alive)
- Request size reduction

Return suggestions as a bulleted list with specific header recommendations.`, requestInfo)

	response, err := c.CreateMessage(ctx, prompt, 2048)
	if err != nil {
		return nil, err
	}

	suggestions := []string{}
	for _, line := range splitLines(response) {
		trimmed := trim(line)
		if trimmed != "" && (startsWith(trimmed, "-") || startsWith(trimmed, "*")) {
			suggestions = append(suggestions, trimAfter(trimmed, 1))
		}
	}

	return suggestions, nil
}

func (c *OpenAIClient) GenerateFromNaturalLanguage(ctx context.Context, description string) (string, error) {
	prompt := fmt.Sprintf(`Convert this natural language description into a NEXUS-API collection YAML:

Description: %s

Return a complete YAML collection with:
- Collection name
- Base URL (infer from description or use placeholder)
- At least one request with method, URL, headers, and body if applicable
- Test assertions

Return ONLY the YAML, no explanation or markdown formatting.`, description)

	return c.CreateMessage(ctx, prompt, 4096)
}

func (c *OpenAIClient) AnalyzeAPIChanges(ctx context.Context, oldSpec, newSpec string) (string, error) {
	prompt := fmt.Sprintf(`Compare these two API specifications and identify breaking changes:

OLD SPEC:
%s

NEW SPEC:
%s

List all changes categorized as:
- Breaking changes (removed endpoints, changed signatures)
- Non-breaking changes (new endpoints, optional fields)
- Deprecations

Be specific about what changed.`, oldSpec, newSpec)

	return c.CreateMessage(ctx, prompt, 4096)
}

func trim(s string) string {
	result := ""
	start := 0
	end := len(s)

	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}

	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}

	result = s[start:end]
	return result
}

func startsWith(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}

func trimAfter(s string, pos int) string {
	if pos >= len(s) {
		return ""
	}
	return trim(s[pos:])
}
