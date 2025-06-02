import { useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { TriangleAlert } from 'lucide-react';
import { useAuth } from '~/auth';
import { MutationErrors } from '~/components/MutationErrors';
import { Alert, AlertDescription, AlertTitle } from '~/components/ui/alert';
import { Badge } from '~/components/ui/badge';
import { Button } from '~/components/ui/button';
import {
  ChangesetChangesetChange,
  ChangesetChangesetDto,
  getChangesetsChangesetIdQueryKey,
  usePutChangesetsChangesetIdChangesChangeIdConflictsConfirmDelete,
  usePutChangesetsChangesetIdChangesChangeIdConflictsConfirmUpdate,
  usePutChangesetsChangesetIdChangesChangeIdConflictsCreateToUpdate,
  usePutChangesetsChangesetIdChangesChangeIdConflictsRevalidate,
  usePutChangesetsChangesetIdChangesChangeIdConflictsUpdateToCreate,
} from '~/gen';

function getChangeTypeText(type: ChangesetChangesetChange['type']) {
  switch (type) {
    case 'create':
      return 'Created';
    case 'update':
      return 'Updated';
    case 'delete':
      return 'Deleted';
  }
}

function Diff({ added, removed }: { added?: string; removed?: string }) {
  return (
    <div className="flex">
      {removed && <div className="py-1 bg-diff-remove text-diff-remove-foreground line-through">{removed}</div>}
      {added && <div className="py-1 bg-diff-add text-diff-add-foreground">{added}</div>}
    </div>
  );
}

export function ChangesetChange({
  changeset,
  change,
  onDiscard,
}: {
  changeset: ChangesetChangesetDto;
  change: ChangesetChangesetChange;
  onDiscard: () => void;
}) {
  const { user } = useAuth();

  const canDiscard =
    changeset.state === 'open' &&
    changeset.userId === user.id &&
    !(change.type === 'create' && change.variation != null && Object.keys(change.variation).length === 0);

  return (
    <div className="flex flex-row justify-between items-center gap-2">
      <div className="flex flex-col gap-4">
        <ChangesetChangeDescription change={change} />
        <Conflict changeset={changeset} change={change} />
      </div>
      {canDiscard && (
        <div>
          <Button variant="destructive" size="sm" onClick={() => onDiscard()}>
            Discard
          </Button>
        </div>
      )}
    </div>
  );
}

function ChangesetChangeDescription({ change }: { change: ChangesetChangesetChange }) {
  if (change.newVariationValueId || change.oldVariationValueId) {
    return <ValueChange change={change} />;
  } else if (change.keyId) {
    return <KeyChange change={change} />;
  } else if (change.featureVersionServiceVersionId) {
    return <FeatureVersionServiceVersionChange change={change} />;
  } else if (change.featureVersionId) {
    return <FeatureVersionChange change={change} />;
  } else if (change.serviceVersionId) {
    return <ServiceVersionChange change={change} />;
  }

  return <div>{change.type}</div>;
}

function ValueChange({ change }: { change: ChangesetChangesetChange }) {
  return (
    <div className="flex flex-col gap-2">
      <div>
        {getChangeTypeText(change.type)} <span className="font-semibold">Value</span>
        <span> for </span>
        <Link
          className="link"
          to="/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/values"
          params={{
            serviceVersionId: change.serviceVersionId,
            featureVersionId: change.featureVersionId!,
            keyId: change.keyId!,
          }}
        >
          {change.keyName}
        </Link>
        <span> in </span>
        <Link
          className="link"
          to="/services/$serviceVersionId/features/$featureVersionId"
          params={{
            serviceVersionId: change.serviceVersionId,
            featureVersionId: change.featureVersionId!,
          }}
        >
          {change.featureName} v{change.featureVersion}
        </Link>
      </div>
      {change.variation && Object.keys(change.variation).length > 0 && (
        <div className="flex flex-wrap gap-2 mb-2">
          {Object.values(change.variation).map((value) => (
            <Badge key={value} variant="outline">
              {value}
            </Badge>
          ))}
        </div>
      )}
      <Diff
        added={change.newVariationValueData === '' ? '<empty>' : change.newVariationValueData}
        removed={change.oldVariationValueData === '' ? '<empty>' : change.oldVariationValueData}
      />
    </div>
  );
}

function getChangeTypeClass(type: ChangesetChangesetChange['type']) {
  switch (type) {
    case 'create':
      return 'bg-diff-add text-diff-add-foreground';
    case 'delete':
      return 'bg-diff-remove text-diff-remove-foreground';
  }
}

function KeyChange({ change }: { change: ChangesetChangesetChange }) {
  return (
    <div>
      {getChangeTypeText(change.type)}
      <span className="font-semibold"> Key </span>
      <span className={getChangeTypeClass(change.type)}>
        <Link
          to="/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/values"
          params={{
            serviceVersionId: change.serviceVersionId,
            featureVersionId: change.featureVersionId!,
            keyId: change.keyId!,
          }}
        >
          {change.keyName}
        </Link>
      </span>
      <span> for </span>
      <Link
        className="link"
        to="/services/$serviceVersionId/features/$featureVersionId"
        params={{
          serviceVersionId: change.serviceVersionId,
          featureVersionId: change.featureVersionId!,
        }}
      >
        {change.featureName} v{change.featureVersion}
      </Link>
    </div>
  );
}

function FeatureVersionServiceVersionChange({ change }: { change: ChangesetChangesetChange }) {
  return (
    <div>
      {getChangeTypeText(change.type)}
      <span className="font-semibold"> Service-Feature Link </span>
      <span className={getChangeTypeClass(change.type)}>
        <Link
          to="/services/$serviceVersionId"
          params={{
            serviceVersionId: change.serviceVersionId,
          }}
        >
          {change.serviceName} v{change.serviceVersion}
        </Link>
        <span className="mx-2"> ↔ </span>
        <Link
          to="/services/$serviceVersionId/features/$featureVersionId"
          params={{
            serviceVersionId: change.serviceVersionId,
            featureVersionId: change.featureVersionId!,
          }}
        >
          {change.featureName} v{change.featureVersion}
        </Link>
      </span>
    </div>
  );
}

function FeatureVersionChange({ change }: { change: ChangesetChangesetChange }) {
  return (
    <div>
      {getChangeTypeText(change.type)}
      <span className="font-semibold"> Feature Version </span>
      <span className={getChangeTypeClass(change.type)}>
        <Link
          to="/services/$serviceVersionId/features/$featureVersionId"
          params={{
            serviceVersionId: change.serviceVersionId,
            featureVersionId: change.featureVersionId!,
          }}
        >
          {change.featureName} v{change.featureVersion}
        </Link>
      </span>
    </div>
  );
}

function ServiceVersionChange({ change }: { change: ChangesetChangesetChange }) {
  return (
    <div>
      {getChangeTypeText(change.type)}
      <span className="font-semibold"> Service Version </span>
      <span className={getChangeTypeClass(change.type)}>
        <Link
          to="/services/$serviceVersionId"
          params={{
            serviceVersionId: change.serviceVersionId,
          }}
        >
          {change.serviceName} v{change.serviceVersion}
        </Link>
      </span>
    </div>
  );
}

function Conflict({ changeset, change }: { changeset: ChangesetChangesetDto; change: ChangesetChangesetChange }) {
  if (!change.conflict) {
    return null;
  }

  const queryClient = useQueryClient();

  switch (change.conflict.kind) {
    case 'new_value_duplicate_variation':
      const createToUpdateMutation = usePutChangesetsChangesetIdChangesChangeIdConflictsCreateToUpdate({
        mutation: {
          onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: getChangesetsChangesetIdQueryKey(changeset.id) });
          },
        },
      });

      return (
        <ConflictAlert>
          <div>
            {change.newVariationValueData !== change.conflict.existingValueData
              ? `Value for this variation was already created with value "${change.conflict.existingValueData}". Either make update change instead of create, or discard this change.`
              : `Value for this variation was already created with the same value. This change must be discarded.`}
          </div>
          {changeset.state === 'open' && change.newVariationValueData !== change.conflict.existingValueData && (
            <ConflictActions>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => createToUpdateMutation.mutate({ changeset_id: changeset.id, change_id: change.id })}
              >
                Create &rarr; Update
              </Button>
            </ConflictActions>
          )}
          <MutationErrors mutations={[createToUpdateMutation]} />
        </ConflictAlert>
      );
    case 'old_value_updated':
      if (change.type === 'update') {
        const mutation = usePutChangesetsChangesetIdChangesChangeIdConflictsConfirmUpdate({
          mutation: {
            onSuccess: () => {
              queryClient.invalidateQueries({ queryKey: getChangesetsChangesetIdQueryKey(changeset.id) });
            },
          },
        });

        return (
          <ConflictAlert>
            {change.newVariationValueData !== change.conflict.existingValueData
              ? `Value for this variation was already updated with value "${change.conflict.existingValueData}". Either confirm the update, or discard this change.`
              : `Value for this variation was already updated with the same value. This change must be discarded.`}
            {changeset.state === 'open' && change.newVariationValueData !== change.conflict.existingValueData && (
              <ConflictActions>
                <Button variant="secondary" size="sm" onClick={() => mutation.mutate({ changeset_id: changeset.id, change_id: change.id })}>
                  Confirm update
                </Button>
              </ConflictActions>
            )}
            <MutationErrors mutations={[mutation]} />
          </ConflictAlert>
        );
      } else {
        const mutation = usePutChangesetsChangesetIdChangesChangeIdConflictsConfirmDelete({
          mutation: {
            onSuccess: () => {
              queryClient.invalidateQueries({ queryKey: getChangesetsChangesetIdQueryKey(changeset.id) });
            },
          },
        });

        return (
          <ConflictAlert>
            <div>
              Value for this variation was updated with value "{change.conflict.existingValueData}". Either confirm the deletion, or discard
              this change.
            </div>
            {changeset.state === 'open' && (
              <ConflictActions>
                <Button variant="secondary" size="sm" onClick={() => mutation.mutate({ changeset_id: changeset.id, change_id: change.id })}>
                  Confirm deletion
                </Button>
              </ConflictActions>
            )}
            <MutationErrors mutations={[mutation]} />
          </ConflictAlert>
        );
      }
    case 'old_value_deleted':
      if (change.type === 'update') {
        const mutation = usePutChangesetsChangesetIdChangesChangeIdConflictsUpdateToCreate({
          mutation: {
            onSuccess: () => {
              queryClient.invalidateQueries({ queryKey: getChangesetsChangesetIdQueryKey(changeset.id) });
            },
          },
        });

        return (
          <ConflictAlert>
            <div>Value for this variation was deleted. Either make create change instead of update, or discard this change.</div>
            {changeset.state === 'open' && (
              <ConflictActions>
                <Button variant="secondary" size="sm" onClick={() => mutation.mutate({ changeset_id: changeset.id, change_id: change.id })}>
                  Update &rarr; Create
                </Button>
              </ConflictActions>
            )}
            <MutationErrors mutations={[mutation]} />
          </ConflictAlert>
        );
      } else {
        return (
          <ConflictAlert>
            <div>Value was already deleted. This change must be discarded.</div>
          </ConflictAlert>
        );
      }
    case 'value_in_deleted_key':
      return (
        <ConflictAlert>
          <div>Key of this value was deleted. This change must be discarded.</div>
        </ConflictAlert>
      );
    case 'value_in_deleted_feature':
      return (
        <ConflictAlert>
          <div>Feature of the key this value belongs to was deleted. This change must be discarded.</div>
        </ConflictAlert>
      );
    case 'key_validators_updated':
      const revalidateMutation = usePutChangesetsChangesetIdChangesChangeIdConflictsRevalidate({
        mutation: {
          onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: getChangesetsChangesetIdQueryKey(changeset.id) });
          },
        },
      });

      return (
        <ConflictAlert>
          <div>
            Validators for key of this value were updated. The value must be revalidated using the updated validators. If it's not valid, it
            must be changed or discarded.
          </div>
          {changeset.state === 'open' && (
            <ConflictActions>
              <Button
                variant="secondary"
                size="sm"
                onClick={() => revalidateMutation.mutate({ changeset_id: changeset.id, change_id: change.id })}
              >
                Re-validate value
              </Button>
            </ConflictActions>
          )}
          <MutationErrors mutations={[revalidateMutation]} />
        </ConflictAlert>
      );
    case 'key_in_deleted_feature':
      return (
        <ConflictAlert>
          <div>Feature of this key was deleted. This change must be discarded.</div>
        </ConflictAlert>
      );
    case 'key_duplicate_name':
      return (
        <ConflictAlert>
          <div>Key with this name already exists in the feature. This change must be discarded.</div>
        </ConflictAlert>
      );
    case 'duplicate_link':
      return (
        <ConflictAlert>
          <div>Feature and service are already linked. This change must be discarded.</div>
        </ConflictAlert>
      );
    case 'deleted_link':
      return (
        <ConflictAlert>
          <div>Feature and service link was already deleted. This change must be discarded.</div>
        </ConflictAlert>
      );
    case 'inconsistent_feature_version':
      return (
        <ConflictAlert>
          <div>Feature version with this version number was already created. This change must be discarded.</div>
        </ConflictAlert>
      );
    case 'inconsistent_service_version':
      return (
        <ConflictAlert>
          <div>Service version with this version number was already created. This change must be discarded.</div>
        </ConflictAlert>
      );
    case 'change_in_published_service_version':
      switch (change.kind) {
        case 'service_version':
          return (
            <ConflictAlert>
              <div>Service version can't be deleted after is has been published. This change must be discarded.</div>
            </ConflictAlert>
          );
        case 'feature_version_service_version':
          return (
            <ConflictAlert>
              <div>
                Feature version can't be unlinked from a service version after is has been published. This change must be discarded.
              </div>
            </ConflictAlert>
          );
        case 'key':
          return (
            <ConflictAlert>
              <div>
                Key can't be deleted from a feature version that is linked to a published service version. This change must be discarded.
              </div>
            </ConflictAlert>
          );
      }
  }
}

function ConflictAlert({ children }: { children: React.ReactNode }) {
  return (
    <Alert variant="destructive">
      <TriangleAlert />
      <AlertTitle>Conflict Detected</AlertTitle>
      <AlertDescription>
        <div className="flex flex-col gap-2">{children}</div>
      </AlertDescription>
    </Alert>
  );
}

function ConflictActions({ children }: { children: React.ReactNode }) {
  return <div className="flex flex-row gap-2">{children}</div>;
}
