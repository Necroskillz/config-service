import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { SlimPage } from '~/components/SlimPage';
import { ChangesetChange } from './-components/ChangesetChange';
import { PageTitle } from '~/components/PageTitle';
import {
  getChangesetsChangesetIdQueryKey,
  getChangesetsChangesetIdQueryOptions,
  useDeleteChangesetsChangesetIdChangesChangeId,
  useGetChangesetsChangesetIdSuspense,
} from '~/gen';
import { List, ListItem } from '~/components/List';
import { ChangesetActions } from './-components/ChangesetActions';
import { seo, appTitle } from '~/utils/seo';
import { ChangesetState } from './-components/ChangesetState';
import { useQueryClient } from '@tanstack/react-query';
import { useChangeset } from '~/hooks/use-changeset';
import { MutationErrors } from '~/components/MutationErrors';
import { Alert, AlertDescription } from '~/components/ui/alert';
import { TriangleAlert } from 'lucide-react';

export const Route = createFileRoute('/_auth/(changesets)/changesets/$changesetId')({
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
  const queryClient = useQueryClient();
  const { refresh } = useChangeset();

  const { data: changeset } = useGetChangesetsChangesetIdSuspense(changesetId);

  const discardChangeMutation = useDeleteChangesetsChangesetIdChangesChangeId({
    mutation: {
      onSuccess: () => {
        queryClient.refetchQueries({ queryKey: getChangesetsChangesetIdQueryKey(changeset.id) });
        refresh();
      },
    },
  });

  return (
    <SlimPage>
      <PageTitle>Changeset #{changesetId}</PageTitle>
      <div className="flex flex-col gap-4">
        <div className="flex gap-2">
          <span className="font-semibold">State:</span>
          <ChangesetState state={changeset.state} />
        </div>
        {changeset.conflictCount > 0 && (
          <Alert variant="destructive">
            <TriangleAlert />
            <AlertDescription>
              <div>Changeset has conflicts that need to be resolved before it can be applied or committed.</div>
              <div>Number of conflicts: {changeset.conflictCount}</div>
            </AlertDescription>
          </Alert>
        )}
        <MutationErrors mutations={[discardChangeMutation]} />
        {changeset.changes.length > 0 ? (
          <List>
            {changeset.changes.map((change) => (
              <ListItem key={change.id}>
                <ChangesetChange
                  changeset={changeset}
                  change={change}
                  onDiscard={() => discardChangeMutation.mutate({ changeset_id: changeset.id, change_id: change.id })}
                />
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
