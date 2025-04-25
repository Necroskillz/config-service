import { createFileRoute } from '@tanstack/react-router';
import { ServiceDetail } from './-components/ServiceDetail';
import { z } from 'zod';
import { Suspense } from 'react';
import { getServicesServiceVersionIdQueryOptions, getServicesServiceVersionIdFeaturesQueryOptions } from '~/gen';
import { SlimPage } from '~/components/SlimPage';

const ParamsSchema = z.object({
  serviceVersionId: z.coerce.number(),
});

export const Route = createFileRoute('/(services)/services/$serviceVersionId')({
  component: RouteComponent,
  params: {
    parse: ParamsSchema.parse,
  },
  loader: async ({ context, params }) => {
    context.queryClient.prefetchQuery(getServicesServiceVersionIdQueryOptions(params.serviceVersionId));
    context.queryClient.prefetchQuery(getServicesServiceVersionIdFeaturesQueryOptions(params.serviceVersionId));
  },
});

function RouteComponent() {
  const { serviceVersionId } = Route.useParams();

  return (
    <SlimPage>
      <Suspense fallback={<div>Loading...</div>}>
        <ServiceDetail serviceVersionId={serviceVersionId} />
      </Suspense>
    </SlimPage>
  );
}
