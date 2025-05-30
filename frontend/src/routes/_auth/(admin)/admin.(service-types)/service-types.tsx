import { createFileRoute, Link, Outlet } from '@tanstack/react-router';
import { buttonVariants } from '~/components/ui/button';
import { getServiceTypesQueryOptions, useGetServiceTypesSuspense } from '~/gen';
import { seo, appTitle } from '~/utils/seo';

export const Route = createFileRoute('/_auth/(admin)/admin/(service-types)/service-types')({
  component: RouteComponent,
  loader: async ({ context }) => {
    return context.queryClient.ensureQueryData(getServiceTypesQueryOptions());
  },
  head: () => ({
    meta: [...seo({ title: appTitle(['Service Types', 'Admin']) })],
  }),
});

function RouteComponent() {
  const { data: serviceTypes } = useGetServiceTypesSuspense();

  return (
    <div className="p-4 flex flex-row">
      <div className="w-64 flex flex-col gap-2">
        {serviceTypes.map((serviceType) => (
          <Link key={serviceType.id} to="/admin/service-types/$serviceTypeId" params={{ serviceTypeId: serviceType.id }}>
            {serviceType.name}
          </Link>
        ))}
        <div className="mt-4">
          <Link to="/admin/service-types/create" className={buttonVariants({ variant: 'default', size: 'sm' })}>
            Create New Service Type
          </Link>
        </div>
      </div>
      <Outlet />
    </div>
  );
}
