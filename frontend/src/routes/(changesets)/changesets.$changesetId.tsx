import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { SlimPage } from '~/components/SlimPage';
import { ChangesetChange } from './-components/ChangesetChange';
import { PageTitle } from '~/components/PageTitle';
import { DbChangesetStateEnum, getChangesetsChangesetIdQueryOptions, useGetChangesetsChangesetIdSuspense } from '~/gen';
import { VariantProps } from 'class-variance-authority';
import { Badge } from '~/components/ui/badge';
import { List, ListItem } from '~/components/List';
import { badgeVariants } from '~/components/ui/badge';
import { ChangesetActions } from './-components/ChangesetActions';
import { seo, appTitle } from '~/utils/seo';
const Schema = z.object({
  changesetId: z.coerce.number(),
});

export const Route = createFileRoute('/(changesets)/changesets/$changesetId')({
  component: RouteComponent,
  params: {
    parse: Schema.parse,
  },
  loader: async ({ context, params }) => {
    return context.queryClient.ensureQueryData(getChangesetsChangesetIdQueryOptions(params.changesetId));
  },
  head: ({ params }) => {
    return {
      meta: [...seo({ title: appTitle([`Changeset #${params.changesetId}`]) })],
    };
  },
});

function RouteComponent() {
  const { changesetId } = Route.useParams();

  const { data: changeset } = useGetChangesetsChangesetIdSuspense(changesetId);

  function getStateBadgeVariant(state: DbChangesetStateEnum): VariantProps<typeof badgeVariants>['variant'] {
    switch (state) {
      case 'open':
        return 'default';
      case 'discarded':
        return 'destructive';
      case 'applied':
        return 'secondary';
      case 'stashed':
        return 'outline';
      case 'committed':
        return 'outline';
      default:
        return 'default';
    }
  }

  return (
    <SlimPage>
      <PageTitle>Changeset #{changesetId}</PageTitle>
      <div className="flex flex-col gap-4">
        <div className="flex gap-2">
          <span className="font-semibold">State:</span>
          <Badge variant={getStateBadgeVariant(changeset.state)}>{changeset.state}</Badge>
        </div>
        {changeset.changes.length > 0 ? (
          <List>
            {changeset.changes.map((change) => (
              <ListItem key={change.id}>
                <ChangesetChange change={change} />
              </ListItem>
            ))}
          </List>
        ) : (
          <div className="text-muted-foreground">
            {changeset.state === 'discarded' ? 'Changeset has been discarded' : 'Changeset contains no changes'}
          </div>
        )}
        <ChangesetActions changeset={changeset} />
      </div>
    </SlimPage>
  );
}
