import { createFileRoute } from '@tanstack/react-router';
import { appTitle, seo, versionedTitle } from '~/utils/seo';
import { z } from 'zod';
import {
  getMembershipPermissionsQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions,
  getServicesServiceVersionIdQueryOptions,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense,
  useGetServicesServiceVersionIdSuspense,
} from '~/gen';
import { PermissionEditor } from '~/components/PermissionEditor';
import { SlimPage } from '~/components/SlimPage';
import { Breadcrumbs } from '~/components/Breadcrumbs';
import { PageTitle } from '~/components/PageTitle';

export const Route = createFileRoute('/_auth/(features)/services/$serviceVersionId/features/$featureVersionId/permissions')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
      featureVersionId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions(params.serviceVersionId, params.featureVersionId)
      ),
      context.queryClient.ensureQueryData(
        getMembershipPermissionsQueryOptions({ serviceVersionId: params.serviceVersionId, featureVersionId: params.featureVersionId })
      ),
    ]);
  },
  head: ({ loaderData: [serviceVersion, featureVersion] }) => {
    return {
      meta: [
        ...seo({
          title: appTitle(['Permissions', versionedTitle(featureVersion), versionedTitle(serviceVersion)]),
          description: featureVersion.description,
        }),
      ],
    };
  },
});

function RouteComponent() {
  const { serviceVersionId, featureVersionId } = Route.useParams();
  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);
  const { data: featureVersion } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense(serviceVersionId, featureVersionId);

  return (
    <SlimPage>
      <Breadcrumbs path={[serviceVersion, featureVersion]} />
      <PageTitle>Permissions</PageTitle>
      <PermissionEditor serviceVersionId={serviceVersionId} featureVersionId={featureVersionId} />
    </SlimPage>
  );
}
