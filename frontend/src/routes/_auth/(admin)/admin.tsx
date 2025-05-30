import { createFileRoute, Outlet } from '@tanstack/react-router';
import { AdminSidebar } from './-components/AdminSidebar';
import { seo, appTitle } from '~/utils/seo';

export const Route = createFileRoute('/_auth/(admin)/admin')({
  component: RouteComponent,
  head: () => ({
    meta: [...seo({ title: appTitle(['Admin']) })],
  }),
});

function RouteComponent() {
  return (
    <div className="flex flex-row">
      <AdminSidebar />
      <Outlet />
    </div>
  );
}
