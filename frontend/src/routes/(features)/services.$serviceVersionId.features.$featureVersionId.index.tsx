import { createFileRoute, Link } from '@tanstack/react-router';
import { z } from 'zod';
import { SlimPage } from '~/components/SlimPage';
import {
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdVersionsQueryOptions,
  getServicesServiceVersionIdQueryOptions,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspense,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdVersions,
} from '~/gen';
import { useQueryClient } from '@tanstack/react-query';
import { DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem } from '~/components/ui/dropdown-menu';
import { Badge } from '~/components/ui/badge';
import { ChevronDownIcon, EllipsisIcon } from 'lucide-react';
import { List, ListItem } from '~/components/List';
import { PageTitle } from '~/components/PageTitle';
import { Button, buttonVariants } from '~/components/ui/button';
import { DropdownMenuTriggerLabel } from '~/components/ui/dropdown-menu';
import { seo } from '~/utils/seo';
import { appTitle } from '~/utils/seo';
import { versionedTitle } from '~/utils/seo';

const ParamsSchema = z.object({
  serviceVersionId: z.coerce.number(),
  featureVersionId: z.coerce.number(),
});

export const Route = createFileRoute('/(features)/services/$serviceVersionId/features/$featureVersionId/')({
  component: RouteComponent,
  params: {
    parse: ParamsSchema.parse,
  },
  loader: async ({ context, params }) => {
    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions(params.serviceVersionId, params.featureVersionId)
      ),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdKeysQueryOptions(params.serviceVersionId, params.featureVersionId)
      ),
    ]);
  },
  head: ({ loaderData: [serviceVersion, featureVersion] }) => {
    return {
      meta: [
        ...seo({
          title: appTitle([versionedTitle(featureVersion), versionedTitle(serviceVersion)]),
          description: featureVersion.description,
        }),
      ],
    };
  },
});

function RouteComponent() {
  const { serviceVersionId, featureVersionId } = Route.useParams();
  const queryClient = useQueryClient();
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
    <SlimPage>
      <PageTitle>
        {featureVersion.name}
        <Badge className="ml-2">v{featureVersion.version}</Badge>
      </PageTitle>
      <div className="flex flex-col gap-4">
        <div className="flex flex-row gap-2 items-center">
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
                    <Link
                      to="/services/$serviceVersionId/features/$featureVersionId"
                      params={{ serviceVersionId, featureVersionId: fv.id }}
                    >
                      v{fv.version}
                    </Link>
                  )}
                </DropdownMenuItem>
              ))}
            </DropdownMenuContent>
          </DropdownMenu>
          <div className="flex items-center w-full justify-end">
            {featureVersion.canEdit && (
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" size="icon">
                    <EllipsisIcon className="size-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent>
                  <DropdownMenuItem>
                    <Link className="w-full" to="/services/$serviceVersionId/features/$featureVersionId/edit" params={{ serviceVersionId, featureVersionId }}>
                      Edit
                    </Link>
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            )}
          </div>
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
    </SlimPage>
  );
}
