import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { Breadcrumbs } from '~/components/Breadcrumbs';
import { PageTitle } from '~/components/PageTitle';
import { PermissionEditor } from '~/components/PermissionEditor';
import { SlimPage } from '~/components/SlimPage';
import {
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions,
  getServicesServiceVersionIdQueryOptions,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdSuspense,
  useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense,
  useGetServicesServiceVersionIdSuspense,
} from '~/gen';
import { appTitle, seo, versionedTitle } from '~/utils/seo';

export const Route = createFileRoute('/_auth/(keys)/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/permissions')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
      featureVersionId: z.coerce.number(),
      keyId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions(params.serviceVersionId, params.featureVersionId)
      ),
      context.queryClient.ensureQueryData(
        getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdQueryOptions(
          params.serviceVersionId,
          params.featureVersionId,
          params.keyId
        )
      ),
    ]);
  },
  head: ({ loaderData: [serviceVersion, featureVersion, key] }) => {
    return {
      meta: [...seo({ title: appTitle(['Permissions', key.name, versionedTitle(featureVersion), versionedTitle(serviceVersion)]) })],
      description: key.description,
    };
  },
});

function RouteComponent() {
  const { serviceVersionId, featureVersionId, keyId } = Route.useParams();
  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);
  const { data: featureVersion } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdSuspense(serviceVersionId, featureVersionId);
  const { data: key } = useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdSuspense(serviceVersionId, featureVersionId, keyId);

  return (
    <SlimPage>
      <Breadcrumbs path={[serviceVersion, featureVersion, key]} />
      <PageTitle>Permissions</PageTitle>
      <PermissionEditor
        serviceVersionId={serviceVersionId}
        featureVersionId={featureVersionId}
        keyId={keyId}
        serviceTypeId={serviceVersion.serviceTypeId}
      />
    </SlimPage>
  );
}
