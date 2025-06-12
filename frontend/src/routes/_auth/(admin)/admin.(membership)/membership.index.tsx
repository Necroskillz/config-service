import { createFileRoute, Link, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import { useEffect, useRef, useState } from 'react';
import { buttonVariants } from '~/components/ui/button';
import { getMembershipQueryOptions, MembershipMembershipObjectType, useGetMembership } from '~/gen';
import { seo, appTitle } from '~/utils/seo';
import { zodValidator } from '@tanstack/zod-adapter';
import { SearchBox } from '~/components/ui/search-box';
import { useDebouncedCallback } from '@tanstack/react-pacer/debouncer';
import { RenderPagedQuery } from '~/components/RenderPagedQuery';
import { List, ListItem } from '~/components/List';

const PAGE_SIZE = 20;

export const Route = createFileRoute('/_auth/(admin)/admin/(membership)/membership/')({
  component: RouteComponent,
  validateSearch: zodValidator(
    z.object({
      page: z.number().min(1).default(1),
      name: z.string().optional(),
    })
  ),
  loaderDeps: ({ search: { name, page } }) => ({ name, page }),
  loader: async ({ context, deps }) => {
    return context.queryClient.ensureQueryData(
      getMembershipQueryOptions({ page: deps.page, pageSize: PAGE_SIZE, name: deps.name || undefined })
    );
  },
  head: () => ({
    meta: [...seo({ title: appTitle(['Users', 'Admin']) })],
  }),
});

const MembershipObjectType: Record<MembershipMembershipObjectType, string> = {
  user: 'User',
  global_administrator: 'Global Administrator',
  group: 'Group',
} as const;

function RouteComponent() {
  const { page, name } = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });
  const [search, setSearch] = useState(name ?? '');
  const query = useGetMembership({
    page: page,
    pageSize: PAGE_SIZE,
    name: name || undefined,
  });

  const pageRef = useRef(page);
  const isFirstSearch = useRef(true);

  const debouncedSearchCallback = useDebouncedCallback(
    (value: string) => {
      navigate({
        search: {
          page: 1,
          name: value || undefined,
        },
        replace: pageRef.current === 1 && !isFirstSearch.current,
      });
      isFirstSearch.current = false;
    },
    { wait: 200 }
  );

  function handleSearch(value: string) {
    setSearch(value);
    debouncedSearchCallback(value);
  }

  useEffect(() => {
    setSearch(name ?? '');
  }, [name]);

  useEffect(() => {
    pageRef.current = page;
  }, [page]);

  return (
    <div className="w-[720px] p-4 flex flex-row">
      <div className="w-full flex flex-col gap-4">
        <SearchBox value={search} onChange={(e) => handleSearch(e.target.value)} placeholder="Search users" />
        <RenderPagedQuery
          query={query}
          page={page}
          pageSize={PAGE_SIZE}
          linkTo="/admin/membership"
          linkSearch={{ page, name: search || undefined }}
          emptyMessage="No users found"
          pageKey="page"
        >
          {(data) => (
            <List>
              {data.map((obj) => (
                <ListItem key={obj.id} variant="slim">
                  <div className="flex flex-row gap-2 justify-between items-center">
                    {obj.type === 'group' ? (
                      <Link to="/admin/membership/groups/$groupId" params={{ groupId: obj.id }}>
                        {obj.name}
                      </Link>
                    ) : (
                      <Link to="/admin/membership/users/$userId" params={{ userId: obj.id }}>
                        {obj.name}
                      </Link>
                    )}
                    <p className="text-sm text-muted-foreground">{MembershipObjectType[obj.type]}</p>
                  </div>
                </ListItem>
              ))}
            </List>
          )}
        </RenderPagedQuery>
        <div className="mt-4 flex gap-4">
          <Link to="/admin/membership/users/create" className={buttonVariants({ variant: 'default', size: 'sm' })}>
            Create New User
          </Link>
          <Link to="/admin/membership/groups/create" className={buttonVariants({ variant: 'default', size: 'sm' })}>
            Create New Group
          </Link>
        </div>
      </div>
    </div>
  );
}
