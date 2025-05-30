import { Link } from '@tanstack/react-router';
import { useAuth } from '~/auth';
import { Badge } from '~/components/ui/badge';
import { Button } from '~/components/ui/button';
import { ChangesetChangesetChange, ChangesetChangesetDto } from '~/gen';

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
    <div className="flex flex-row justify-between items-center">
      <div>
        <ChangesetChangeDescription change={change} />
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
