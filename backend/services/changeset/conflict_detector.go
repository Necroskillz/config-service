package changeset

import (
	"github.com/necroskillz/config-service/db"
)

type ConflictKind string

const (
	ConflictKindNewValueDuplicateVariation      ConflictKind = "new_value_duplicate_variation"
	ConflictKindOldValueDeleted                 ConflictKind = "old_value_deleted"
	ConflictKindOldValueUpdated                 ConflictKind = "old_value_updated"
	ConflictKindValueInDeletedKey               ConflictKind = "value_in_deleted_key"
	ConflictKindValueInDeletedFeature           ConflictKind = "value_in_deleted_feature"
	ConflictKindKeyValidatorsUpdated            ConflictKind = "key_validators_updated"
	ConflictKindKeyInDeletedFeature             ConflictKind = "key_in_deleted_feature"
	ConflictKindKeyDuplicateName                ConflictKind = "key_duplicate_name"
	ConflictKindDuplicateLink                   ConflictKind = "duplicate_link"
	ConflictKindDeletedLink                     ConflictKind = "deleted_link"
	ConflictKindInconsistentFeatureVersion      ConflictKind = "inconsistent_feature_version"
	ConflictKindInconsistentServiceVersion      ConflictKind = "inconsistent_service_version"
	ConflictKindChangeInPublishedServiceVersion ConflictKind = "change_in_published_service_version"
)

type Conflict struct {
	Kind              ConflictKind `json:"kind" validate:"required"`
	ExistingValueData *string      `json:"existingValueData,omitempty"`
}

type LinkKey struct {
	FeatureID uint
	ServiceID uint
}

type ConflictCheckerContext struct {
	Change              db.GetChangesetChangesRow
	DeletedLinks        map[LinkKey]bool
	DeletedKeys         map[string]bool
	LastFeatureVersions map[uint]int
	LastServiceVersions map[uint]int
}

type ConflictCheckerFunc func(ctx ConflictCheckerContext) ConflictKind

type ConflictDetector struct {
	checkers []ConflictCheckerFunc
}

func NewConflictDetector() *ConflictDetector {
	detector := &ConflictDetector{
		checkers: []ConflictCheckerFunc{
			valueInDeletedFeatureChecker,
			valueInDeletedKeyChecker,
			keyValidatorsUpdatedChecker,
			oldValueUpdatedChecker,
			oldValueDeletedChecker,
			newValueDuplicateVariationChecker,
			changeInPublishedServiceVersionChecker,
			keyInDeletedFeatureChecker,
			keyDuplicateNameChecker,
			duplicateLinkChecker,
			deletedLinkChecker,
			inconsistentFeatureVersionChecker,
			inconsistentServiceVersionChecker,
		},
	}

	return detector
}

func (c *ConflictDetector) DetectConflicts(raw []db.GetChangesetChangesRow, changes []ChangesetChange) int {
	deletedLinks := map[LinkKey]bool{}
	deletedKeys := map[string]bool{}
	lastFeatureVersions := map[uint]int{}
	lastServiceVersions := map[uint]int{}
	count := 0

	for i, change := range raw {
		ctx := ConflictCheckerContext{
			Change:              change,
			DeletedLinks:        deletedLinks,
			DeletedKeys:         deletedKeys,
			LastFeatureVersions: lastFeatureVersions,
			LastServiceVersions: lastServiceVersions,
		}

		conflict := c.detect(ctx)
		if conflict != nil {
			changes[i].Conflict = conflict
			count++
		}

		if change.Kind == db.ChangesetChangeKindFeatureVersionServiceVersion && change.Type == db.ChangesetChangeTypeDelete {
			deletedLinks[LinkKey{FeatureID: *change.FeatureID, ServiceID: change.ServiceID}] = true
		} else if change.Kind == db.ChangesetChangeKindKey && change.Type == db.ChangesetChangeTypeDelete {
			deletedKeys[*change.KeyName] = true
		} else if change.Kind == db.ChangesetChangeKindFeatureVersion && change.Type == db.ChangesetChangeTypeCreate {
			ctx.LastFeatureVersions[*change.FeatureID] = *change.FeatureVersion
		} else if change.Kind == db.ChangesetChangeKindServiceVersion && change.Type == db.ChangesetChangeTypeCreate {
			ctx.LastServiceVersions[change.ServiceID] = change.ServiceVersion
		}
	}

	return count
}

