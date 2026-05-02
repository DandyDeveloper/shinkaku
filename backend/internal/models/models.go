package models

import "time"

// GrammarPoint represents a single Japanese grammar pattern.
type GrammarPoint struct {
	ID         int64     `json:"id"`
	JLPTLevel  string    `json:"jlpt_level"` // e.g. "N5", "N4", "N3", "N2", "N1"
	Pattern    string    `json:"pattern"`    // e.g. "〜てもいい"
	Meaning    string    `json:"meaning"`    // English meaning
	ExampleJP  string    `json:"example_jp"` // Japanese example sentence
	ExampleEN  string    `json:"example_en"` // English translation
	Notes      string    `json:"notes"`
	Source     string    `json:"source"`    // e.g. "hanabira", "manual"
	CreatedAt  time.Time `json:"created_at"`
}

// ReviewCard is the SRS card associated with a grammar point.
type ReviewCard struct {
	ID             int64      `json:"id"`
	GrammarPointID int64      `json:"grammar_point_id"`
	Interval       int        `json:"interval"`     // days until next review
	Repetitions    int        `json:"repetitions"`  // consecutive correct reviews
	EFactor        float64    `json:"e_factor"`     // ease factor, min 1.3
	DueDate        time.Time  `json:"due_date"`
	LastReviewed   *time.Time `json:"last_reviewed,omitempty"`
}

// ReviewHistory records each review event for a card.
type ReviewHistory struct {
	ID          int64     `json:"id"`
	CardID      int64     `json:"card_id"`
	Grade       int       `json:"grade"` // 0–5
	ReviewedAt  time.Time `json:"reviewed_at"`
	LLMFeedback string    `json:"llm_feedback,omitempty"`
}

// Session represents an active review session (lightweight, in-memory).
type Session struct {
	Cards     []ReviewCard `json:"cards"`
	Total     int          `json:"total"`
	Remaining int          `json:"remaining"`
}

// LLMGrade is the structured response from Ollama grading.
type LLMGrade struct {
	Correct         bool   `json:"correct"`
	Explanation     string `json:"explanation"`
	Correction      string `json:"correction,omitempty"`
	NaturalAlt      string `json:"natural_alternative,omitempty"`
	RawResponse     string `json:"raw_response,omitempty"`
}

// ChallengeRequest is the payload sent by the frontend for LLM grading.
type ChallengeRequest struct {
	GrammarPointID int64  `json:"grammar_point_id"`
	UserSentence   string `json:"user_sentence"`
}

// GradeRequest is the payload for submitting an SRS card grade.
type GradeRequest struct {
	Grade int `json:"grade"` // 0–5
}
