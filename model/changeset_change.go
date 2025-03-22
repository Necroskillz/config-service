package model

import (
	"time"
)

type ChangesetChangeType uint

const (
	ChangesetChangeTypeCreate ChangesetChangeType = iota
	ChangesetChangeTypeUpdate
	ChangesetChangeTypeDelete
)

type ChangesetChange struct {
	ID                             uint
	CreatedAt                      time.Time
	Changeset                      Changeset
	ChangesetID                    uint
	Type                           ChangesetChangeType
	FeatureVersion                 *FeatureVersion
	FeatureVersionID               *uint
	PreviousFeatureVersion         *FeatureVersion
	PreviousFeatureVersionID       *uint
	ServiceVersion                 *ServiceVersion
	ServiceVersionID               *uint
	PreviousServiceVersion         *ServiceVersion
	PreviousServiceVersionID       *uint
	FeatureVersionServiceVersion   *FeatureVersionServiceVersion
	FeatureVersionServiceVersionID *uint
	Key                            *Key
	KeyID                          *uint
	NewVariationValue              *VariationValue
	NewVariationValueID            *uint
	OldVariationValue              *VariationValue
	OldVariationValueID            *uint
}
