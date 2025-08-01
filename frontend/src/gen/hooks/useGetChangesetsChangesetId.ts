/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, QueryObserverOptions, UseQueryResult } from '@tanstack/react-query'
import type {
  GetChangesetsChangesetIdQueryResponse,
  GetChangesetsChangesetIdPathParams,
  GetChangesetsChangesetId400,
  GetChangesetsChangesetId401,
  GetChangesetsChangesetId404,
  GetChangesetsChangesetId500,
} from '../types/GetChangesetsChangesetId.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useQuery } from '@tanstack/react-query'

export const getChangesetsChangesetIdQueryKey = (changeset_id: GetChangesetsChangesetIdPathParams['changeset_id']) =>
  [{ url: '/changesets/:changeset_id', params: { changeset_id: changeset_id } }] as const

export type GetChangesetsChangesetIdQueryKey = ReturnType<typeof getChangesetsChangesetIdQueryKey>

/**
 * @description Get a changeset by ID
 * @summary Get a changeset
 * {@link /changesets/:changeset_id}
 */
export async function getChangesetsChangesetId(
  changeset_id: GetChangesetsChangesetIdPathParams['changeset_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetChangesetsChangesetIdQueryResponse,
    ResponseErrorConfig<GetChangesetsChangesetId400 | GetChangesetsChangesetId401 | GetChangesetsChangesetId404 | GetChangesetsChangesetId500>,
    unknown
  >({ method: 'GET', url: `/changesets/${changeset_id}`, ...requestConfig })
  return res.data
}

export function getChangesetsChangesetIdQueryOptions(
  changeset_id: GetChangesetsChangesetIdPathParams['changeset_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getChangesetsChangesetIdQueryKey(changeset_id)
  return queryOptions<
    GetChangesetsChangesetIdQueryResponse,
    ResponseErrorConfig<GetChangesetsChangesetId400 | GetChangesetsChangesetId401 | GetChangesetsChangesetId404 | GetChangesetsChangesetId500>,
    GetChangesetsChangesetIdQueryResponse,
    typeof queryKey
  >({
    enabled: !!changeset_id,
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getChangesetsChangesetId(changeset_id, config)
    },
  })
}

/**
 * @description Get a changeset by ID
 * @summary Get a changeset
 * {@link /changesets/:changeset_id}
 */
export function useGetChangesetsChangesetId<
  TData = GetChangesetsChangesetIdQueryResponse,
  TQueryData = GetChangesetsChangesetIdQueryResponse,
  TQueryKey extends QueryKey = GetChangesetsChangesetIdQueryKey,
>(
  changeset_id: GetChangesetsChangesetIdPathParams['changeset_id'],
  options: {
    query?: Partial<
      QueryObserverOptions<
        GetChangesetsChangesetIdQueryResponse,
        ResponseErrorConfig<GetChangesetsChangesetId400 | GetChangesetsChangesetId401 | GetChangesetsChangesetId404 | GetChangesetsChangesetId500>,
        TData,
        TQueryData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getChangesetsChangesetIdQueryKey(changeset_id)

  const query = useQuery(
    {
      ...(getChangesetsChangesetIdQueryOptions(changeset_id, config) as unknown as QueryObserverOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<QueryObserverOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseQueryResult<
    TData,
    ResponseErrorConfig<GetChangesetsChangesetId400 | GetChangesetsChangesetId401 | GetChangesetsChangesetId404 | GetChangesetsChangesetId500>
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}