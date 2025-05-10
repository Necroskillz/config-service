import { ChevronRight } from 'lucide-react';
import { Link, LinkComponentProps } from '@tanstack/react-router';
import { ServiceFeatureVersionDto, ServiceKeyDto, ServiceServiceVersionDto } from '~/gen';
import { cn } from '~/lib/utils';
export function Breadcrumbs({
  path,
  className,
}: {
  path: [ServiceServiceVersionDto, ServiceFeatureVersionDto?, ServiceKeyDto?];
  className?: string;
}) {
  const [serviceVersion, featureVersion, key] = path;

  if (key && !featureVersion) {
    throw new Error('Key was specified without a feature version');
  }

  return (
    <div className={cn('flex gap-1 mb-2', className)}>
      <BreadcrumbLink to="/services/$serviceVersionId" params={{ serviceVersionId: serviceVersion.id }}>
        {serviceVersion.name} v{serviceVersion.version}
      </BreadcrumbLink>
      {featureVersion && (
        <>
          <BreadcrumbLink
            to="/services/$serviceVersionId/features/$featureVersionId"
            params={{ serviceVersionId: serviceVersion.id, featureVersionId: featureVersion.id }}
          >
            {featureVersion.name} v{featureVersion.version}
          </BreadcrumbLink>
        </>
      )}
      {key && (
        <>
          <BreadcrumbLink
            to="/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/values"
            params={{ serviceVersionId: serviceVersion.id, featureVersionId: featureVersion!.id, keyId: key.id }}
          >
            {key.name}
          </BreadcrumbLink>
        </>
      )}
    </div>
  );
}

function BreadcrumbLink(props: LinkComponentProps<'a'>) {
  return (
    <>
      <Link className="link" {...props} />
      <div className="flex items-end">
        <ChevronRight className="w-4 h-5 text-muted-foreground" />
      </div>
    </>
  );
}
