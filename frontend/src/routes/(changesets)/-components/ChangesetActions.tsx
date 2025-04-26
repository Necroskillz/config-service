import { useQueryClient } from '@tanstack/react-query';
import { Link } from '@tanstack/react-router';
import { useAuth } from '~/auth';
import { List, ListItem } from '~/components/List';
import { MutationErrors } from '~/components/MutationErrors';
import { Button } from '~/components/ui/button';
import { useAppForm } from '~/components/ui/tanstack-form';
import { Textarea } from '~/components/ui/textarea';
import TimeAgo from 'react-timeago';
import {
  DbChangesetActionTypeEnum,
  ServiceChangesetDto,
  useDeleteChangesetsChangesetIdDiscard,
  getChangesetsChangesetIdQueryKey,
  usePostChangesetsChangesetIdComment,
  usePutChangesetsChangesetIdApply,
  usePutChangesetsChangesetIdCommit,
  usePutChangesetsChangesetIdReopen,
  usePutChangesetsChangesetIdStash,
} from '~/gen';

function getActionText(type: DbChangesetActionTypeEnum): string {
  switch (type) {
    case 'comment':
      return 'commented';
    case 'apply':
      return 'applied changeset';
    case 'commit':
      return 'committed changeset';
    case 'discard':
      return 'discarded changeset';
    case 'reopen':
      return 'reopened changeset';
    case 'stash':
      return 'stashed changeset';
    default:
      return 'unknown action';
  }
}

export function ChangesetActions({ changeset }: { changeset: ServiceChangesetDto }) {
  const { user } = useAuth();

  const queryClient = useQueryClient();

  const mutations = {
    apply: usePutChangesetsChangesetIdApply(),
    commit: usePutChangesetsChangesetIdCommit(),
    discard: useDeleteChangesetsChangesetIdDiscard(),
    reopen: usePutChangesetsChangesetIdReopen(),
    stash: usePutChangesetsChangesetIdStash(),
    comment: usePostChangesetsChangesetIdComment(),
  };

  const form = useAppForm({
    defaultValues: {
      comment: '',
    },
    onSubmit: async ({ value, meta }: { value: { comment: string }; meta: { action: DbChangesetActionTypeEnum } }) => {
      const mutation = mutations[meta.action];
      await mutation.mutateAsync({ changeset_id: changeset.id, data: { comment: value.comment } });
      queryClient.refetchQueries({ queryKey: getChangesetsChangesetIdQueryKey(changeset.id) });
      form.reset();
    },
  });

  return (
    <div className="flex flex-col gap-4">
      <h2 className="text-lg font-semibold">Actions</h2>
      {changeset.actions.length > 0 && (
        <List>
          {changeset.actions.map((action) => (
            <ListItem key={action.id}>
              <div className="flex items-center gap-2">
                <div className="text-sm text-muted-foreground">
                  <TimeAgo title={new Date(action.createdAt).toLocaleString()} date={action.createdAt} />
                </div>
                <div>
                  <Link className="link" to="/users/$userId" params={{ userId: action.userId }}>
                    {action.userName}
                  </Link>
                </div>
                <em>{getActionText(action.type)}</em>
              </div>
              {action.comment && <div>{action.comment}</div>}
            </ListItem>
          ))}
        </List>
      )}
      <MutationErrors mutations={Object.values(mutations)} />
      <div>
        {changeset.changes.length > 0 && (
          <form.AppForm>
            <div className="flex flex-col gap-4">
              <form.AppField
                name="comment"
                children={(field) => (
                  <field.FormItem>
                    <field.FormLabel htmlFor={field.name}>Comment</field.FormLabel>
                    <field.FormControl>
                      <Textarea
                        id={field.name}
                        name={field.name}
                        value={field.state.value}
                        onChange={(e) => field.handleChange(e.target.value)}
                        onBlur={field.handleBlur}
                      />
                    </field.FormControl>
                    <field.FormMessage />
                  </field.FormItem>
                )}
              />

              <form.Subscribe
                selector={(state) => ({
                  canSubmit: state.canSubmit,
                  isSubmitting: state.isSubmitting,
                  comment: state.values.comment,
                })}
                children={({ canSubmit, isSubmitting, comment }) => (
                  <div className="flex gap-2">
                    <Button disabled={!canSubmit || isSubmitting} onClick={() => form.handleSubmit({ action: 'apply' })}>
                      {comment ? 'Comment and Apply' : 'Apply'}
                    </Button>
                    {changeset.userId === user.id && changeset.state === 'open' && (
                      <>
                        <Button
                          variant={changeset.canApply ? 'secondary' : 'default'}
                          disabled={isSubmitting}
                          onClick={() => form.handleSubmit({ action: 'commit' })}
                        >
                          {comment ? 'Comment and Commit' : 'Commit'}
                        </Button>
                      </>
                    )}
                    <Button
                      variant="outline"
                      disabled={!canSubmit || isSubmitting || !comment}
                      onClick={() => form.handleSubmit({ action: 'comment' })}
                    >
                      Comment
                    </Button>
                    {changeset.userId === user.id && (
                      <>
                        {changeset.state === 'open' && (
                          <Button variant={'outline'} disabled={isSubmitting} onClick={() => form.handleSubmit({ action: 'stash' })}>
                            Stash
                          </Button>
                        )}
                        {(changeset.state === 'open' || changeset.state === 'committed' || changeset.state === 'stashed') && (
                          <Button variant={'destructive'} disabled={isSubmitting} onClick={() => form.handleSubmit({ action: 'discard' })}>
                            Discard
                          </Button>
                        )}
                        {(changeset.state === 'stashed' || changeset.state === 'committed') && (
                          <Button variant={'outline'} disabled={isSubmitting} onClick={() => form.handleSubmit({ action: 'reopen' })}>
                            Reopen
                          </Button>
                        )}
                      </>
                    )}
                  </div>
                )}
              />
            </div>
          </form.AppForm>
        )}
      </div>
    </div>
  );
}
