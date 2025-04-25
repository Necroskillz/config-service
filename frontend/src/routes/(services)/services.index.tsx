import { createFileRoute, Link } from '@tanstack/react-router';
import { useAuth } from '~/auth';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import { buttonVariants } from '~/components/ui/button';
import { getServicesQueryOptions } from '~/gen';
import { ServiceList } from './-components/ServiceList';
import { Suspense } from 'react';

export const Route = createFileRoute('/(services)/services/')({
  component: ServicesRouteComponent,
  loader: async ({ context }) => {
    context.queryClient.prefetchQuery(getServicesQueryOptions());
  },
});

export function ServicesRouteComponent() {
  const { user } = useAuth();

  return (
    <SlimPage>
      <PageTitle>Services</PageTitle>
      <Suspense fallback={<div>Loading...</div>}>
        <ServiceList />
      </Suspense>

      {user.isGlobalAdmin && (
        <div className="mt-8">
          <Link className={buttonVariants({ variant: 'default', size: 'sm' })} to="/services/create">
            Create New Service
          </Link>
        </div>
      )}
    </SlimPage>
  );
}
