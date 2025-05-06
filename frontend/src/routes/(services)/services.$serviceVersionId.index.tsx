import { createFileRoute, Link, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import {
  getServicesServiceVersionIdQueryOptions,
  getServicesServiceVersionIdFeaturesQueryOptions,
  getServicesServiceVersionIdVersionsQueryOptions,
  usePutServicesServiceVersionIdPublish,
  usePostServicesServiceVersionIdVersions,
} from '~/gen';
import { SlimPage } from '~/components/SlimPage';
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '~/components/ui/dropdown-menu';
import { List, ListItem } from '~/components/List';
import { PageTitle } from '~/components/PageTitle';
import { Button, buttonVariants } from '~/components/ui/button';
import { DropdownMenuTriggerLabel } from '~/components/ui/dropdown-menu';
import { useQueryClient } from '@tanstack/react-query';
import { Badge } from '~/components/ui/badge';
import { ChevronDownIcon, EllipsisIcon } from 'lucide-react';
import { useGetServicesServiceVersionIdSuspense } from '~/gen/hooks/useGetServicesServiceVersionIdSuspense';
import { useGetServicesServiceVersionIdFeaturesSuspense } from '~/gen/hooks/useGetServicesServiceVersionIdFeaturesSuspense';
import { useGetServicesServiceVersionIdVersions } from '~/gen/hooks/useGetServicesServiceVersionIdVersions';
import { seo, appTitle, versionedTitle } from '~/utils/seo';
import { MutationErrors } from '~/components/MutationErrors';

export const Route = createFileRoute('/(services)/services/$serviceVersionId/')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(getServicesServiceVersionIdFeaturesQueryOptions(params.serviceVersionId)),
    ]);
  },
  head: ({ loaderData: [serviceVersion] }) => {
    return {
      meta: [
        ...seo({
          title: appTitle([versionedTitle(serviceVersion)]),
          description: serviceVersion.description,
        }),
      ],
    };
  },
});

function RouteComponent() {
  const { serviceVersionId } = Route.useParams();

  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);
  const { data: features } = useGetServicesServiceVersionIdFeaturesSuspense(serviceVersionId);
  const { data: allServiceVersions } = useGetServicesServiceVersionIdVersions(serviceVersionId, {
    query: {
      enabled: false,
      gcTime: 0,
    },
  });

  const createNewVersionMutation = usePostServicesServiceVersionIdVersions({
    mutation: {
      onSuccess: ({ newId }) => {
        navigate({ to: '/services/$serviceVersionId', params: { serviceVersionId: newId } });
      },
    },
  });

  const publishMutation = usePutServicesServiceVersionIdPublish({
    mutation: {
      onSuccess: () => {
        queryClient.refetchQueries({ queryKey: getServicesServiceVersionIdQueryOptions(serviceVersionId).queryKey });
      },
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
    <SlimPage>
      <div className="flex items-center justify-between mb-8">
        <PageTitle className="mb-0">
          {serviceVersion.name}
          <Badge variant={serviceVersion.published ? 'default' : 'outline'} className="ml-2">
            v{serviceVersion.version} ({serviceVersion.published ? 'published' : 'draft'})
          </Badge>
        </PageTitle>
        <div className="flex items-center">
          {serviceVersion.canEdit && (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="icon">
                  <EllipsisIcon className="size-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                <DropdownMenuItem onClick={() => navigate({ to: '/services/$serviceVersionId/edit', params: { serviceVersionId } })}>
                  Edit
                </DropdownMenuItem>
                <DropdownMenuItem onClick={() => navigate({ to: '/services/$serviceVersionId/link', params: { serviceVersionId } })}>
                  Link/Unlink features
                </DropdownMenuItem>
                {!serviceVersion.published && (
                  <DropdownMenuItem onClick={() => publishMutation.mutate({ service_version_id: serviceVersionId })}>
                    Publish
                  </DropdownMenuItem>
                )}
                {serviceVersion.isLastVersion && (
                  <DropdownMenuItem onClick={() => createNewVersionMutation.mutate({ service_version_id: serviceVersionId })}>
                    Create new version
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </div>

      <div className="flex flex-col gap-4">
        <MutationErrors mutations={[publishMutation]} />
        <div className="flex flex-row gap-2 items-center">
          <span>Version</span>
          <DropdownMenu>
            <DropdownMenuTrigger onMouseOver={() => fetchOtherVersions()} onTouchStart={() => fetchOtherVersions()} asChild>
              <DropdownMenuTriggerLabel className="flex flex-row gap-1 items-center">
                <span className="text-accent-foreground">v{serviceVersion.version}</span>
                <ChevronDownIcon className="size-4 opacity-50 text-muted-foreground" />
              </DropdownMenuTriggerLabel>
            </DropdownMenuTrigger>
            <DropdownMenuContent>
              {allServiceVersions?.map((sv) =>
                sv.id === serviceVersionId ? (
                  <DropdownMenuItem>
                    <span className="text-accent-foreground font-bold">v{sv.version}</span>
                  </DropdownMenuItem>
                ) : (
                  <DropdownMenuItem
                    key={sv.id}
                    onClick={() => navigate({ to: '/services/$serviceVersionId', params: { serviceVersionId: sv.id } })}
                  >
                    <span>v{sv.version}</span>
                  </DropdownMenuItem>
                )
              )}
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
    </SlimPage>
  );
}
