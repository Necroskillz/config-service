import { Button } from '~/components/ui/button';
import { MembershipPermissionDto } from '~/gen';
import { VariationBadges } from '~/components/VariationBadges';
import { Link } from '@tanstack/react-router';

export function Permission({
  permission,
  onDelete,
  disabled,
  readOnly = false,
}: {
  permission: MembershipPermissionDto;
  onDelete?: () => void;
  disabled?: boolean;
  readOnly?: boolean;
}) {
  const inherited = permission.groupId !== null;

  switch (permission.kind) {
    case 'service':
      return (
        <PermissionItem onDelete={onDelete} disabled={disabled} showRemove={!inherited && !readOnly}>
          Permission <strong>{permission.permission}</strong> for service <strong>{permission.serviceName}</strong>{' '}
          {inherited && <InheritedPermission permission={permission} />}
        </PermissionItem>
      );
    case 'feature':
      return (
        <PermissionItem onDelete={onDelete} disabled={disabled} showRemove={!inherited && !readOnly}>
          Permission <strong>{permission.permission}</strong> for feature <strong>{permission.featureName}</strong> in service{' '}
          <strong>{permission.serviceName}</strong>{' '}
          {inherited && <InheritedPermission permission={permission} />}
        </PermissionItem>
      );
    case 'key':
      return (
        <PermissionItem onDelete={onDelete} disabled={disabled} showRemove={!inherited && !readOnly}>
          Permission <strong>{permission.permission}</strong> for key <strong>{permission.keyName}</strong> in feature{' '}
          <strong>{permission.featureName}</strong> in service <strong>{permission.serviceName}</strong>{' '}
          {inherited && <InheritedPermission permission={permission} />}
        </PermissionItem>
      );
    case 'variation':
      return (
        <PermissionItem onDelete={onDelete} disabled={disabled} showRemove={!inherited && !readOnly}>
          Permission <strong>{permission.permission}</strong> for variation <VariationBadges variation={permission.variation!} /> in key{' '}
          <strong>{permission.keyName}</strong> in feature <strong>{permission.featureName}</strong> in service{' '}
          <strong>{permission.serviceName}</strong>{' '}
          {inherited && <InheritedPermission permission={permission} />}
        </PermissionItem>
      );
  }
}

function InheritedPermission({ permission }: { permission: MembershipPermissionDto }) {
  return (
    <span className="text-muted-foreground">
      (inherited from{' '}
      <Link to="/admin/membership/groups/$groupId" params={{ groupId: permission.groupId! }} className="link">
        {permission.groupName}
      </Link>
      )
    </span>
  );
}

function PermissionItem({
  children,
  onDelete,
  showRemove,
  disabled,
}: {
  children: React.ReactNode;
  onDelete?: () => void;
  showRemove: boolean;
  disabled?: boolean;
}) {
  return (
    <div className="inline-flex flex-row items-center gap-2 justify-between">
      <div>{children}</div>
      {showRemove && (
        <Button variant="destructive" size="sm" onClick={onDelete} disabled={disabled}>
          Remove
        </Button>
      )}
    </div>
  );
}
