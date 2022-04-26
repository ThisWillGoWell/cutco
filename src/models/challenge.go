package models

import (
	"time"
)

const ChallengePrefix = "challenge#"
const InfoPrefix = "info#"
const ChallengeCompletionPrefix = "completion#"

type ID string

type Challenge struct {
	ID          ID
	Name        string
	Description string
	PointReward int
	VersionID   string
}

type ChallengeCompletion struct {
	ChallengeID ID
	CompletedBy ID
	CompletedAt time.Time
	ApprovedBy  ID
	PhotoID     ID
}

// Version
// enforce version on updates
func (c Challenge) Version() string {
	return c.VersionID
}

func TimeToString(t time.Time) string {
	return t.Format(time.RFC3339)
}

// Keys Load ChallengeInfo by ID
// PrimaryKey challenge#ID / info#
func (c Challenge) Keys() (string, string) {
	return string(ChallengePrefix + c.ID), InfoPrefix
}

// Keys Load all completions for a challengeID by time
// PrimaryKey: challenge#ChallengeID
// SortKey: completion#CompletedAtCompletedBy
func (c ChallengeCompletion) Keys() (string, string) {
	return string(ChallengePrefix + c.ChallengeID), ChallengeCompletionPrefix + TimeToString(c.CompletedAt) + string(c.CompletedBy)
}

// GSI0 Load all completions for a userID by time
// PrimaryKey: completion#CompletedBy
// SortKey: challenge#CompletedAtChallengeID
func (c ChallengeCompletion) GSI0() (string, string) {
	return string(ChallengeCompletionPrefix + c.CompletedBy), ChallengePrefix + TimeToString(c.CompletedAt) + string(c.ChallengeID)
}
