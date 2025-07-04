/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, UseSuspenseQueryOptions, UseSuspenseQueryResult } from '@tanstack/react-query'
import type {
  GetConfigurationChangesetsQueryResponse,
  GetConfigurationChangesetsQueryParams,
  GetConfigurationChangesets400,
  GetConfigurationChangesets404,
  GetConfigurationChangesets500,
} from '../types/GetConfigurationChangesets.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useSuspenseQuery } from '@tanstack/react-query'

export const getConfigurationChangesetsSuspenseQueryKey = (params: GetConfigurationChangesetsQueryParams) =>
  [{ url: '/configuration/changesets' }, ...(params ? [params] : [])] as const

export type GetConfigurationChangesetsSuspenseQueryKey = ReturnType<typeof getConfigurationChangesetsSuspenseQueryKey>

/**
 * @description Get next changesets
 * @summary Get next changesets
 * {@link /configuration/changesets}
 */
export async function getConfigurationChangesetsSuspense(
  params: GetConfigurationChangesetsQueryParams,
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetConfigurationChangesetsQueryResponse,
    ResponseErrorConfig<GetConfigurationChangesets400 | GetConfigurationChangesets404 | GetConfigurationChangesets500>,
    unknown
  >({ method: 'GET', url: `/configuration/changesets`, params, ...requestConfig })
  return res.data
}

export function getConfigurationChangesetsSuspenseQueryOptions(
  params: GetConfigurationChangesetsQueryParams,
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getConfigurationChangesetsSuspenseQueryKey(params)
  return queryOptions<
    GetConfigurationChangesetsQueryResponse,
    ResponseErrorConfig<GetConfigurationChangesets400 | GetConfigurationChangesets404 | GetConfigurationChangesets500>,
    GetConfigurationChangesetsQueryResponse,
    typeof queryKey
  >({
    enabled: !!params,
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getConfigurationChangesetsSuspense(params, config)
    },
  })
}

/**
 * @description Get next changesets
 * @summary Get next changesets
 * {@link /configuration/changesets}
 */
export function useGetConfigurationChangesetsSuspense<
  TData = GetConfigurationChangesetsQueryResponse,
  TQueryKey extends QueryKey = GetConfigurationChangesetsSuspenseQueryKey,
>(
  params: GetConfigurationChangesetsQueryParams,
  options: {
    query?: Partial<
      UseSuspenseQueryOptions<
        GetConfigurationChangesetsQueryResponse,
        ResponseErrorConfig<GetConfigurationChangesets400 | GetConfigurationChangesets404 | GetConfigurationChangesets500>,
        TData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getConfigurationChangesetsSuspenseQueryKey(params)

  const query = useSuspenseQuery(
    {
      ...(getConfigurationChangesetsSuspenseQueryOptions(params, config) as unknown as UseSuspenseQueryOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<UseSuspenseQueryOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseSuspenseQueryResult<TData, ResponseErrorConfig<GetConfigurationChangesets400 | GetConfigurationChangesets404 | GetConfigurationChangesets500>> & {
    queryKey: TQueryKey
  }

  query.queryKey = queryKey as TQueryKey

  return query
}