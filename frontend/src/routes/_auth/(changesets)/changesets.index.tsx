import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { z } from 'zod';
import { useAuth } from '~/auth';
import { PageTitle } from '~/components/PageTitle';
import { SlimPage } from '~/components/SlimPage';
import { useGetChangesets, getChangesetsQueryOptions } from '~/gen';
import { zodValidator } from '@tanstack/zod-adapter';
import { List, ListItem } from '~/components/List';
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from '~/components/ui/select';
import { Link } from '@tanstack/react-router';
import { ChangesetState } from './-components/ChangesetState';
import { Spinner } from '~/components/ui/spinner';
import { MessageSquareText } from 'lucide-react';
import { TimeAgo } from '~/components/TimeAgo';
import {
  Pagination,
  PaginationContent,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '~/components/ui/pagination';

const MODES = ['my', 'all', 'approvable'] as const;
const PAGE_SIZE = 20;

export const Route = createFileRoute('/_auth/(changesets)/changesets/')({
  component: RouteComponent,
  validateSearch: zodValidator(
    z.object({
      page: z.number().min(1).default(1),
      mode: z.enum(MODES).default('my'),
    })
  ),
  loader: async ({ context }) => {
    return context.queryClient.ensureQueryData(getChangesetsQueryOptions({ page: 1, pageSize: PAGE_SIZE, approvable: false }));
  },
});

function RouteComponent() {
  const { user } = useAuth();
  const { page, mode } = Route.useSearch();
  const navigate = useNavigate({ from: Route.fullPath });
  const { data: changesets, isLoading } = useGetChangesets({
    page: page,
    pageSize: PAGE_SIZE,
    authorId: mode === 'my' ? user?.id : undefined,
    approvable: mode === 'approvable',
  });

  return (
    <SlimPage>
      <PageTitle>Changesets</PageTitle>
      <div className="flex flex-col gap-4">
        <div className="flex gap-2">
          <Select value={mode} onValueChange={(value) => navigate({ search: { mode: value as (typeof MODES)[number] } })}>
            <SelectTrigger>
              <SelectValue placeholder="Select a mode" />
            </SelectTrigger>
            <SelectContent>
              {MODES.map((mode) => (
                <SelectItem key={mode} value={mode}>
                  {mode}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        {isLoading ? (
          <div className="flex justify-center items-center h-full mt-16">
            <Spinner />
          </div>
        ) : (
          <>
            <List>
              {changesets?.items.map((changeset) => (
                <ListItem key={changeset.id}>
                  <div className="flex justify-between items-center">
                    <div className="flex flex-col gap-2">
                      <h2 className="text-lg font-bold flex gap-2 items-center">
                        <Link to="/changesets/$changesetId" params={{ changesetId: changeset.id }}>
                          Changeset #{changeset.id}
                        </Link>
                        <ChangesetState state={changeset.state} />
                      </h2>
                      <p className="text-sm text-muted-foreground flex flex-col">
                        <span>
                          Author:{' '}
                          <Link className="link" to="/users/$userId" params={{ userId: changeset.userId }}>
                            {changeset.userName}
                          </Link>
                        </span>
                      </p>
                    </div>
                    <div className="flex flex-col items-end">
                      <div className="flex gap-2 items-center">
                        <span>{changeset.actionCount}</span>
                        <MessageSquareText className="w-6" />
                      </div>
                      <div className="flex gap-2 items-center">
                        <span className="text-sm text-muted-foreground">
                          Last action: <TimeAgo datetime={changeset.lastActionAt} />
                        </span>
                      </div>
                    </div>
                  </div>
                </ListItem>
              ))}
            </List>
            <Pagination>
              <PaginationContent>
                {page > 1 && (
                  <PaginationItem>
                    <PaginationPrevious to="/changesets" search={{ mode, page: page - 1 }} />
                  </PaginationItem>
                )}
                <PaginationItem>
                  <PaginationLink to="/changesets" search={{ mode, page }} isActive>
                    {page}
                  </PaginationLink>
                </PaginationItem>
                {page < Math.ceil(changesets!.totalCount / PAGE_SIZE) && (
                  <PaginationItem>
                    <PaginationNext to="/changesets" search={{ mode, page: page + 1 }} />
                  </PaginationItem>
                )}
              </PaginationContent>
            </Pagination>
          </>
        )}
      </div>
    </SlimPage>
  );
}
