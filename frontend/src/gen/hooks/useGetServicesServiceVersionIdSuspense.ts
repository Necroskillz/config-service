/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, UseSuspenseQueryOptions, UseSuspenseQueryResult } from '@tanstack/react-query'
import type {
  GetServicesServiceVersionIdQueryResponse,
  GetServicesServiceVersionIdPathParams,
  GetServicesServiceVersionId400,
  GetServicesServiceVersionId401,
  GetServicesServiceVersionId404,
  GetServicesServiceVersionId500,
} from '../types/GetServicesServiceVersionId.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useSuspenseQuery } from '@tanstack/react-query'

export const getServicesServiceVersionIdSuspenseQueryKey = (service_version_id: GetServicesServiceVersionIdPathParams['service_version_id']) =>
  [{ url: '/services/:service_version_id', params: { service_version_id: service_version_id } }] as const

export type GetServicesServiceVersionIdSuspenseQueryKey = ReturnType<typeof getServicesServiceVersionIdSuspenseQueryKey>

/**
 * @description Get service
 * @summary Get service
 * {@link /services/:service_version_id}
 */
export async function getServicesServiceVersionIdSuspense(
  service_version_id: GetServicesServiceVersionIdPathParams['service_version_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetServicesServiceVersionIdQueryResponse,
    ResponseErrorConfig<GetServicesServiceVersionId400 | GetServicesServiceVersionId401 | GetServicesServiceVersionId404 | GetServicesServiceVersionId500>,
    unknown
  >({ method: 'GET', url: `/services/${service_version_id}`, ...requestConfig })
  return res.data
}

export function getServicesServiceVersionIdSuspenseQueryOptions(
  service_version_id: GetServicesServiceVersionIdPathParams['service_version_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getServicesServiceVersionIdSuspenseQueryKey(service_version_id)
  return queryOptions<
    GetServicesServiceVersionIdQueryResponse,
    ResponseErrorConfig<GetServicesServiceVersionId400 | GetServicesServiceVersionId401 | GetServicesServiceVersionId404 | GetServicesServiceVersionId500>,
    GetServicesServiceVersionIdQueryResponse,
    typeof queryKey
  >({
    enabled: !!service_version_id,
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getServicesServiceVersionIdSuspense(service_version_id, config)
    },
  })
}

/**
 * @description Get service
 * @summary Get service
 * {@link /services/:service_version_id}
 */
export function useGetServicesServiceVersionIdSuspense<
  TData = GetServicesServiceVersionIdQueryResponse,
  TQueryKey extends QueryKey = GetServicesServiceVersionIdSuspenseQueryKey,
>(
  service_version_id: GetServicesServiceVersionIdPathParams['service_version_id'],
  options: {
    query?: Partial<
      UseSuspenseQueryOptions<
        GetServicesServiceVersionIdQueryResponse,
        ResponseErrorConfig<GetServicesServiceVersionId400 | GetServicesServiceVersionId401 | GetServicesServiceVersionId404 | GetServicesServiceVersionId500>,
        TData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getServicesServiceVersionIdSuspenseQueryKey(service_version_id)

  const query = useSuspenseQuery(
    {
      ...(getServicesServiceVersionIdSuspenseQueryOptions(service_version_id, config) as unknown as UseSuspenseQueryOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<UseSuspenseQueryOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseSuspenseQueryResult<
    TData,
    ResponseErrorConfig<GetServicesServiceVersionId400 | GetServicesServiceVersionId401 | GetServicesServiceVersionId404 | GetServicesServiceVersionId500>
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}