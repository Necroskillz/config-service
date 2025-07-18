/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, UseSuspenseQueryOptions, UseSuspenseQueryResult } from '@tanstack/react-query'
import type {
  GetServicesServiceIdVersionsQueryResponse,
  GetServicesServiceIdVersionsPathParams,
  GetServicesServiceIdVersions400,
  GetServicesServiceIdVersions401,
  GetServicesServiceIdVersions404,
  GetServicesServiceIdVersions500,
} from '../types/GetServicesServiceIdVersions.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useSuspenseQuery } from '@tanstack/react-query'

export const getServicesServiceIdVersionsSuspenseQueryKey = (service_id: GetServicesServiceIdVersionsPathParams['service_id']) =>
  [{ url: '/services/:service_id/versions', params: { service_id: service_id } }] as const

export type GetServicesServiceIdVersionsSuspenseQueryKey = ReturnType<typeof getServicesServiceIdVersionsSuspenseQueryKey>

/**
 * @description Get service versions
 * @summary Get service versions
 * {@link /services/:service_id/versions}
 */
export async function getServicesServiceIdVersionsSuspense(
  service_id: GetServicesServiceIdVersionsPathParams['service_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetServicesServiceIdVersionsQueryResponse,
    ResponseErrorConfig<GetServicesServiceIdVersions400 | GetServicesServiceIdVersions401 | GetServicesServiceIdVersions404 | GetServicesServiceIdVersions500>,
    unknown
  >({ method: 'GET', url: `/services/${service_id}/versions`, ...requestConfig })
  return res.data
}

export function getServicesServiceIdVersionsSuspenseQueryOptions(
  service_id: GetServicesServiceIdVersionsPathParams['service_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getServicesServiceIdVersionsSuspenseQueryKey(service_id)
  return queryOptions<
    GetServicesServiceIdVersionsQueryResponse,
    ResponseErrorConfig<GetServicesServiceIdVersions400 | GetServicesServiceIdVersions401 | GetServicesServiceIdVersions404 | GetServicesServiceIdVersions500>,
    GetServicesServiceIdVersionsQueryResponse,
    typeof queryKey
  >({
    enabled: !!service_id,
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getServicesServiceIdVersionsSuspense(service_id, config)
    },
  })
}

/**
 * @description Get service versions
 * @summary Get service versions
 * {@link /services/:service_id/versions}
 */
export function useGetServicesServiceIdVersionsSuspense<
  TData = GetServicesServiceIdVersionsQueryResponse,
  TQueryKey extends QueryKey = GetServicesServiceIdVersionsSuspenseQueryKey,
>(
  service_id: GetServicesServiceIdVersionsPathParams['service_id'],
  options: {
    query?: Partial<
      UseSuspenseQueryOptions<
        GetServicesServiceIdVersionsQueryResponse,
        ResponseErrorConfig<
          GetServicesServiceIdVersions400 | GetServicesServiceIdVersions401 | GetServicesServiceIdVersions404 | GetServicesServiceIdVersions500
        >,
        TData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getServicesServiceIdVersionsSuspenseQueryKey(service_id)

  const query = useSuspenseQuery(
    {
      ...(getServicesServiceIdVersionsSuspenseQueryOptions(service_id, config) as unknown as UseSuspenseQueryOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<UseSuspenseQueryOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseSuspenseQueryResult<
    TData,
    ResponseErrorConfig<GetServicesServiceIdVersions400 | GetServicesServiceIdVersions401 | GetServicesServiceIdVersions404 | GetServicesServiceIdVersions500>
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}