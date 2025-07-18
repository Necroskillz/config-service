/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, UseSuspenseQueryOptions, UseSuspenseQueryResult } from '@tanstack/react-query'
import type {
  GetChangeHistoryServicesQueryResponse,
  GetChangeHistoryServices400,
  GetChangeHistoryServices401,
  GetChangeHistoryServices404,
  GetChangeHistoryServices500,
} from '../types/GetChangeHistoryServices.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useSuspenseQuery } from '@tanstack/react-query'

export const getChangeHistoryServicesSuspenseQueryKey = () => [{ url: '/change-history/services' }] as const

export type GetChangeHistoryServicesSuspenseQueryKey = ReturnType<typeof getChangeHistoryServicesSuspenseQueryKey>

/**
 * @description Get applied services
 * @summary Get applied services
 * {@link /change-history/services}
 */
export async function getChangeHistoryServicesSuspense(config: Partial<RequestConfig> & { client?: typeof client } = {}) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetChangeHistoryServicesQueryResponse,
    ResponseErrorConfig<GetChangeHistoryServices400 | GetChangeHistoryServices401 | GetChangeHistoryServices404 | GetChangeHistoryServices500>,
    unknown
  >({ method: 'GET', url: `/change-history/services`, ...requestConfig })
  return res.data
}

export function getChangeHistoryServicesSuspenseQueryOptions(config: Partial<RequestConfig> & { client?: typeof client } = {}) {
  const queryKey = getChangeHistoryServicesSuspenseQueryKey()
  return queryOptions<
    GetChangeHistoryServicesQueryResponse,
    ResponseErrorConfig<GetChangeHistoryServices400 | GetChangeHistoryServices401 | GetChangeHistoryServices404 | GetChangeHistoryServices500>,
    GetChangeHistoryServicesQueryResponse,
    typeof queryKey
  >({
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getChangeHistoryServicesSuspense(config)
    },
  })
}

/**
 * @description Get applied services
 * @summary Get applied services
 * {@link /change-history/services}
 */
export function useGetChangeHistoryServicesSuspense<
  TData = GetChangeHistoryServicesQueryResponse,
  TQueryKey extends QueryKey = GetChangeHistoryServicesSuspenseQueryKey,
>(
  options: {
    query?: Partial<
      UseSuspenseQueryOptions<
        GetChangeHistoryServicesQueryResponse,
        ResponseErrorConfig<GetChangeHistoryServices400 | GetChangeHistoryServices401 | GetChangeHistoryServices404 | GetChangeHistoryServices500>,
        TData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getChangeHistoryServicesSuspenseQueryKey()

  const query = useSuspenseQuery(
    {
      ...(getChangeHistoryServicesSuspenseQueryOptions(config) as unknown as UseSuspenseQueryOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<UseSuspenseQueryOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseSuspenseQueryResult<
    TData,
    ResponseErrorConfig<GetChangeHistoryServices400 | GetChangeHistoryServices401 | GetChangeHistoryServices404 | GetChangeHistoryServices500>
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}