// Package llm provides a client for Ollama's local LLM API.
package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/user/shinkaku/backend/internal/models"
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

const conversationPromptTemplate = `You are a Japanese conversation coach preparing a short roleplay for a learner.

Grammar point: %s
Meaning: %s
Example: %s (%s)
Notes: %s

Create a very short realistic conversation setup where the learner should naturally reply using the target grammar point.
Respond ONLY with valid JSON in this exact format:
{
  "scenario": "A one-sentence English description of the situation",
  "assistant_message": "A short Japanese message from the conversation partner"
}

Keep the assistant message short, natural, and clearly answerable in one or two sentences.
%s
Do not include any text outside the JSON object.`

const conversationPromptRepairTemplate = `Your previous response did not follow the required furigana format.

Rewrite ONLY valid JSON with the same schema:
{
	"scenario": "A one-sentence English description of the situation",
	"assistant_message": "A short Japanese message from the conversation partner"
}

Furigana requirement:
- For Japanese text with kanji, annotate as 漢字(かんじ)
- Do not use HTML ruby tags
- If a sentence has no kanji, plain kana is acceptable

Previous response:
%s`

const conversationGradingPromptTemplate = `You are a strict but helpful Japanese teacher grading a learner's reply in a roleplay.

Target grammar point: %s
Meaning: %s
Reference example: %s (%s)
Scenario: %s
Partner's message: %s
Learner's reply: %s

Evaluate whether the learner gave a natural reply and correctly used the target grammar point.
Respond ONLY with valid JSON in this exact format:
{
  "correct": true or false,
  "explanation": "Brief explanation in English",
  "correction": "Corrected Japanese reply if needed, otherwise empty string",
  "natural_alternative": "A natural alternative Japanese reply using the same grammar point",
  "assistant_reply": "A short natural Japanese follow-up from the partner"
}

%s
Do not include any text outside the JSON object.`

const conversationGradingRepairTemplate = `Your previous response did not follow the required furigana format.

Rewrite ONLY valid JSON with the exact same schema:
{
	"correct": true or false,
	"explanation": "Brief explanation in English",
	"correction": "Corrected Japanese reply if needed, otherwise empty string",
	"natural_alternative": "A natural alternative Japanese reply using the same grammar point",
	"assistant_reply": "A short natural Japanese follow-up from the partner"
}

Furigana requirement for Japanese fields (correction, natural_alternative, assistant_reply):
- For Japanese text with kanji, annotate as 漢字(かんじ)
- Do not use HTML ruby tags
- If a field has no kanji, plain kana is acceptable

Previous response:
%s`

var furiganaTokenPattern = regexp.MustCompile(`[一-龯々〆ヵヶ]+\([ぁ-ゖー]+\)`)
var kanjiPattern = regexp.MustCompile(`[一-龯々〆ヵヶ]`)

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

	raw, err := c.generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var grade models.LLMGrade
	if err := json.Unmarshal([]byte(raw), &grade); err != nil {
		return &models.LLMGrade{
			Correct:     false,
			Explanation: "Could not parse LLM response.",
			RawResponse: raw,
		}, nil
	}
	grade.RawResponse = raw
	return &grade, nil
}

// GenerateConversationPrompt asks Ollama for a short scenario and opening message.
func (c *Client) GenerateConversationPrompt(ctx context.Context, gp models.GrammarPoint, includeFurigana bool) (*models.ConversationPrompt, error) {
	prompt := fmt.Sprintf(conversationPromptTemplate,
		gp.Pattern,
		gp.Meaning,
		gp.ExampleJP,
		gp.ExampleEN,
		gp.Notes,
		furiganaInstruction(includeFurigana),
	)

	raw, err := c.generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	log.Printf("raw conversation prompt response: %s", raw)

	conversationPrompt, parsed := parseConversationPrompt(raw)
	if !parsed {
		return &models.ConversationPrompt{
			Scenario:         "Could not parse the generated scenario. Try again.",
			AssistantMessage: raw,
		}, nil
	}

	if includeFurigana && missingConversationPromptFurigana(*conversationPrompt) {
		retryRaw, retryErr := c.generate(ctx, fmt.Sprintf(conversationPromptRepairTemplate, raw))
		if retryErr == nil {
			retried, ok := parseConversationPrompt(retryRaw)
			if ok {
				conversationPrompt = retried
			}
		}
	}
	return conversationPrompt, nil
}

