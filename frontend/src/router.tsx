import { createRouter as createTanStackRouter } from '@tanstack/react-router';
import { routerWithQueryClient } from '@tanstack/react-router-with-query';
import { routeTree } from './routeTree.gen';
import { DefaultCatchBoundary } from './components/DefaultCatchBoundary';
import { NotFound } from './components/NotFound';
import { QueryClient } from '@tanstack/react-query';
import { AnonymousUser } from './auth';
import { Pending } from './components/Pending';
export function createRouter() {
  const queryClient = new QueryClient();

  return routerWithQueryClient(
    createTanStackRouter({
      routeTree,
      context: { queryClient, accessToken: null, user: AnonymousUser },
      defaultPreload: 'intent',
      defaultPendingMs: 200,
      defaultPreloadStaleTime: 0,
      scrollRestoration: true,
      defaultErrorComponent: DefaultCatchBoundary,
      defaultNotFoundComponent: () => <NotFound />,
      defaultPendingComponent: () => <Pending />,
    }),
    queryClient
  );
}

declare module '@tanstack/react-router' {
  interface Register {
    router: ReturnType<typeof createRouter>;
  }
}
