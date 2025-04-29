import { HeadContent, Outlet, Scripts, createRootRouteWithContext } from '@tanstack/react-router';
import { TanStackRouterDevtools } from '@tanstack/react-router-devtools';
import * as React from 'react';
import { DefaultCatchBoundary } from '~/components/DefaultCatchBoundary';
import { NotFound } from '~/components/NotFound';
import { Header } from '~/components/Header';
import appCss from '~/styles/app.css?url';
import { seo } from '~/utils/seo';
import { isServer, QueryClient } from '@tanstack/react-query';
import { AuthProvider, getAccessToken, getRefreshToken, refreshFn, setRequestAccessToken } from '~/auth';
import { useState } from 'react';
import { ChangesetProvider } from '~/hooks/useChangeset';

export const Route = createRootRouteWithContext<{
  queryClient: QueryClient;
  accessToken: string | null;
}>()({
  head: () => ({
    meta: [
      {
        charSet: 'utf-8',
      },
      {
        name: 'viewport',
        content: 'width=device-width, initial-scale=1',
      },
      ...seo({
        title: 'Config Service',
        description: `Config Service is a tool for managing configuration for your application.`,
      }),
    ],
    links: [
      { rel: 'stylesheet', href: appCss },
      {
        rel: 'apple-touch-icon',
        sizes: '180x180',
        href: '/apple-touch-icon.png',
      },
      {
        rel: 'icon',
        type: 'image/png',
        sizes: '32x32',
        href: '/favicon-32x32.png',
      },
      {
        rel: 'icon',
        type: 'image/png',
        sizes: '16x16',
        href: '/favicon-16x16.png',
      },
      { rel: 'manifest', href: '/site.webmanifest', color: '#fffff' },
      { rel: 'icon', href: '/favicon.ico' },
    ],
    scripts: [
      {
        type: 'text/javascript',
        children: `
          if (typeof window !== 'undefined') {
            document.documentElement.classList.toggle(
              "dark",
              localStorage.theme === "dark" ||
                (!("theme" in localStorage) && window.matchMedia("(prefers-color-scheme: dark)").matches),
            );
          }`,
      },
    ],
  }),
  errorComponent: (props) => {
    return (
      <RootDocument>
        <DefaultCatchBoundary {...props} />
      </RootDocument>
    );
  },
  notFoundComponent: () => <NotFound />,
  component: RootComponent,
  beforeLoad: async () => {
    if (!isServer) {
      return null;
    }

    let accessToken = getAccessToken();
    if (!accessToken && getRefreshToken()) {
      const refreshResponse = await refreshFn();
      accessToken = refreshResponse.accessToken;
    }

    setRequestAccessToken(accessToken);

    return { accessToken };
  },
});

function RootComponent() {
  const serverData = Route.useRouteContext();

  const [accessToken] = useState<string | null>(serverData.accessToken);

  return (
    <RootDocument>
      <AuthProvider accessToken={accessToken}>
        <ChangesetProvider>
          <Header />
          <Outlet />
        </ChangesetProvider>
      </AuthProvider>
    </RootDocument>
  );
}

function RootDocument({ children }: { children: React.ReactNode }) {
  return (
    <html suppressHydrationWarning>
      <head>
        <HeadContent />
      </head>
      <body>
        {children}
        <TanStackRouterDevtools position="bottom-right" />
        <Scripts />
      </body>
    </html>
  );
}
