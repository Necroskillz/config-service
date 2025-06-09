import { createFileRoute, Link } from '@tanstack/react-router';
import { zodValidator } from '@tanstack/zod-adapter';
import { z } from 'zod';
import { useAuth } from '~/auth';
import { List, ListItem } from '~/components/List';
import { PageTitle } from '~/components/PageTitle';
import { Permission } from '~/components/Permission';
import { RenderPagedQuery } from '~/components/RenderPagedQuery';
import { SlimPage } from '~/components/SlimPage';
import { buttonVariants } from '~/components/ui/button';
import { getMembershipGroupsGroupIdQueryOptions, useGetMembershipGroupsGroupIdSuspense, useGetMembershipGroupsGroupIdUsers } from '~/gen';
import { appTitle, seo } from '~/utils/seo';

export const Route = createFileRoute('/groups/$groupId')({
  component: RouteComponent,
  params: {
    parse: z.object({
      groupId: z.coerce.number(),
    }).parse,
  },
  validateSearch: zodValidator(
    z.object({
      userListPage: z.coerce.number().optional(),
    })
  ),
  loader: async ({ params, context }) => {
    return context.queryClient.ensureQueryData(getMembershipGroupsGroupIdQueryOptions(params.groupId));
  },
  head: ({ loaderData: group }) => {
    return {
      meta: [
        ...seo({
          title: appTitle([`Group ${group.name}`]),
          description: group.name,
        }),
      ],
    };
  },
});

function RouteComponent() {
  const { groupId } = Route.useParams();
  const { userListPage } = Route.useSearch();
  const { user: currentUser } = useAuth();

  const { data: group } = useGetMembershipGroupsGroupIdSuspense(groupId);
  const usersQuery = useGetMembershipGroupsGroupIdUsers(groupId, {
    page: userListPage,
    pageSize: 20,
  });

  return (
    <SlimPage>
      <PageTitle>{group.name}</PageTitle>
      <div className="flex flex-col gap-4">
        <h2 className="text-lg font-semibold">Permissions</h2>
        <div className="flex flex-col gap-2">
          {group.permissions.length === 0 ? (
            <div className="text-muted-foreground">No permissions</div>
          ) : (
            <ul className="list-disc list-inside pl-4">
              {group.permissions.map((permission) => (
                <li key={permission.id}>
                  <Permission permission={permission} readOnly={true} />
                </li>
              ))}
            </ul>
          )}
        </div>
        <h2 className="text-lg font-semibold">Members</h2>
        <RenderPagedQuery
          query={usersQuery}
          page={userListPage ?? 1}
          pageSize={20}
          linkTo="/groups/$groupId"
          linkParams={{ groupId }}
          linkSearch={{ userListPage }}
          emptyMessage="Group has no users"
          pageKey="userListPage"
        >
          {(data) => (
            <List>
              {data.items.map((user) => (
                <ListItem key={user.id} variant="slim">
                  <Link to="/users/$userId" params={{ userId: user.id }}>
                    {user.name}
                  </Link>
                </ListItem>
              ))}
            </List>
          )}
        </RenderPagedQuery>
        {currentUser.isGlobalAdmin && (
          <div>
            <Link
              to="/admin/membership/groups/$groupId"
              params={{ groupId }}
              className={buttonVariants({ variant: 'default', size: 'sm' })}
            >
              Go to admin view
            </Link>
          </div>
        )}
      </div>
    </SlimPage>
  );
}
