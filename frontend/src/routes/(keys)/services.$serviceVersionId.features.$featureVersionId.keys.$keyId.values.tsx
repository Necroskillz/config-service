import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { Suspense } from 'react';
import { ValueMatrix } from './-components/ValueMatrix';
import {
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdQueryOptions,
  getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryOptions,
  getServicesServiceVersionIdQueryOptions,
  getServiceTypesServiceTypeIdVariationPropertiesQueryOptions,
} from '~/gen';
const Schema = z.object({
  serviceVersionId: z.coerce.number(),
  featureVersionId: z.coerce.number(),
  keyId: z.coerce.number(),
});

export const Route = createFileRoute('/(keys)/services/$serviceVersionId/features/$featureVersionId/keys/$keyId/values')({
  component: RouteComponent,
  params: {
    parse: Schema.parse,
  },
  loader: async ({ context, params }) => {
    context.queryClient.prefetchQuery(getServicesServiceVersionIdQueryOptions(params.serviceVersionId));
    context.queryClient.prefetchQuery(getServiceTypesServiceTypeIdVariationPropertiesQueryOptions(params.serviceVersionId));
    context.queryClient.prefetchQuery(
      getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryOptions(
        params.serviceVersionId,
        params.featureVersionId,
        params.keyId
      )
    );
    context.queryClient.prefetchQuery(
      getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdQueryOptions(
        params.serviceVersionId,
        params.featureVersionId,
        params.keyId
      )
    );
  },
});

function RouteComponent() {
  const { serviceVersionId, featureVersionId, keyId } = Route.useParams();

  return (
    <div className="p-4">
      <Suspense fallback={<div>Loading...</div>}>
        <ValueMatrix serviceVersionId={serviceVersionId} featureVersionId={featureVersionId} keyId={keyId} />
      </Suspense>
    </div>
  );
}
