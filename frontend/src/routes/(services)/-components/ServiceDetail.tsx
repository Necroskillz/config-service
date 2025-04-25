import { List } from '~/components/List';
import { ListItem } from '~/components/List';
import { Badge } from '~/components/ui/badge';
import { Link } from '@tanstack/react-router';
import {
  useGetServicesServiceVersionIdSuspense,
  useGetServicesServiceVersionIdFeaturesSuspense,
  useGetServicesServiceVersionIdVersions,
  getServicesServiceVersionIdVersionsQueryOptions,
} from '~/gen';
import { PageTitle } from '~/components/PageTitle';
import { buttonVariants } from '~/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTriggerLabel,
} from '~/components/ui/dropdown-menu';
import { ChevronDownIcon } from 'lucide-react';
import { useQueryClient } from '@tanstack/react-query';
export function ServiceDetail({ serviceVersionId }: { serviceVersionId: number }) {
  const queryClient = useQueryClient();
  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);
  const { data: features } = useGetServicesServiceVersionIdFeaturesSuspense(serviceVersionId);
  const { data: allServiceVersions } = useGetServicesServiceVersionIdVersions(serviceVersionId, {
    query: {
      enabled: false,
      gcTime: 0,
    },
  });

  function fetchOtherVersions() {
    queryClient.fetchQuery({
      ...getServicesServiceVersionIdVersionsQueryOptions(serviceVersionId),
      staleTime: Infinity,
      gcTime: 0,
    });
  }

  return (
    <>
      <PageTitle>
        {serviceVersion.name}
        <Badge className="ml-2">v{serviceVersion.version}</Badge>
      </PageTitle>
      <div className="flex flex-col gap-4">
        <div className="flex flex-row gap-2">
          <span>Version</span>
          <DropdownMenu>
            <DropdownMenuTrigger onMouseOver={() => fetchOtherVersions()} onTouchStart={() => fetchOtherVersions()} asChild>
              <DropdownMenuTriggerLabel className="flex flex-row gap-1 items-center">
                <span className="text-accent-foreground">v{serviceVersion.version}</span>
                <ChevronDownIcon className="size-4 opacity-50 text-muted-foreground" />
              </DropdownMenuTriggerLabel>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              {allServiceVersions?.map((sv) => (
                <DropdownMenuItem key={sv.id}>
                  {sv.id === serviceVersionId ? (
                    <span>v{sv.version}</span>
                  ) : (
                    <Link to="/services/$serviceVersionId" params={{ serviceVersionId: sv.id }}>
                      v{sv.version}
                    </Link>
                  )}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <div className="text-muted-foreground">{serviceVersion.description}</div>
        <List>
          {features.map((featureVersion) => (
            <ListItem key={featureVersion.id}>
              <h2 className="text-lg font-bold">
                <Link
                  to="/services/$serviceVersionId/features/$featureVersionId"
                  params={{ serviceVersionId: serviceVersionId, featureVersionId: featureVersion.id }}
                >
                  {featureVersion.name}
                </Link>
                <Badge className="ml-2">v{featureVersion.version}</Badge>
              </h2>
            </ListItem>
          ))}
        </List>
        {serviceVersion.canEdit && (
          <div>
            <Link
              className={buttonVariants({ variant: 'default', size: 'sm' })}
              to="/services/$serviceVersionId/features/create"
              params={{ serviceVersionId }}
            >
              Create Feature
            </Link>
          </div>
        )}
      </div>
    </>
  );
}
