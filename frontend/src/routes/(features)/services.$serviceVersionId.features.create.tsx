import { createFileRoute, Link } from '@tanstack/react-router';
import { SlimPage } from '~/components/SlimPage';
import { FeatureForm } from './-components/FeatureForm';
import { PageTitle } from '~/components/PageTitle';
import { z } from 'zod';
import { getServicesServiceVersionIdQueryOptions, useGetServicesServiceVersionId } from '~/gen';
import { Skeleton } from '~/components/ui/skeleton';

const Schema = z.object({
  serviceVersionId: z.coerce.number(),
});

export const Route = createFileRoute('/(features)/services/$serviceVersionId/features/create')({
  component: RouteComponent,
  params: {
    parse: Schema.parse,
  },
  loader: async ({ context, params }) => {
    context.queryClient.prefetchQuery(getServicesServiceVersionIdQueryOptions(params.serviceVersionId));
  },
});

function RouteComponent() {
  const { serviceVersionId } = Route.useParams();
  const { data: serviceVersion, isLoading } = useGetServicesServiceVersionId(serviceVersionId);

  return (
    <SlimPage>
      <PageTitle>Create Feature</PageTitle>

      <div className="text-muted-foreground mb-4">
        {isLoading ? (
          <Skeleton className="w-[350px] h-6" />
        ) : (
          <p>
            Created feature will be linked to{' '}
            <Link className="text-accent-foreground" to="/services/$serviceVersionId" params={{ serviceVersionId }}>
              {serviceVersion?.name} v{serviceVersion?.version}
            </Link>
          </p>
        )}
      </div>

      <FeatureForm serviceVersionId={serviceVersionId} />
    </SlimPage>
  );
}
