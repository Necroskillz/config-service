import { createFileRoute, Link, Outlet } from '@tanstack/react-router';
import { buttonVariants } from '~/components/ui/button';
import { getVariationPropertiesQueryOptions, useGetVariationPropertiesSuspense } from '~/gen';
import { seo, appTitle } from '~/utils/seo';

export const Route = createFileRoute('/_auth/(admin)/admin/(variation-properties)/variation-properties')({
  component: RouteComponent,
  loader: async ({ context }) => {
    return context.queryClient.ensureQueryData(getVariationPropertiesQueryOptions());
  },
  head: () => ({
    meta: [...seo({ title: appTitle(['Variation Properties', 'Admin']) })],
  }),
});

function RouteComponent() {
  const { data: variationProperties } = useGetVariationPropertiesSuspense();

  return (
    <div className="p-4 flex flex-row">
      <div className="w-52 flex flex-col gap-2">
        {variationProperties.map((property) => (
          <Link
            className="text-ellipsis overflow-hidden text-nowrap"
            title={`${property.name}${property.displayName !== property.name ? ` (${property.displayName})` : ''}`}
            key={property.id}
            to="/admin/variation-properties/$propertyId"
            params={{ propertyId: property.id }}
          >
            {property.name} {property.displayName !== property.name && `(${property.displayName})`}
          </Link>
        ))}
        <div className="mt-4">
          <Link to="/admin/variation-properties/create" className={buttonVariants({ variant: 'default', size: 'sm' })}>
            Create New Property
          </Link>
        </div>
      </div>
      <Outlet />
    </div>
  );
}
