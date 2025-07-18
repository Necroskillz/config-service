import { createFileRoute, Link, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import {
  getServicesServiceVersionIdQueryOptions,
  getServicesServiceVersionIdFeaturesQueryOptions,
  getServicesServiceVersionIdVersionsQueryOptions,
  usePutServicesServiceVersionIdPublish,
  usePostServicesServiceVersionIdVersions,
  getServicesServiceVersionIdQueryKey,
} from '~/gen';
import { SlimPage } from '~/components/SlimPage';
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '~/components/ui/dropdown-menu';
import { List, ListItem } from '~/components/List';
import { PageTitle } from '~/components/PageTitle';
import { buttonVariants } from '~/components/ui/button';
import { DropdownMenuTriggerLabel } from '~/components/ui/dropdown-menu';
import { useQueryClient } from '@tanstack/react-query';
import { Badge } from '~/components/ui/badge';
import { ChevronDownIcon } from 'lucide-react';
import { useGetServicesServiceVersionIdSuspense } from '~/gen/hooks/useGetServicesServiceVersionIdSuspense';
import { useGetServicesServiceVersionIdFeaturesSuspense } from '~/gen/hooks/useGetServicesServiceVersionIdFeaturesSuspense';
import { useGetServicesServiceVersionIdVersions } from '~/gen/hooks/useGetServicesServiceVersionIdVersions';
import { seo, appTitle, versionedTitle } from '~/utils/seo';
import { MutationErrors } from '~/components/MutationErrors';
import { useChangeset } from '~/hooks/use-changeset';
import ServiceAdministrators from './-components/ServiceAdministrators';
import { DotDotDot } from '~/components/DotDotDot';

export const Route = createFileRoute('/_auth/(services)/services/$serviceVersionId/')({
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
  const { refresh } = useChangeset();
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
        refresh();
        navigate({ to: '/services/$serviceVersionId', params: { serviceVersionId: newId } });
      },
    },
  });

  const publishMutation = usePutServicesServiceVersionIdPublish({
    mutation: {
      onSuccess: () => {
        queryClient.refetchQueries({ queryKey: getServicesServiceVersionIdQueryKey(serviceVersionId) });
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
            <DotDotDot>
              <Link to="/services/$serviceVersionId/edit" params={{ serviceVersionId }}>
                <DropdownMenuItem>Edit</DropdownMenuItem>
              </Link>
              <Link to="/services/$serviceVersionId/permissions" params={{ serviceVersionId }}>
                <DropdownMenuItem>Permissions</DropdownMenuItem>
              </Link>
              <Link to="/services/$serviceVersionId/link" params={{ serviceVersionId }}>
                <DropdownMenuItem>Link/Unlink features</DropdownMenuItem>
              </Link>
              <Link to="/change-history" search={{ serviceId: serviceVersion.serviceId, serviceVersionId }}>
                <DropdownMenuItem>History</DropdownMenuItem>
              </Link>
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
            </DotDotDot>
          )}
        </div>
      </div>

      <div className="flex flex-col gap-4">
        <MutationErrors mutations={[publishMutation, createNewVersionMutation]} />
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
                  <DropdownMenuItem key={sv.id}>
                    <span className="text-accent-foreground font-bold">v{sv.version}</span>
                  </DropdownMenuItem>
                ) : (
                  <Link key={sv.id} to="/services/$serviceVersionId" params={{ serviceVersionId: sv.id }}>
                    <DropdownMenuItem>
                      <span>v{sv.version}</span>
                    </DropdownMenuItem>
                  </Link>
                )
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
        <div className="text-muted-foreground">{serviceVersion.description}</div>
        <p className="text-sm">
          <ServiceAdministrators admins={serviceVersion.admins} />
        </p>
        <List>
          {features.length ? (
            features.map((featureVersion) => (
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
            ))
          ) : (
            <ListItem>No features</ListItem>
          )}
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
