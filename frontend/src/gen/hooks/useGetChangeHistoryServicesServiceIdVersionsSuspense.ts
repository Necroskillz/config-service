/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, UseSuspenseQueryOptions, UseSuspenseQueryResult } from '@tanstack/react-query'
import type {
  GetChangeHistoryServicesServiceIdVersionsQueryResponse,
  GetChangeHistoryServicesServiceIdVersionsPathParams,
  GetChangeHistoryServicesServiceIdVersions400,
  GetChangeHistoryServicesServiceIdVersions401,
  GetChangeHistoryServicesServiceIdVersions404,
  GetChangeHistoryServicesServiceIdVersions500,
} from '../types/GetChangeHistoryServicesServiceIdVersions.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useSuspenseQuery } from '@tanstack/react-query'

export const getChangeHistoryServicesServiceIdVersionsSuspenseQueryKey = (service_id: GetChangeHistoryServicesServiceIdVersionsPathParams['service_id']) =>
  [{ url: '/change-history/services/:service_id/versions', params: { service_id: service_id } }] as const

export type GetChangeHistoryServicesServiceIdVersionsSuspenseQueryKey = ReturnType<typeof getChangeHistoryServicesServiceIdVersionsSuspenseQueryKey>

/**
 * @description Get applied service versions
 * @summary Get applied service versions
 * {@link /change-history/services/:service_id/versions}
 */
export async function getChangeHistoryServicesServiceIdVersionsSuspense(
  service_id: GetChangeHistoryServicesServiceIdVersionsPathParams['service_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetChangeHistoryServicesServiceIdVersionsQueryResponse,
    ResponseErrorConfig<
      | GetChangeHistoryServicesServiceIdVersions400
      | GetChangeHistoryServicesServiceIdVersions401
      | GetChangeHistoryServicesServiceIdVersions404
      | GetChangeHistoryServicesServiceIdVersions500
    >,
    unknown
  >({ method: 'GET', url: `/change-history/services/${service_id}/versions`, ...requestConfig })
  return res.data
}

export function getChangeHistoryServicesServiceIdVersionsSuspenseQueryOptions(
  service_id: GetChangeHistoryServicesServiceIdVersionsPathParams['service_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getChangeHistoryServicesServiceIdVersionsSuspenseQueryKey(service_id)
  return queryOptions<
    GetChangeHistoryServicesServiceIdVersionsQueryResponse,
    ResponseErrorConfig<
      | GetChangeHistoryServicesServiceIdVersions400
      | GetChangeHistoryServicesServiceIdVersions401
      | GetChangeHistoryServicesServiceIdVersions404
      | GetChangeHistoryServicesServiceIdVersions500
    >,
    GetChangeHistoryServicesServiceIdVersionsQueryResponse,
    typeof queryKey
  >({
    enabled: !!service_id,
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getChangeHistoryServicesServiceIdVersionsSuspense(service_id, config)
    },
  })
}

/**
 * @description Get applied service versions
 * @summary Get applied service versions
 * {@link /change-history/services/:service_id/versions}
 */
export function useGetChangeHistoryServicesServiceIdVersionsSuspense<
  TData = GetChangeHistoryServicesServiceIdVersionsQueryResponse,
  TQueryKey extends QueryKey = GetChangeHistoryServicesServiceIdVersionsSuspenseQueryKey,
>(
  service_id: GetChangeHistoryServicesServiceIdVersionsPathParams['service_id'],
  options: {
    query?: Partial<
      UseSuspenseQueryOptions<
        GetChangeHistoryServicesServiceIdVersionsQueryResponse,
        ResponseErrorConfig<
          | GetChangeHistoryServicesServiceIdVersions400
          | GetChangeHistoryServicesServiceIdVersions401
          | GetChangeHistoryServicesServiceIdVersions404
          | GetChangeHistoryServicesServiceIdVersions500
        >,
        TData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getChangeHistoryServicesServiceIdVersionsSuspenseQueryKey(service_id)

  const query = useSuspenseQuery(
    {
      ...(getChangeHistoryServicesServiceIdVersionsSuspenseQueryOptions(service_id, config) as unknown as UseSuspenseQueryOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<UseSuspenseQueryOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseSuspenseQueryResult<
    TData,
    ResponseErrorConfig<
      | GetChangeHistoryServicesServiceIdVersions400
      | GetChangeHistoryServicesServiceIdVersions401
      | GetChangeHistoryServicesServiceIdVersions404
      | GetChangeHistoryServicesServiceIdVersions500
    >
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}