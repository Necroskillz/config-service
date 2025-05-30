import { createFileRoute } from '@tanstack/react-router';
import { getAccessToken, useAuth } from '../auth';
import { redirect, Outlet } from '@tanstack/react-router';

export const Route = createFileRoute('/_auth')({
  component: RouteComponent,
  beforeLoad: async () => {
    const accessToken = getAccessToken();
    if (!accessToken) {
      throw redirect({ to: '/login' });
    }
  },
});

function RouteComponent() {
  return <Outlet />;
}
