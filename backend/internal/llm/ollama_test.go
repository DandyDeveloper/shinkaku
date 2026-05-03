package llm

import "testing"

func TestParseConversationPrompt_UnquotedAssistantMessageWithComma(t *testing.T) {
	raw := `{}"scenario": "At a coffee shop, a customer is choosing a drink.", "assistant_message": 様, コーヒーにしましょうか? 明日(あした)は忙しいです.}`

	parsed, ok := parseConversationPrompt(raw)
	if !ok || parsed == nil {
		t.Fatalf("expected parser to recover JSON-like response")
	}

	if parsed.Scenario != "At a coffee shop, a customer is choosing a drink." {
		t.Fatalf("unexpected scenario: %q", parsed.Scenario)
	}

	wantAssistant := "様, コーヒーにしましょうか? 明日(あした)は忙しいです."
	if parsed.AssistantMessage != wantAssistant {
		t.Fatalf("unexpected assistant message: got %q want %q", parsed.AssistantMessage, wantAssistant)
	}
}
