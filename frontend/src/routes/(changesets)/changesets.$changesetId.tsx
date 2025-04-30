import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { SlimPage } from '~/components/SlimPage';
import { ChangesetChange } from './-components/ChangesetChange';
import { PageTitle } from '~/components/PageTitle';
import { getChangesetsChangesetIdQueryOptions, useGetChangesetsChangesetIdSuspense } from '~/gen';
import { List, ListItem } from '~/components/List';
import { ChangesetActions } from './-components/ChangesetActions';
import { seo, appTitle } from '~/utils/seo';
import { ChangesetState } from './-components/ChangesetState';

export const Route = createFileRoute('/(changesets)/changesets/$changesetId')({
  component: RouteComponent,
  params: {
    parse: z.object({
      changesetId: z.coerce.number(),
    }).parse,
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


  return (
    <SlimPage>
      <PageTitle>Changeset #{changesetId}</PageTitle>
      <div className="flex flex-col gap-4">
        <div className="flex gap-2">
          <span className="font-semibold">State:</span>
          <ChangesetState state={changeset.state} />
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
