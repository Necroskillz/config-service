import { useQueryClient } from '@tanstack/react-query';
import { createFileRoute, Link, useNavigate } from '@tanstack/react-router';
import { zodValidator } from '@tanstack/zod-adapter';
import { z } from 'zod';
import { DotDotDot } from '~/components/DotDotDot';
import { List, ListItem } from '~/components/List';
import { MutationErrors } from '~/components/MutationErrors';
import { PageTitle } from '~/components/PageTitle';
import { RenderPagedQuery } from '~/components/RenderPagedQuery';
import { Button } from '~/components/ui/button';
import { DropdownMenuItem } from '~/components/ui/dropdown-menu';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '~/components/ui/tabs';
import {
  CorePaginatedResultMembershipGroupUserDto,
  getMembershipGroupsGroupIdQueryOptions,
  getMembershipGroupsGroupIdUsersQueryOptions,
  MembershipGroupDto,
  MembershipMembershipObjectDto,
  useDeleteMembershipGroupsGroupId,
  useDeleteMembershipGroupsGroupIdUsersUserId,
  useDeleteMembershipPermissionsPermissionId,
  useGetMembershipGroupsGroupIdSuspense,
  useGetMembershipGroupsGroupIdUsers,
  usePostMembershipGroupsGroupIdUsersUserId,
} from '~/gen';
import { appTitle, seo } from '~/utils/seo';
import { Permission } from '~/components/Permission';
import { useAppForm } from '~/components/ui/tanstack-form-hook';
import { MemberPicker, requiredMember } from '~/components/MemberPicker';

const PAGE_SIZE = 20;

export const Route = createFileRoute('/_auth/(admin)/admin/(membership)/membership/groups/$groupId')({
  component: RouteComponent,
  validateSearch: zodValidator(
    z.object({
      tab: z.enum(['permissions', 'users']).optional(),
      userListPage: z.coerce.number().optional(),
    })
  ),
  params: {
    parse: z.object({
      groupId: z.coerce.number(),
    }).parse,
  },
  loaderDeps: ({ search: { userListPage, tab } }) => ({ userListPage, tab }),
  loader: async ({ context, params, deps }) => {
    const promises: [Promise<MembershipGroupDto>, Promise<CorePaginatedResultMembershipGroupUserDto> | undefined] = [
      context.queryClient.ensureQueryData(getMembershipGroupsGroupIdQueryOptions(params.groupId)),
      deps.tab === 'users'
        ? context.queryClient.ensureQueryData(
            getMembershipGroupsGroupIdUsersQueryOptions(params.groupId, { page: deps.userListPage ?? 1, pageSize: PAGE_SIZE })
          )
        : undefined,
    ];

    return Promise.all(promises);
  },
  head: ({ loaderData: [group] }) => ({
    meta: [...seo({ title: appTitle([`Edit Group ${group.name}`, 'Admin']) })],
  }),
});

function RouteComponent() {
  const { groupId } = Route.useParams();
  const { userListPage, tab } = Route.useSearch();
  const queryClient = useQueryClient();
  const navigate = useNavigate({ from: Route.fullPath });

  const { data: group } = useGetMembershipGroupsGroupIdSuspense(groupId);

  const deleteMutation = useDeleteMembershipGroupsGroupId({
    mutation: {
      onSuccess: () => {
        navigate({ to: '/admin/membership' });
      },
    },
  });

  const deletePermissionMutation = useDeleteMembershipPermissionsPermissionId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(getMembershipGroupsGroupIdQueryOptions(groupId));
      },
    },
  });

  const removeFromGroupMutation = useDeleteMembershipGroupsGroupIdUsersUserId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(getMembershipGroupsGroupIdUsersQueryOptions(groupId));
      },
    },
  });

  const usersQuery = useGetMembershipGroupsGroupIdUsers(
    groupId,
    {
      page: userListPage,
      pageSize: PAGE_SIZE,
    },
    {
      query: {
        enabled: !!userListPage,
      },
    }
  );

  const addUserToGroupMutation = usePostMembershipGroupsGroupIdUsersUserId({
    mutation: {
      onSuccess: () => {
        queryClient.invalidateQueries(getMembershipGroupsGroupIdUsersQueryOptions(groupId));
        addUserToGroupForm.reset();
      },
    },
  });

  const addUserToGroupForm = useAppForm({
    defaultValues: {
      member: undefined,
    } as { member?: MembershipMembershipObjectDto },
    validators: {
      onChange: z.object({
        member: requiredMember('User is required'),
      }),
    },
    onSubmit: ({ value }) => {
      addUserToGroupMutation.mutate({ user_id: value.member!.id, group_id: groupId });
    },
  });

  return (
    <div className="w-[1080px] flex flex-col gap-4 mt-4">
      <div className="flex items-center justify-between mb-4">
        <PageTitle className="mb-0">
          Group <pre className="inline">{group.name}</pre>
        </PageTitle>
        <div className="flex items-center">
          <DotDotDot>
            <DropdownMenuItem variant="destructive" onClick={() => deleteMutation.mutate({ group_id: groupId })}>
              Delete
            </DropdownMenuItem>
          </DotDotDot>
        </div>
      </div>
      <MutationErrors mutations={[deleteMutation]} />
      <Tabs
        defaultValue="permissions"
        value={tab}
        onValueChange={(value) =>
          navigate({
            to: '/admin/membership/groups/$groupId',
            search: { tab: value as any, userListPage: value === 'users' && !userListPage ? 1 : userListPage },
          })
        }
      >
        <TabsList>
          <TabsTrigger value="permissions">Permissions</TabsTrigger>
          <TabsTrigger value="users">Users</TabsTrigger>
        </TabsList>
        <TabsContent value="permissions">
          {group.permissions.length === 0 && <div className="text-muted-foreground">Group has no permissions</div>}
          {group.permissions.length > 0 && (
            <List>
              {group.permissions.map((permission) => (
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
        <TabsContent value="users" className="flex flex-col gap-4">
          <h2 className="text-lg font-semibold">Add user</h2>
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
                        type="user"
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

          <RenderPagedQuery
            query={usersQuery}
            page={userListPage ?? 1}
            pageSize={20}
            linkTo="/admin/membership/groups/$groupId"
            linkParams={{ groupId }}
            linkSearch={{ tab, userListPage }}
            emptyMessage="Group has no users"
            pageKey="userListPage"
          >
            {(data) => (
              <>
                <MutationErrors mutations={[removeFromGroupMutation]} />
                <List>
                  {data.map((user) => (
                    <ListItem key={user.id} variant="slim">
                      <div className="flex flex-row items-center gap-2 justify-between">
                        <Link to="/admin/membership/users/$userId" params={{ userId: user.id }}>
                          {user.name}
                        </Link>
                        <Button
                          variant="destructive"
                          size="sm"
                          onClick={() => removeFromGroupMutation.mutate({ user_id: user.id, group_id: groupId })}
                          disabled={removeFromGroupMutation.isPending}
                        >
                          Remove
                        </Button>
                      </div>
                    </ListItem>
                  ))}
                </List>
              </>
            )}
          </RenderPagedQuery>
        </TabsContent>
      </Tabs>
    </div>
  );
}
