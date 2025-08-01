/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, QueryObserverOptions, UseQueryResult } from '@tanstack/react-query'
import type {
  GetServicesServiceVersionIdVersionsQueryResponse,
  GetServicesServiceVersionIdVersionsPathParams,
  GetServicesServiceVersionIdVersions400,
  GetServicesServiceVersionIdVersions401,
  GetServicesServiceVersionIdVersions404,
  GetServicesServiceVersionIdVersions500,
} from '../types/GetServicesServiceVersionIdVersions.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useQuery } from '@tanstack/react-query'

export const getServicesServiceVersionIdVersionsQueryKey = (service_version_id: GetServicesServiceVersionIdVersionsPathParams['service_version_id']) =>
  [{ url: '/services/:service_version_id/versions', params: { service_version_id: service_version_id } }] as const

export type GetServicesServiceVersionIdVersionsQueryKey = ReturnType<typeof getServicesServiceVersionIdVersionsQueryKey>

/**
 * @description Get service versions
 * @summary Get service versions
 * {@link /services/:service_version_id/versions}
 */
export async function getServicesServiceVersionIdVersions(
  service_version_id: GetServicesServiceVersionIdVersionsPathParams['service_version_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetServicesServiceVersionIdVersionsQueryResponse,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdVersions400
      | GetServicesServiceVersionIdVersions401
      | GetServicesServiceVersionIdVersions404
      | GetServicesServiceVersionIdVersions500
    >,
    unknown
  >({ method: 'GET', url: `/services/${service_version_id}/versions`, ...requestConfig })
  return res.data
}

export function getServicesServiceVersionIdVersionsQueryOptions(
  service_version_id: GetServicesServiceVersionIdVersionsPathParams['service_version_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getServicesServiceVersionIdVersionsQueryKey(service_version_id)
  return queryOptions<
    GetServicesServiceVersionIdVersionsQueryResponse,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdVersions400
      | GetServicesServiceVersionIdVersions401
      | GetServicesServiceVersionIdVersions404
      | GetServicesServiceVersionIdVersions500
    >,
    GetServicesServiceVersionIdVersionsQueryResponse,
    typeof queryKey
  >({
    enabled: !!service_version_id,
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getServicesServiceVersionIdVersions(service_version_id, config)
    },
  })
}

/**
 * @description Get service versions
 * @summary Get service versions
 * {@link /services/:service_version_id/versions}
 */
export function useGetServicesServiceVersionIdVersions<
  TData = GetServicesServiceVersionIdVersionsQueryResponse,
  TQueryData = GetServicesServiceVersionIdVersionsQueryResponse,
  TQueryKey extends QueryKey = GetServicesServiceVersionIdVersionsQueryKey,
>(
  service_version_id: GetServicesServiceVersionIdVersionsPathParams['service_version_id'],
  options: {
    query?: Partial<
      QueryObserverOptions<
        GetServicesServiceVersionIdVersionsQueryResponse,
        ResponseErrorConfig<
          | GetServicesServiceVersionIdVersions400
          | GetServicesServiceVersionIdVersions401
          | GetServicesServiceVersionIdVersions404
          | GetServicesServiceVersionIdVersions500
        >,
        TData,
        TQueryData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getServicesServiceVersionIdVersionsQueryKey(service_version_id)

  const query = useQuery(
    {
      ...(getServicesServiceVersionIdVersionsQueryOptions(service_version_id, config) as unknown as QueryObserverOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<QueryObserverOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseQueryResult<
    TData,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdVersions400
      | GetServicesServiceVersionIdVersions401
      | GetServicesServiceVersionIdVersions404
      | GetServicesServiceVersionIdVersions500
    >
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}