// GradeConversationReply asks Ollama to evaluate a reply in a grammar-focused conversation.
func (c *Client) GradeConversationReply(ctx context.Context, gp models.GrammarPoint, scenario, assistantMessage, userReply string, includeFurigana bool) (*models.ConversationGrade, error) {
	prompt := fmt.Sprintf(conversationGradingPromptTemplate,
		gp.Pattern,
		gp.Meaning,
		gp.ExampleJP,
		gp.ExampleEN,
		scenario,
		assistantMessage,
		userReply,
		furiganaInstruction(includeFurigana),
	)

	raw, err := c.generate(ctx, prompt)
	if err != nil {
		return nil, err
	}

	var grade models.ConversationGrade
	if err := json.Unmarshal([]byte(raw), &grade); err != nil {
		return &models.ConversationGrade{
			Correct:     false,
			Explanation: "Could not parse LLM response.",
			RawResponse: raw,
		}, nil
	}

	if includeFurigana && missingConversationGradeFurigana(grade) {
		retryRaw, retryErr := c.generate(ctx, fmt.Sprintf(conversationGradingRepairTemplate, raw))
		if retryErr == nil {
			var retried models.ConversationGrade
			if err := json.Unmarshal([]byte(retryRaw), &retried); err == nil {
				grade = retried
				raw = retryRaw
			}
		}
	}
	grade.RawResponse = raw
	return &grade, nil
}

func (c *Client) generate(ctx context.Context, prompt string) (string, error) {
	reqBody := generateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: true, // stream so we can handle long responses gracefully
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(body))
	}

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
			return "", fmt.Errorf("ollama error: %s", chunk.Error)
		}
		fullResponse.WriteString(chunk.Response)
		if chunk.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("reading stream: %w", err)
	}

	raw := strings.TrimSpace(fullResponse.String())
	return extractJSON(raw), nil
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

func furiganaInstruction(includeFurigana bool) string {
	if includeFurigana {
		return "Furigana is REQUIRED for Japanese text with kanji. Use plain text format 漢字(かんじ), for example: 今日(きょう)は学校(がっこう)へ行(い)きます. Do not use HTML ruby tags."
	}
	return "Do not add furigana annotations; output plain natural Japanese."
}

func needsFurigana(text string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	return kanjiPattern.MatchString(text)
}

func hasFurigana(text string) bool {
	return furiganaTokenPattern.MatchString(text)
}

func missingConversationPromptFurigana(prompt models.ConversationPrompt) bool {
	if needsFurigana(prompt.AssistantMessage) && !hasFurigana(prompt.AssistantMessage) {
		return true
	}
	return false
}

func parseConversationPrompt(raw string) (*models.ConversationPrompt, bool) {
	var parsed models.ConversationPrompt
	if err := json.Unmarshal([]byte(raw), &parsed); err == nil {
		parsed.Scenario = strings.TrimSpace(parsed.Scenario)
		parsed.AssistantMessage = strings.TrimSpace(parsed.AssistantMessage)
		if parsed.Scenario != "" && parsed.AssistantMessage != "" {
			return &parsed, true
		}
	}

	// Some models return almost-JSON with unquoted Japanese string values.
	// Recover those fields directly before giving up on parsing.
	scenario := extractJSONLikeStringField(raw, "scenario", "situation", "context")
	assistantMessage := extractJSONLikeStringField(raw,
		"assistant_message",
		"assistantMessage",
		"message",
		"partner_message",
		"conversation_starter",
	)
	if scenario != "" && assistantMessage != "" {
		return &models.ConversationPrompt{
			Scenario:         scenario,
			AssistantMessage: assistantMessage,
		}, true
	}

	var loose map[string]any
	if err := json.Unmarshal([]byte(raw), &loose); err != nil {
		return nil, false
	}

	scenario = firstString(loose,
		"scenario",
		"situation",
		"context",
	)
	assistantMessage = firstString(loose,
		"assistant_message",
		"assistantMessage",
		"message",
		"partner_message",
		"conversation_starter",
	)

	scenario = strings.TrimSpace(scenario)
	assistantMessage = strings.TrimSpace(assistantMessage)
	if scenario == "" || assistantMessage == "" {
		return nil, false
	}

	return &models.ConversationPrompt{
		Scenario:         scenario,
		AssistantMessage: assistantMessage,
	}, true
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func extractJSONLikeStringField(raw string, keys ...string) string {
	for _, key := range keys {
		quoted := regexp.MustCompile(`"` + regexp.QuoteMeta(key) + `"\s*:\s*"((?:\\.|[^"\\])*)"`)
		if match := quoted.FindStringSubmatch(raw); len(match) == 2 {
			if unescaped, err := strconv.Unquote("\"" + match[1] + "\""); err == nil {
				value := strings.TrimSpace(unescaped)
				if value != "" {
					return value
				}
			}
			value := strings.TrimSpace(match[1])
			if value != "" {
				return value
			}
		}

		// Accept unquoted values that may contain commas and stop at next JSON key or closing brace.
		bare := regexp.MustCompile(`(?s)"` + regexp.QuoteMeta(key) + `"\s*:\s*(.*?)(?:,\s*"[A-Za-z0-9_]+"\s*:|\s*})`)
		if match := bare.FindStringSubmatch(raw); len(match) == 2 {
			value := strings.TrimSpace(strings.Trim(match[1], `"'`))
			if value != "" {
				return value
			}
		}
	}
	return ""
}

func missingConversationGradeFurigana(grade models.ConversationGrade) bool {
	fields := []string{grade.Correction, grade.NaturalAlternative, grade.AssistantReply}
	for _, field := range fields {
		if needsFurigana(field) && !hasFurigana(field) {
			return true
		}
	}
	return false
}
