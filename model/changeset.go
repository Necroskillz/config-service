package model

import (
	"time"
)

type ChangesetState uint

const (
	ChangesetStateOpen ChangesetState = iota
	ChangesetStateReview
	ChangesetStateApplied
	ChangesetStateRejected
	ChangesetStateDiscarded
)

type Changeset struct {
	ID               uint
	CreatedAt        time.Time
	UpdatedAt        time.Time
	User             User
	UserID           uint
	ChangesetChanges []ChangesetChange
	State            ChangesetState
}
