/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, QueryObserverOptions, UseQueryResult } from '@tanstack/react-query'
import type {
  GetMembershipUsersUserIdQueryResponse,
  GetMembershipUsersUserIdPathParams,
  GetMembershipUsersUserId400,
  GetMembershipUsersUserId401,
  GetMembershipUsersUserId404,
  GetMembershipUsersUserId500,
} from '../types/GetMembershipUsersUserId.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useQuery } from '@tanstack/react-query'

export const getMembershipUsersUserIdQueryKey = (user_id: GetMembershipUsersUserIdPathParams['user_id']) =>
  [{ url: '/membership/users/:user_id', params: { user_id: user_id } }] as const

export type GetMembershipUsersUserIdQueryKey = ReturnType<typeof getMembershipUsersUserIdQueryKey>

/**
 * @description Get a user by ID
 * @summary Get a user
 * {@link /membership/users/:user_id}
 */
export async function getMembershipUsersUserId(
  user_id: GetMembershipUsersUserIdPathParams['user_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetMembershipUsersUserIdQueryResponse,
    ResponseErrorConfig<GetMembershipUsersUserId400 | GetMembershipUsersUserId401 | GetMembershipUsersUserId404 | GetMembershipUsersUserId500>,
    unknown
  >({ method: 'GET', url: `/membership/users/${user_id}`, ...requestConfig })
  return res.data
}

export function getMembershipUsersUserIdQueryOptions(
  user_id: GetMembershipUsersUserIdPathParams['user_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getMembershipUsersUserIdQueryKey(user_id)
  return queryOptions<
    GetMembershipUsersUserIdQueryResponse,
    ResponseErrorConfig<GetMembershipUsersUserId400 | GetMembershipUsersUserId401 | GetMembershipUsersUserId404 | GetMembershipUsersUserId500>,
    GetMembershipUsersUserIdQueryResponse,
    typeof queryKey
  >({
    enabled: !!user_id,
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getMembershipUsersUserId(user_id, config)
    },
  })
}

/**
 * @description Get a user by ID
 * @summary Get a user
 * {@link /membership/users/:user_id}
 */
export function useGetMembershipUsersUserId<
  TData = GetMembershipUsersUserIdQueryResponse,
  TQueryData = GetMembershipUsersUserIdQueryResponse,
  TQueryKey extends QueryKey = GetMembershipUsersUserIdQueryKey,
>(
  user_id: GetMembershipUsersUserIdPathParams['user_id'],
  options: {
    query?: Partial<
      QueryObserverOptions<
        GetMembershipUsersUserIdQueryResponse,
        ResponseErrorConfig<GetMembershipUsersUserId400 | GetMembershipUsersUserId401 | GetMembershipUsersUserId404 | GetMembershipUsersUserId500>,
        TData,
        TQueryData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getMembershipUsersUserIdQueryKey(user_id)

  const query = useQuery(
    {
      ...(getMembershipUsersUserIdQueryOptions(user_id, config) as unknown as QueryObserverOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<QueryObserverOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseQueryResult<
    TData,
    ResponseErrorConfig<GetMembershipUsersUserId400 | GetMembershipUsersUserId401 | GetMembershipUsersUserId404 | GetMembershipUsersUserId500>
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}