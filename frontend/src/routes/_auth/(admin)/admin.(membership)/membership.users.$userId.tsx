import { createFileRoute, Link, useNavigate } from '@tanstack/react-router';
import { Switch } from '~/components/ui/switch';
import { Button } from '~/components/ui/button';
import { seo, appTitle } from '~/utils/seo';
import { z } from 'zod';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import {
  getMembershipUsersUserIdQueryOptions,
  MembershipMembershipObjectDto,
  useDeleteMembershipGroupsGroupIdUsersUserId,
  useDeleteMembershipPermissionsPermissionId,
  useDeleteMembershipUsersUserId,
  useGetMembershipUsersUserIdSuspense,
  usePostMembershipGroupsGroupIdUsersUserId,
  usePutMembershipUsersUserId,
} from '~/gen';
import { MutationErrors } from '~/components/MutationErrors';
import { DropdownMenuItem } from '~/components/ui/dropdown-menu';
import { DotDotDot } from '~/components/DotDotDot';
import { PageTitle } from '~/components/PageTitle';
import { useQueryClient } from '@tanstack/react-query';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '~/components/ui/tabs';
import { List, ListItem } from '~/components/List';
import { zodValidator } from '@tanstack/zod-adapter';
import { Permission } from '~/components/Permission';
import { MemberPicker, requiredMember } from '~/components/MemberPicker';

export const Route = createFileRoute('/_auth/(admin)/admin/(membership)/membership/users/$userId')({
  component: RouteComponent,
  params: {
    parse: z.object({
      userId: z.coerce.number(),
    }).parse,
  },
  validateSearch: zodValidator(
    z.object({
      tab: z.enum(['permissions', 'groups']).optional(),
    })
  ),
  loader: async ({ context, params }) => {
    return context.queryClient.ensureQueryData(getMembershipUsersUserIdQueryOptions(params.userId));
  },
  head: ({ loaderData }) => ({
    meta: [...seo({ title: appTitle([loaderData ? 'Edit User' : 'Create User', 'Admin']) })],
  }),
});

