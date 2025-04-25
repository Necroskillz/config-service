import { DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger, DropdownMenuTriggerLabel } from '~/components/ui/dropdown-menu';
import { PageTitle } from '~/components/PageTitle';
import { Badge } from '~/components/ui/badge';
import { DropdownMenu } from '~/components/ui/dropdown-menu';
import {
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspense,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense,
  useGetServicesServiceVersionIdSuspense,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdVersions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdVersionsQueryOptions,
} from '~/gen';
import { ChevronDownIcon } from 'lucide-react';
import { List, ListItem } from '~/components/List';
import { buttonVariants } from '~/components/ui/button';
import { useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';

export function FeatureDetail({ serviceVersionId, featureVersionId }: { serviceVersionId: number; featureVersionId: number }) {
  const queryClient = useQueryClient();
  //const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);
  const { data: featureVersion } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense(serviceVersionId, featureVersionId);
  const { data: keys } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspense(serviceVersionId, featureVersionId);
  const { data: allFeatureVersions } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdVersions(serviceVersionId, featureVersionId, {
    query: {
      enabled: false,
      gcTime: 0,
    },
  });

  function fetchOtherVersions() {
    queryClient.fetchQuery({
      ...getServicesServiceVersionIdFeaturesFeatureVersionIdVersionsQueryOptions(serviceVersionId, featureVersionId),
      staleTime: Infinity,
      gcTime: 0,
    });
  }
  return (
    <>
      <PageTitle>
        {featureVersion.name}
        <Badge className="ml-2">v{featureVersion.version}</Badge>
      </PageTitle>
      <div className="flex flex-col gap-4">
        <div className="flex flex-row gap-2">
          <span>Version</span>
          <DropdownMenu>
            <DropdownMenuTrigger onMouseOver={() => fetchOtherVersions()} onTouchStart={() => fetchOtherVersions()} asChild>
              <DropdownMenuTriggerLabel className="flex flex-row gap-1 items-center">
                <span className="text-accent-foreground">v{featureVersion.version}</span>
                <ChevronDownIcon className="size-4 opacity-50 text-muted-foreground" />
              </DropdownMenuTriggerLabel>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              {allFeatureVersions?.map((fv) => (
                <DropdownMenuItem key={fv.id}>
                  {fv.id === featureVersionId ? (
                    <span>v{fv.version}</span>
                  ) : (
                    <Link to="/services/$serviceVersionId/features/$featureVersionId" params={{ serviceVersionId, featureVersionId: fv.id }}>
                      v{fv.version}
                    </Link>
                  )}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <div className="text-muted-foreground">{featureVersion.description}</div>
        <List>
          {keys.map((key) => (
            <ListItem key={key.id}>
              <h2 className="text-lg font-bold">
                <Link
                  to="/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/values"
                  params={{ serviceVersionId, featureVersionId, keyId: key.id }}
                >
                  {key.name}
                </Link>
              </h2>
            </ListItem>
          ))}
        </List>
        {featureVersion.canEdit && (
          <div>
            <Link
              className={buttonVariants({ variant: 'default', size: 'sm' })}
              to="/services/$serviceVersionId/features/$featureVersionId/keys/create"
              params={{ serviceVersionId, featureVersionId }}
            >
              Create Key
            </Link>
          </div>
        )}
      </div>
    </>
  );
}
