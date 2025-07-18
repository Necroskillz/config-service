/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, QueryObserverOptions, UseQueryResult } from '@tanstack/react-query'
import type {
  GetMembershipGroupsGroupIdUsersQueryResponse,
  GetMembershipGroupsGroupIdUsersPathParams,
  GetMembershipGroupsGroupIdUsersQueryParams,
  GetMembershipGroupsGroupIdUsers401,
  GetMembershipGroupsGroupIdUsers404,
  GetMembershipGroupsGroupIdUsers500,
} from '../types/GetMembershipGroupsGroupIdUsers.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useQuery } from '@tanstack/react-query'

export const getMembershipGroupsGroupIdUsersQueryKey = (
  group_id: GetMembershipGroupsGroupIdUsersPathParams['group_id'],
  params?: GetMembershipGroupsGroupIdUsersQueryParams,
) => [{ url: '/membership/groups/:group_id/users', params: { group_id: group_id } }, ...(params ? [params] : [])] as const

export type GetMembershipGroupsGroupIdUsersQueryKey = ReturnType<typeof getMembershipGroupsGroupIdUsersQueryKey>

/**
 * @description Get group users by ID
 * @summary Get group users
 * {@link /membership/groups/:group_id/users}
 */
export async function getMembershipGroupsGroupIdUsers(
  group_id: GetMembershipGroupsGroupIdUsersPathParams['group_id'],
  params?: GetMembershipGroupsGroupIdUsersQueryParams,
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetMembershipGroupsGroupIdUsersQueryResponse,
    ResponseErrorConfig<GetMembershipGroupsGroupIdUsers401 | GetMembershipGroupsGroupIdUsers404 | GetMembershipGroupsGroupIdUsers500>,
    unknown
  >({ method: 'GET', url: `/membership/groups/${group_id}/users`, params, ...requestConfig })
  return res.data
}

export function getMembershipGroupsGroupIdUsersQueryOptions(
  group_id: GetMembershipGroupsGroupIdUsersPathParams['group_id'],
  params?: GetMembershipGroupsGroupIdUsersQueryParams,
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getMembershipGroupsGroupIdUsersQueryKey(group_id, params)
  return queryOptions<
    GetMembershipGroupsGroupIdUsersQueryResponse,
    ResponseErrorConfig<GetMembershipGroupsGroupIdUsers401 | GetMembershipGroupsGroupIdUsers404 | GetMembershipGroupsGroupIdUsers500>,
    GetMembershipGroupsGroupIdUsersQueryResponse,
    typeof queryKey
  >({
    enabled: !!group_id,
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getMembershipGroupsGroupIdUsers(group_id, params, config)
    },
  })
}

/**
 * @description Get group users by ID
 * @summary Get group users
 * {@link /membership/groups/:group_id/users}
 */
export function useGetMembershipGroupsGroupIdUsers<
  TData = GetMembershipGroupsGroupIdUsersQueryResponse,
  TQueryData = GetMembershipGroupsGroupIdUsersQueryResponse,
  TQueryKey extends QueryKey = GetMembershipGroupsGroupIdUsersQueryKey,
>(
  group_id: GetMembershipGroupsGroupIdUsersPathParams['group_id'],
  params?: GetMembershipGroupsGroupIdUsersQueryParams,
  options: {
    query?: Partial<
      QueryObserverOptions<
        GetMembershipGroupsGroupIdUsersQueryResponse,
        ResponseErrorConfig<GetMembershipGroupsGroupIdUsers401 | GetMembershipGroupsGroupIdUsers404 | GetMembershipGroupsGroupIdUsers500>,
        TData,
        TQueryData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getMembershipGroupsGroupIdUsersQueryKey(group_id, params)

  const query = useQuery(
    {
      ...(getMembershipGroupsGroupIdUsersQueryOptions(group_id, params, config) as unknown as QueryObserverOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<QueryObserverOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseQueryResult<
    TData,
    ResponseErrorConfig<GetMembershipGroupsGroupIdUsers401 | GetMembershipGroupsGroupIdUsers404 | GetMembershipGroupsGroupIdUsers500>
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}