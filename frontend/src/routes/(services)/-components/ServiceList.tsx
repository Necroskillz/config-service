import { Link } from '@tanstack/react-router';
import { List, ListItem } from '~/components/List';
import { useGetServicesSuspense } from '~/gen';
import { Badge } from '~/components/ui/badge';

export function ServiceList() {
  const { data: servicesVersions } = useGetServicesSuspense();

  return (
    <List>
      {servicesVersions?.map((serviceVersion) => (
        <ListItem key={serviceVersion.id}>
          <h2 className="text-lg font-bold">
            <Link to="/services/$serviceVersionId" params={{ serviceVersionId: serviceVersion.id }}>
              {serviceVersion.name}
            </Link>
            <Badge className="ml-2">v{serviceVersion.version}</Badge>
          </h2>
          <p className="text-sm text-muted-foreground">{serviceVersion.description}</p>
        </ListItem>
      ))}
    </List>
  );
}
