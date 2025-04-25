import { createFileRoute } from '@tanstack/react-router'
import { ServicesRouteComponent } from './(services)/services.index'
import { getServicesQueryOptions } from '~/gen';

export const Route = createFileRoute('/')({
  component: ServicesRouteComponent,
  loader: async ({ context }) => {
    context.queryClient.prefetchQuery(getServicesQueryOptions());
  },
})
