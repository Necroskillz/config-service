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
	ServiceVersion                 *ServiceVersion
	ServiceVersionID               *uint
	FeatureVersionServiceVersion   *FeatureVersionServiceVersion
	FeatureVersionServiceVersionID *uint
	Key                            *Key
	KeyID                          *uint
	VariationValue                 *VariationValue
	VariationValueID               *uint
	OldValue                       *string
	NewValue                       *string
}
