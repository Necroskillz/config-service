import { createFileRoute, Link } from '@tanstack/react-router';
import { useAuth } from '~/auth';
import { SlimPage } from '~/components/SlimPage';
import { PageTitle } from '~/components/PageTitle';
import { buttonVariants } from '~/components/ui/button';
import { getServicesQueryOptions, useGetServicesSuspense } from '~/gen';
import { List, ListItem } from '~/components/List';
import { Badge } from '~/components/ui/badge';
import { seo, appTitle } from '~/utils/seo';

export const Route = createFileRoute('/(services)/services/')({
  component: ServicesRouteComponent,
  loader: async ({ context }) => {
    return context.queryClient.ensureQueryData(getServicesQueryOptions());
  },
  head: () => {
    return {
      meta: [...seo({ title: appTitle(['Services']) })],
    };
  },
});

export function ServicesRouteComponent() {
  const { user } = useAuth();
  const { data: services } = useGetServicesSuspense();

  return (
    <SlimPage>
      <PageTitle>Services</PageTitle>
      <List>
        {services.map((service) => (
          <ListItem key={service.name}>
            <h2 className="text-lg font-bold">
              <Link to="/services/$serviceVersionId" params={{ serviceVersionId: service.versions[service.versions.length - 1].id }}>
                {service.name}
              </Link>
              {service.versions.map((version) => (
                <Link to="/services/$serviceVersionId" params={{ serviceVersionId: version.id }}>
                  <Badge variant={version.published ? 'default' : 'secondary'} className="ml-2" key={version.id}>
                    v{version.version} ({version.published ? 'published' : 'draft'})
                  </Badge>
                </Link>
              ))}
            </h2>
            <p className="text-sm text-muted-foreground">{service.description}</p>
          </ListItem>
        ))}
      </List>

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
