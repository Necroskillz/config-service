package changeset

import (
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
)

type ChangesetChange struct {
	ID                             uint                   `json:"id" validate:"required"`
	Type                           db.ChangesetChangeType `json:"type" validate:"required"`
	ServiceVersionID               uint                   `json:"serviceVersionId" validate:"required"`
	ServiceName                    string                 `json:"serviceName" validate:"required"`
	ServiceVersion                 int                    `json:"serviceVersion" validate:"required"`
	ServiceID                      uint                   `json:"serviceId"`
	PreviousServiceVersionID       *uint                  `json:"previousServiceVersionId"`
	FeatureVersionID               *uint                  `json:"featureVersionId"`
	FeatureName                    *string                `json:"featureName"`
	FeatureVersion                 *int                   `json:"featureVersion"`
	FeatureID                      *uint                  `json:"featureId"`
	PreviousFeatureVersionID       *uint                  `json:"previousFeatureVersionId"`
	FeatureVersionServiceVersionID *uint                  `json:"featureVersionServiceVersionId"`
	KeyID                          *uint                  `json:"keyId"`
	KeyName                        *string                `json:"keyName"`
	NewVariationValueID            *uint                  `json:"newVariationValueId"`
	NewVariationValueData          *string                `json:"newVariationValueData"`
	OldVariationValueID            *uint                  `json:"oldVariationValueId"`
	OldVariationValueData          *string                `json:"oldVariationValueData"`
	Variation                      map[uint]string        `json:"variation"`
}

type Changeset struct {
	ID       uint
	UserID   uint
	UserName string
	State    db.ChangesetState
}

type ChangesetWithChanges struct {
	Changeset
	ChangesetChanges []ChangesetChange
}

func (c ChangesetWithChanges) CanBeAppliedBy(user *auth.User) bool {
	if !c.IsOpen() && !c.IsCommitted() {
		return false
	}

	if c.IsOpen() && !c.BelongsTo(user.ID) {
		return false
	}

	if user.IsGlobalAdmin {
		return true
	}

	for _, change := range c.ChangesetChanges {
		if user.GetPermissionForService(change.ServiceVersionID) != constants.PermissionAdmin {
			return false
		}
	}

	return true
}

func (c Changeset) BelongsTo(userID uint) bool {
	return c.UserID == userID
}

func (c Changeset) IsOpen() bool {
	return c.State == db.ChangesetStateOpen
}

func (c Changeset) IsCommitted() bool {
	return c.State == db.ChangesetStateCommitted
}

func (c Changeset) IsDiscarded() bool {
	return c.State == db.ChangesetStateDiscarded
}

func (c Changeset) IsStashed() bool {
	return c.State == db.ChangesetStateStashed
}

func (c ChangesetWithChanges) IsEmpty() bool {
	return len(c.ChangesetChanges) == 0
}
