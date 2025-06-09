import { createFileRoute, Link } from '@tanstack/react-router';
import { z } from 'zod';
import { useAuth } from '~/auth';
import { PageTitle } from '~/components/PageTitle';
import { Permission } from '~/components/Permission';
import { SlimPage } from '~/components/SlimPage';
import { buttonVariants } from '~/components/ui/button';
import { getMembershipUsersUserIdQueryOptions, useGetMembershipUsersUserIdSuspense } from '~/gen';
import { appTitle, seo } from '~/utils/seo';

export const Route = createFileRoute('/users/$userId')({
  component: RouteComponent,
  params: {
    parse: z.object({
      userId: z.coerce.number(),
    }).parse,
  },
  loader: async ({ params, context }) => {
    return context.queryClient.ensureQueryData(getMembershipUsersUserIdQueryOptions(params.userId));
  },
  head: ({ loaderData: user }) => {
    return {
      meta: [
        ...seo({
          title: appTitle([`User ${user.username}`]),
          description: user.username,
        }),
      ],
    };
  },
});

function RouteComponent() {
  const { userId } = Route.useParams();
  const { user: currentUser } = useAuth();

  const { data: user } = useGetMembershipUsersUserIdSuspense(userId);

  return (
    <SlimPage>
      <PageTitle>{user.username}</PageTitle>
      <div className="flex flex-col gap-4">
        <h2 className="text-lg font-semibold">User information</h2>
        <div className="flex flex-row gap-2">
          <div className="w-54">Is global admin</div>
          <div>{user.globalAdministrator ? 'Yes' : 'No'}</div>
        </div>
        <h2 className="text-lg font-semibold">Groups</h2>
        <div className="flex flex-col gap-2">
          {user.groups.length === 0 ? (
            <div className="text-muted-foreground">No groups</div>
          ) : (
            <ul className="list-disc list-inside pl-4">
              {user.groups.map((group) => (
                <li key={group.id}>
                  <Link to="/groups/$groupId" params={{ groupId: group.id }} className="link">
                    {group.name}
                  </Link>
                </li>
              ))}
            </ul>
          )}
        </div>
        <h2 className="text-lg font-semibold">Permissions</h2>
        <div className="flex flex-col gap-2">
          {user.permissions.length === 0 ? (
            <div className="text-muted-foreground">No permissions</div>
          ) : (
            <ul className="list-disc list-inside pl-4">
              {user.permissions.map((permission) => (
                <li key={permission.id}>
                  <Permission permission={permission} readOnly={true} />
                </li>
              ))}
            </ul>
          )}
        </div>
        {currentUser.isGlobalAdmin && (
          <div>
            <Link to="/admin/membership/users/$userId" params={{ userId }} className={buttonVariants({ variant: 'default', size: 'sm' })}>
              Go to admin view
            </Link>
          </div>
        )}
      </div>
    </SlimPage>
  );
}
