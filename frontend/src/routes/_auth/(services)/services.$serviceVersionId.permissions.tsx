import { createFileRoute } from '@tanstack/react-router';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import {
  getMembershipPermissionsQueryOptions,
  getServicesServiceVersionIdQueryOptions,
  useGetServicesServiceVersionIdSuspense,
} from '~/gen';
import { z } from 'zod';
import { versionedTitle, seo, appTitle } from '~/utils/seo';
import { Breadcrumbs } from '~/components/Breadcrumbs';
import { PermissionEditor } from '~/components/PermissionEditor';

export const Route = createFileRoute('/_auth/(services)/services/$serviceVersionId/permissions')({
  component: RouteComponent,
  params: {
    parse: z.object({
      serviceVersionId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ context, params }) => {
    return Promise.all([
      context.queryClient.ensureQueryData(getServicesServiceVersionIdQueryOptions(params.serviceVersionId)),
      context.queryClient.ensureQueryData(getMembershipPermissionsQueryOptions({ serviceVersionId: params.serviceVersionId })),
    ]);
  },
  head: ({ loaderData: [serviceVersion] }) => {
    return {
      meta: [
        ...seo({
          title: appTitle(['Permissions', versionedTitle(serviceVersion)]),
          description: serviceVersion.description,
        }),
      ],
    };
  },
});

function RouteComponent() {
  const { serviceVersionId } = Route.useParams();
  const { data: serviceVersion } = useGetServicesServiceVersionIdSuspense(serviceVersionId);

  return (
    <SlimPage>
      <Breadcrumbs path={[serviceVersion]} />
      <PageTitle>Permissions</PageTitle>
      <PermissionEditor serviceVersionId={serviceVersionId} />
    </SlimPage>
  );
}
