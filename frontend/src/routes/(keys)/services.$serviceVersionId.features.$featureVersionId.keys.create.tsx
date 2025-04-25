import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { KeyForm } from './-components/KeyForm';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import { useGetServicesServiceVersionIdFeaturesFeatureVersionId } from '~/gen';
import { Skeleton } from '~/components/ui/skeleton';

const Schema = z.object({
  serviceVersionId: z.coerce.number(),
  featureVersionId: z.coerce.number(),
});

export const Route = createFileRoute('/(keys)/services/$serviceVersionId/features/$featureVersionId/keys/create')({
  component: RouteComponent,
  params: {
    parse: Schema.parse,
  },
});

function RouteComponent() {
  const { serviceVersionId, featureVersionId } = Route.useParams();
  const { data: featureVersion, isLoading } = useGetServicesServiceVersionIdFeaturesFeatureVersionId(serviceVersionId, featureVersionId);

  return (
    <SlimPage>
      <PageTitle>Create Key</PageTitle>
      <div className="text-muted-foreground mb-4">
        {isLoading ? (
          <Skeleton className="w-[350px] h-6" />
        ) : (
          <p>
            Create a new key for {featureVersion?.name} v{featureVersion?.version}
          </p>
        )}
      </div>
      <KeyForm serviceVersionId={serviceVersionId} featureVersionId={featureVersionId} />
    </SlimPage>
  );
}
