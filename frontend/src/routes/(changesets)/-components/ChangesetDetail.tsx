import { Link } from '@tanstack/react-router';
import { VariantProps } from 'class-variance-authority';
import { List, ListItem } from '~/components/List';
import { Badge, badgeVariants } from '~/components/ui/badge';
import { DbChangesetStateEnum, ServiceChangesetChange, useGetChangesetsChangesetIdSuspense } from '~/gen';
import { ChangesetActions } from './ChangesetActions';

export function ChangesetDetail({ changesetId }: { changesetId: number }) {
  const { data: changeset } = useGetChangesetsChangesetIdSuspense(changesetId);

  function getStateBadgeVariant(state: DbChangesetStateEnum): VariantProps<typeof badgeVariants>['variant'] {
    switch (state) {
      case 'open':
        return 'default';
      case 'discarded':
        return 'destructive';
      case 'applied':
        return 'secondary';
      case 'stashed':
        return 'outline';
      case 'committed':
        return 'outline';
      default:
        return 'default';
    }
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex gap-2">
        <span className="font-semibold">State:</span>
        <Badge variant={getStateBadgeVariant(changeset.state)}>{changeset.state}</Badge>
      </div>
      {changeset.changes.length > 0 ? (
        <List>
          {changeset.changes.map((change) => (
            <ListItem key={change.id}>
              <ChangesetChange change={change} />
            </ListItem>
          ))}
        </List>
      ) : (
        <div className="text-muted-foreground">
          {changeset.state === 'discarded' ? 'Changeset has been discarded' : 'Changeset contains no changes'}
        </div>
      )}
      <ChangesetActions changeset={changeset} />
    </div>
  );
}

function getChangeTypeText(type: ServiceChangesetChange['type']) {
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

function ChangesetChange({ change }: { change: ServiceChangesetChange }) {
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

function ValueChange({ change }: { change: ServiceChangesetChange }) {
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

function getChangeTypeClass(type: ServiceChangesetChange['type']) {
  switch (type) {
    case 'create':
      return 'bg-diff-add text-diff-add-foreground';
    case 'delete':
      return 'bg-diff-remove text-diff-remove-foreground';
  }
}

function KeyChange({ change }: { change: ServiceChangesetChange }) {
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

function FeatureVersionServiceVersionChange({ change }: { change: ServiceChangesetChange }) {
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

function FeatureVersionChange({ change }: { change: ServiceChangesetChange }) {
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

function ServiceVersionChange({ change }: { change: ServiceChangesetChange }) {
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
