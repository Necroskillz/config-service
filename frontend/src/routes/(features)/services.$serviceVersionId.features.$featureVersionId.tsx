import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { SlimPage } from '~/components/SlimPage';
import { FeatureDetail } from './-components/FeatureDetail';
import {
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions,
  getServicesServiceVersionIdQueryOptions,
} from '~/gen';
import { Suspense } from 'react';

const ParamsSchema = z.object({
  serviceVersionId: z.coerce.number(),
  featureVersionId: z.coerce.number(),
});

export const Route = createFileRoute('/(features)/services/$serviceVersionId/features/$featureVersionId')({
  component: RouteComponent,
  params: {
    parse: ParamsSchema.parse,
  },
  loader: async ({ context, params }) => {
    context.queryClient.prefetchQuery(getServicesServiceVersionIdQueryOptions(params.serviceVersionId));
    context.queryClient.prefetchQuery(
      getServicesServiceVersionIdFeaturesFeatureVersionIdQueryOptions(params.serviceVersionId, params.featureVersionId)
    );
    context.queryClient.prefetchQuery(
      getServicesServiceVersionIdFeaturesFeatureVersionIdKeysQueryOptions(params.serviceVersionId, params.featureVersionId)
    );
  },
});

function RouteComponent() {
  const { serviceVersionId, featureVersionId } = Route.useParams();
  return (
    <SlimPage>
      <Suspense fallback={<div>Loading...</div>}>
        <FeatureDetail serviceVersionId={serviceVersionId} featureVersionId={featureVersionId} />
      </Suspense>
    </SlimPage>
  );
}