function RouteComponent() {
  const { userId } = Route.useParams();
  const { tab } = Route.useSearch();
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: user } = useGetMembershipUsersUserIdSuspense(userId);

  const updateMutation = usePutMembershipUsersUserId();

  const deleteMutation = useDeleteMembershipUsersUserId({
    mutation: {
      onSuccess: async () => {
        navigate({ to: '/admin/membership' });
      },
    },
  });

  const deletePermissionMutation = useDeleteMembershipPermissionsPermissionId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(getMembershipUsersUserIdQueryOptions(userId));
      },
    },
  });

  const removeFromGroupMutation = useDeleteMembershipGroupsGroupIdUsersUserId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(getMembershipUsersUserIdQueryOptions(userId));
      },
    },
  });

  const form = useAppForm({
    defaultValues: {
      globalAdministrator: user.globalAdministrator,
    },
    validators: {
      onChange: z.object({
        globalAdministrator: z.boolean(),
      }),
    },
    onSubmit: async ({ value }) => {
      await updateMutation.mutateAsync({ user_id: userId, data: { globalAdministrator: value.globalAdministrator } });
    },
  });

  const addUserToGroupMutation = usePostMembershipGroupsGroupIdUsersUserId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(getMembershipUsersUserIdQueryOptions(userId));
      },
    },
  });

  const addUserToGroupForm = useAppForm({
    defaultValues: {
      member: undefined,
    } as { member?: MembershipMembershipObjectDto },
    validators: {
      onChange: z.object({
        member: requiredMember('Group is required'),
      }),
    },
    onSubmit: ({ value }) => {
      addUserToGroupMutation.mutate({ user_id: userId, group_id: value.member!.id });
    },
  });

  return (
    <div className="p-4 w-[1080px] flex flex-col gap-4">
      <div className="flex items-center justify-between mb-4">
        <PageTitle className="mb-0">
          User <pre className="inline">{user.username}</pre>
        </PageTitle>
        <div className="flex items-center">
          <DotDotDot>
            <DropdownMenuItem variant="destructive" onClick={() => deleteMutation.mutate({ user_id: userId })}>
              Delete
            </DropdownMenuItem>
          </DotDotDot>
        </div>
      </div>
      <form.AppForm>
        <form
          className="flex flex-col gap-4"
          onSubmit={(e) => {
            e.preventDefault();
            e.stopPropagation();
            form.handleSubmit();
          }}
        >
          <MutationErrors mutations={[updateMutation, deleteMutation]} />

          <form.AppField
            name="globalAdministrator"
            children={(field) => (
              <div className="flex items-center space-x-2">
                <field.FormControl>
                  <Switch id={field.name} name={field.name} checked={field.state.value} onCheckedChange={field.handleChange} />
                </field.FormControl>
                <field.FormLabel htmlFor={field.name}>Global Administrator</field.FormLabel>
                <field.FormMessage />
              </div>
            )}
          />

          <div>
            <form.Subscribe
              selector={(state) => [state.canSubmit, state.isSubmitting]}
              children={([canSubmit, isSubmitting]) => (
                <Button type="submit" disabled={!canSubmit || isSubmitting}>
                  Update
                </Button>
              )}
            />
          </div>
        </form>
      </form.AppForm>
      <Tabs
        defaultValue="permissions"
        value={tab}
        onValueChange={(value) =>
          navigate({
            to: '/admin/membership/users/$userId',
            params: { userId },
            search: { tab: value as any },
          })
        }
      >
        <TabsList>
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
          <TabsTrigger value="groups">Groups</TabsTrigger>
        </TabsList>
        <TabsContent value="permissions">
          {user.permissions.length === 0 && <div className="text-muted-foreground">Group has no permissions</div>}
          {user.permissions.length > 0 && (
            <List>
              {user.permissions.map((permission) => (
                <ListItem key={permission.id} variant="slim">
                  <Permission
                    permission={permission}
                    onDelete={() => deletePermissionMutation.mutate({ permission_id: permission.id })}
                    disabled={deletePermissionMutation.isPending}
                  />
                </ListItem>
              ))}
            </List>
          )}
        </TabsContent>
        <TabsContent value="groups" className="flex flex-col gap-4">
          <h2 className="text-lg font-semibold">Add to group</h2>
          <addUserToGroupForm.AppForm>
            <form
              className="flex flex-col gap-4"
              onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                addUserToGroupForm.handleSubmit();
              }}
            >
              <MutationErrors mutations={[addUserToGroupMutation]} />
              <addUserToGroupForm.AppField name="member">
                {(field) => (
                  <>
                    <field.FormControl>
                      <MemberPicker
                        value={field.state.value}
                        onValueChange={(value) => field.handleChange(value)}
                        onBlur={() => field.handleBlur()}
                        type="group"
                      />
                    </field.FormControl>
                    <field.FormMessage />
                  </>
                )}
              </addUserToGroupForm.AppField>
              <div>
                <addUserToGroupForm.Subscribe
                  selector={(state) => [state.canSubmit, state.isSubmitting]}
                  children={([canSubmit, isSubmitting]) => (
                    <Button type="submit" disabled={!canSubmit || isSubmitting}>
                      Add
                    </Button>
                  )}
                />
              </div>
            </form>
          </addUserToGroupForm.AppForm>
          {user.groups.length === 0 && <div className="text-muted-foreground">User is not in any groups</div>}
          {user.groups.length > 0 && (
            <List>
              {user.groups.map((group) => (
                <ListItem key={group.id} variant="slim">
                  <div className="flex flex-row items-center gap-2 justify-between">
                    <Link to="/admin/membership/groups/$groupId" params={{ groupId: group.id }}>
                      {group.name}
                    </Link>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => removeFromGroupMutation.mutate({ user_id: userId, group_id: group.id })}
                      disabled={removeFromGroupMutation.isPending}
                    >
                      Remove
                    </Button>
                  </div>
                </ListItem>
              ))}
            </List>
          )}
        </TabsContent>
      </Tabs>
    </div>
  );
}
