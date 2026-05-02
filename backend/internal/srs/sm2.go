// Package srs implements the SM-2 spaced repetition algorithm.
// Reference: https://www.supermemo.com/en/archives1990-2015/english/ol/sm2
package srs

import (
	"math"
	"time"
)

const (
	minEFactor = 1.3
	initEFactor = 2.5
)

// CardState holds the mutable SRS fields for a review card.
type CardState struct {
	Interval    int     // days until next review
	Repetitions int     // number of consecutive correct reviews (grade >= 3)
	EFactor     float64 // ease factor
}

// NewCard returns a CardState with default SM-2 initial values.
func NewCard() CardState {
	return CardState{
		Interval:    1,
		Repetitions: 0,
		EFactor:     initEFactor,
	}
}

// Review applies the SM-2 algorithm to a card given a quality grade (0–5).
// It returns the updated CardState and the next due date.
//
// Grade semantics:
//
//	5 – perfect response
//	4 – correct response after a hesitation
//	3 – correct response recalled with serious difficulty
//	2 – incorrect response; where the correct one seemed easy to recall
//	1 – incorrect response; the correct one remembered
//	0 – complete blackout
func Review(state CardState, grade int, now time.Time) (CardState, time.Time) {
	if grade < 0 {
		grade = 0
	}
	if grade > 5 {
		grade = 5
	}

	// Update ease factor regardless of pass/fail.
	// EF' = EF + (0.1 - (5-q) * (0.08 + (5-q) * 0.02))
	q := float64(grade)
	newEF := state.EFactor + (0.1 - (5-q)*(0.08+(5-q)*0.02))
	newEF = math.Max(minEFactor, newEF)

	var newInterval int
	var newRepetitions int

	if grade >= 3 {
		// Correct response — advance the schedule.
		switch state.Repetitions {
		case 0:
			newInterval = 1
		case 1:
			newInterval = 6
		default:
			newInterval = int(math.Round(float64(state.Interval) * newEF))
		}
		newRepetitions = state.Repetitions + 1
	} else {
		// Incorrect response — reset to beginning.
		newInterval = 1
		newRepetitions = 0
	}

	nextDue := now.AddDate(0, 0, newInterval)

	return CardState{
		Interval:    newInterval,
		Repetitions: newRepetitions,
		EFactor:     newEF,
	}, nextDue
}
