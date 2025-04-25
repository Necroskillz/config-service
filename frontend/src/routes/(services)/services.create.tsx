import { createFileRoute } from '@tanstack/react-router';
import { SlimPage } from '~/components/SlimPage';
import { ServiceForm } from './-components/ServiceForm';
import { PageTitle } from '~/components/PageTitle';
import { getServiceTypesQueryOptions } from '~/gen';

export const Route = createFileRoute('/(services)/services/create')({
  component: RouteComponent,
  loader: async ({ context }) => {
    context.queryClient.prefetchQuery(getServiceTypesQueryOptions());
  },
});

function RouteComponent() {
  return (
    <SlimPage>
      <PageTitle>Create Service</PageTitle>
      <ServiceForm />
    </SlimPage>
  );
}
