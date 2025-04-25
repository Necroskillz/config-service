import { createRouter as createTanStackRouter } from '@tanstack/react-router';
import { routerWithQueryClient } from '@tanstack/react-router-with-query';
import { routeTree } from './routeTree.gen';
import { DefaultCatchBoundary } from './components/DefaultCatchBoundary';
import { NotFound } from './components/NotFound';
import { QueryClient } from '@tanstack/react-query';
import { AnonymousUser } from './auth';

export function createRouter() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 1000 * 5,
      },
    },
  });

  return routerWithQueryClient(
    createTanStackRouter({
      routeTree,
      context: { queryClient, accessToken: null, user: AnonymousUser },
      defaultPreload: 'intent',
      defaultErrorComponent: DefaultCatchBoundary,
      defaultNotFoundComponent: () => <NotFound />,
    }),
    queryClient
  );
}

declare module '@tanstack/react-router' {
  interface Register {
    router: ReturnType<typeof createRouter>;
  }
}
