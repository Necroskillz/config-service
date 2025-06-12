import { UseQueryResult } from '@tanstack/react-query';
import { ResponseErrorConfig } from '~/axios';
import { EchoHTTPError } from '~/gen';
import { Pagination, PaginationContent, PaginationItem, PaginationLink, PaginationNext, PaginationPrevious } from './ui/pagination';
import { Spinner } from './ui/spinner';
import { LinkComponentProps } from '@tanstack/react-router';
import { Alert, AlertDescription } from './ui/alert';
import { RenderQuery } from './RenderQuery';

export type PagedQueryData<T> = {
  items: T[];
  totalCount: number;
};

export function RenderPagedQuery<T>({
  query,
  page,
  pageSize,
  children,
  linkTo,
  linkParams,
  linkSearch,
  emptyMessage,
  pageKey = 'page',
}: {
  query: UseQueryResult<PagedQueryData<T>, ResponseErrorConfig<EchoHTTPError>>;
  page: number;
  pageSize: number;
  children: (data: T[]) => React.ReactNode;
  linkTo: LinkComponentProps<'a'>['to'];
  linkParams?: LinkComponentProps<'a'>['params'];
  linkSearch?: LinkComponentProps<'a'>['search'];
  emptyMessage: string;
  pageKey: string;
}) {
  const { status } = query;

  const search = linkSearch as Record<string, any>;

  function getLinkSearch(page: number) {
    if (Object.hasOwn(search, pageKey)) {
      const copy = { ...search };
      copy[pageKey] = page;
      return copy;
    }

    return { ...search, [pageKey]: page };
  }

  return (
    <>
      <RenderQuery query={query}>
        {(data) => (
          <>
            {status === 'success' && data?.totalCount === 0 && page === 1 ? (
              <div className="text-muted-foreground">{emptyMessage}</div>
            ) : (
              <div className="flex flex-col gap-4">
                {children(data.items)}
                <Pagination>
                  <PaginationContent>
                    {page > 1 && (
                      <PaginationItem>
                        <PaginationPrevious to={linkTo} params={linkParams} search={getLinkSearch(page - 1)} />
                      </PaginationItem>
                    )}
                    <PaginationItem>
                      <PaginationLink to={linkTo} params={linkParams} search={getLinkSearch(page)} isActive>
                        {page}
                      </PaginationLink>
                    </PaginationItem>
                    {data && page < Math.ceil(data.totalCount / pageSize) && (
                      <PaginationItem>
                        <PaginationNext to={linkTo} params={linkParams} search={getLinkSearch(page + 1)} />
                      </PaginationItem>
                    )}
                  </PaginationContent>
                </Pagination>
              </div>
            )}
          </>
        )}
      </RenderQuery>
    </>
  );
}
