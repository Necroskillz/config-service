import { createFileRoute, Link, Outlet } from '@tanstack/react-router';
import { z } from 'zod';
import { useState } from 'react';
import { useDebounce } from 'use-debounce';
import { buttonVariants } from '~/components/ui/button';
import { getUsersQueryOptions, useGetUsers } from '~/gen';
import { seo, appTitle } from '~/utils/seo';
import { zodValidator } from '@tanstack/zod-adapter';
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '~/components/ui/pagination';
import { Spinner } from '~/components/ui/spinner';
import { SearchBox } from '~/components/ui/search-box';

const PAGE_SIZE = 20;

export const Route = createFileRoute('/(admin)/admin/(users)/users/')({
  component: RouteComponent,
  validateSearch: zodValidator(
    z.object({
      page: z.number().min(1).default(1),
    })
  ),
  loader: async ({ context }) => {
    return context.queryClient.ensureQueryData(getUsersQueryOptions({ page: 1, pageSize: PAGE_SIZE }));
  },
  head: () => ({
    meta: [...seo({ title: appTitle(['Users', 'Admin']) })],
  }),
});

function RouteComponent() {
  const { page } = Route.useSearch();
  const [search, setSearch] = useState('');
  const [debouncedSearch] = useDebounce(search, 200);
  const { data: users, isLoading } = useGetUsers({
    page: page,
    pageSize: PAGE_SIZE,
    name: debouncedSearch.length > 0 ? debouncedSearch : undefined,
  });

  return (
    <div className="w-[720px] p-4 flex flex-row">
      <div className="w-full flex flex-col gap-2">
        <SearchBox value={search} onChange={(e) => setSearch(e.target.value)} placeholder="Search users" />
        {isLoading ? (
          <div className="flex justify-center items-center h-full mt-16">
            <Spinner />
          </div>
        ) : (
          users?.items.map((user) => (
            <Link key={user.id} to="/admin/users/$userId" params={{ userId: user.id }}>
              <div className="flex justify-between items-center">
                <div className="flex flex-col gap-2">
                  <h2 className="text-lg font-bold">{user.username}</h2>
                  <p className="text-sm text-muted-foreground">
                    {user.globalAdministrator ? 'Global Admin' : 'User'}
                  </p>
                </div>
              </div>
            </Link>
          ))
        )}
        <Pagination>
          <PaginationContent>
            {page > 1 && (
              <PaginationItem>
                <PaginationPrevious to="/admin/users" search={{ page: page - 1 }} />
              </PaginationItem>
            )}
            <PaginationItem>
              <PaginationLink to="/admin/users" search={{ page }} isActive>
                {page}
              </PaginationLink>
            </PaginationItem>
            {users && page < Math.ceil(users.totalCount / PAGE_SIZE) && (
              <PaginationItem>
                <PaginationNext to="/admin/users" search={{ page: page + 1 }} />
              </PaginationItem>
            )}
          </PaginationContent>
        </Pagination>
        <div className="mt-4">
          <Link to="/admin/users/$userId" params={{ userId: 'create' }} className={buttonVariants({ variant: 'default', size: 'sm' })}>
            Create New User
          </Link>
        </div>
      </div>
    </div>
  );
} 