func (c *ConflictDetector) detect(ctx ConflictCheckerContext) *Conflict {
	for _, checker := range c.checkers {
		ck := checker(ctx)
		if ck != "" {
			return &Conflict{Kind: ck, ExistingValueData: ctx.Change.ExistingValueData}
		}
	}

	return nil
}

func valueInDeletedFeatureChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindVariationValue && ctx.Change.FeatureVersionValidTo != nil {
		return ConflictKindValueInDeletedFeature
	}

	return ""
}

func valueInDeletedKeyChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindVariationValue && ctx.Change.KeyValidTo != nil {
		return ConflictKindValueInDeletedKey
	}

	return ""
}

func oldValueUpdatedChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindVariationValue && ctx.Change.OldVariationValueValidTo != nil && ctx.Change.ExistingVariationContextID != nil {
		return ConflictKindOldValueUpdated
	}

	return ""
}

func oldValueDeletedChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindVariationValue && ctx.Change.OldVariationValueValidTo != nil && ctx.Change.ExistingVariationContextID == nil {
		return ConflictKindOldValueDeleted
	}

	return ""
}

func newValueDuplicateVariationChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindVariationValue && ctx.Change.Type == db.ChangesetChangeTypeCreate && ctx.Change.ExistingVariationContextID != nil {
		return ConflictKindNewValueDuplicateVariation
	}

	return ""
}

func keyValidatorsUpdatedChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindVariationValue && ctx.Change.KeyValidatorsUpdatedAt.After(ctx.Change.CreatedAt) {
		return ConflictKindKeyValidatorsUpdated
	}

	return ""
}

func keyInDeletedFeatureChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindKey && ctx.Change.FeatureVersionValidTo != nil {
		return ConflictKindKeyInDeletedFeature
	}

	return ""
}

func keyDuplicateNameChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindKey && ctx.Change.Type == db.ChangesetChangeTypeCreate && !ctx.DeletedKeys[*ctx.Change.KeyName] && ctx.Change.ExistingKeyID != nil {
		return ConflictKindKeyDuplicateName
	}

	return ""
}

func duplicateLinkChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindFeatureVersionServiceVersion &&
		ctx.Change.Type == db.ChangesetChangeTypeCreate &&
		!ctx.DeletedLinks[LinkKey{FeatureID: *ctx.Change.FeatureID, ServiceID: ctx.Change.ServiceID}] &&
		ctx.Change.ExistingFeatureVersionServiceVersionID != nil {
		return ConflictKindDuplicateLink
	}

	return ""
}

func deletedLinkChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindFeatureVersionServiceVersion && ctx.Change.Type == db.ChangesetChangeTypeDelete && ctx.Change.FeatureVersionServiceVersionValidTo != nil {
		return ConflictKindDeletedLink
	}

	return ""
}

func inconsistentFeatureVersionChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindFeatureVersion && *ctx.Change.FeatureVersion != max(ctx.LastFeatureVersions[*ctx.Change.FeatureID], ctx.Change.LastFeatureVersionVersion)+1 {
		return ConflictKindInconsistentFeatureVersion
	}

	return ""
}

func inconsistentServiceVersionChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind == db.ChangesetChangeKindServiceVersion && ctx.Change.ServiceVersion != max(ctx.LastServiceVersions[ctx.Change.ServiceID], ctx.Change.LastServiceVersionVersion)+1 {
		return ConflictKindInconsistentServiceVersion
	}

	return ""
}

func changeInPublishedServiceVersionChecker(ctx ConflictCheckerContext) ConflictKind {
	if ctx.Change.Kind != db.ChangesetChangeKindVariationValue && ctx.Change.Type == db.ChangesetChangeTypeDelete && ctx.Change.ServiceVersionPublished {
		return ConflictKindChangeInPublishedServiceVersion
	}

	return ""
}
