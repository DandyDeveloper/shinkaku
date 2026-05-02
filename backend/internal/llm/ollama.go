// Package llm provides a client for Ollama's local LLM API.
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/user/nihongo-sensei/backend/internal/models"
)

const gradingPromptTemplate = `You are a strict but helpful Japanese language teacher grading a student's sentence.

Grammar point: %s
Pattern: %s
Meaning: %s
Example: %s (%s)

Student's sentence: %s

Evaluate whether the student correctly used the grammar pattern above. Respond ONLY with valid JSON in this exact format:
{
  "correct": true or false,
  "explanation": "Brief explanation of why the sentence is correct or incorrect (in English)",
  "correction": "Corrected version of the sentence if incorrect, otherwise empty string",
  "natural_alternative": "A natural-sounding alternative sentence using the same grammar pattern"
}

Do not include any text outside the JSON object.`

// Client is an Ollama API client.
type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// New creates a new Ollama Client.
func New(baseURL, model string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

// generateRequest is the JSON body sent to /api/generate.
type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// generateResponse is a single line from Ollama's NDJSON stream
// (or the full response when stream=false).
type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// GradeSentence sends the user's sentence to Ollama for grading and returns
// a structured LLMGrade. It uses stream=false for simplicity; set the
// Ollama OLLAMA_REQUEST_TIMEOUT env var server-side if you need longer timeouts.
func (c *Client) GradeSentence(ctx context.Context, gp models.GrammarPoint, userSentence string) (*models.LLMGrade, error) {
	prompt := fmt.Sprintf(gradingPromptTemplate,
		gp.Pattern,
		gp.Pattern,
		gp.Meaning,
		gp.ExampleJP,
		gp.ExampleEN,
		userSentence,
	)

	reqBody := generateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: true, // stream so we can handle long responses gracefully
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(body))
	}

	// Ollama streams NDJSON — each line is a generateResponse chunk.
	var fullResponse strings.Builder
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var chunk generateResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			continue
		}
		if chunk.Error != "" {
			return nil, fmt.Errorf("ollama error: %s", chunk.Error)
		}
		fullResponse.WriteString(chunk.Response)
		if chunk.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading stream: %w", err)
	}

	raw := strings.TrimSpace(fullResponse.String())

	// Extract JSON — the model may wrap it in markdown fences.
	raw = extractJSON(raw)

	var grade models.LLMGrade
	if err := json.Unmarshal([]byte(raw), &grade); err != nil {
		// Return a degraded response rather than a hard error so the UI still works.
		return &models.LLMGrade{
			Correct:     false,
			Explanation: "Could not parse LLM response.",
			RawResponse: raw,
		}, nil
	}
	grade.RawResponse = raw
	return &grade, nil
}

// extractJSON attempts to strip markdown code fences from LLM output.
func extractJSON(s string) string {
	// Strip ```json ... ``` or ``` ... ```
	if idx := strings.Index(s, "```"); idx != -1 {
		s = s[idx:]
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		if end := strings.LastIndex(s, "```"); end != -1 {
			s = s[:end]
		}
	}
	// Find first '{' and last '}'
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start != -1 && end != -1 && end > start {
		return s[start : end+1]
	}
	return s
}
