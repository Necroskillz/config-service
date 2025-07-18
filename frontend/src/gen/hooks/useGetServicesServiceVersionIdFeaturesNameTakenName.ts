/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, QueryObserverOptions, UseQueryResult } from '@tanstack/react-query'
import type {
  GetServicesServiceVersionIdFeaturesNameTakenNameQueryResponse,
  GetServicesServiceVersionIdFeaturesNameTakenNamePathParams,
  GetServicesServiceVersionIdFeaturesNameTakenName400,
  GetServicesServiceVersionIdFeaturesNameTakenName401,
  GetServicesServiceVersionIdFeaturesNameTakenName500,
} from '../types/GetServicesServiceVersionIdFeaturesNameTakenName.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useQuery } from '@tanstack/react-query'

export const getServicesServiceVersionIdFeaturesNameTakenNameQueryKey = (
  service_version_id: GetServicesServiceVersionIdFeaturesNameTakenNamePathParams['service_version_id'],
  name: GetServicesServiceVersionIdFeaturesNameTakenNamePathParams['name'],
) => [{ url: '/services/:service_version_id/features/name-taken/:name', params: { service_version_id: service_version_id, name: name } }] as const

export type GetServicesServiceVersionIdFeaturesNameTakenNameQueryKey = ReturnType<typeof getServicesServiceVersionIdFeaturesNameTakenNameQueryKey>

/**
 * @description Check if feature name is taken
 * @summary Check if feature name is taken
 * {@link /services/:service_version_id/features/name-taken/:name}
 */
export async function getServicesServiceVersionIdFeaturesNameTakenName(
  service_version_id: GetServicesServiceVersionIdFeaturesNameTakenNamePathParams['service_version_id'],
  name: GetServicesServiceVersionIdFeaturesNameTakenNamePathParams['name'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetServicesServiceVersionIdFeaturesNameTakenNameQueryResponse,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdFeaturesNameTakenName400
      | GetServicesServiceVersionIdFeaturesNameTakenName401
      | GetServicesServiceVersionIdFeaturesNameTakenName500
    >,
    unknown
  >({ method: 'GET', url: `/services/${service_version_id}/features/name-taken/${name}`, ...requestConfig })
  return res.data
}

export function getServicesServiceVersionIdFeaturesNameTakenNameQueryOptions(
  service_version_id: GetServicesServiceVersionIdFeaturesNameTakenNamePathParams['service_version_id'],
  name: GetServicesServiceVersionIdFeaturesNameTakenNamePathParams['name'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getServicesServiceVersionIdFeaturesNameTakenNameQueryKey(service_version_id, name)
  return queryOptions<
    GetServicesServiceVersionIdFeaturesNameTakenNameQueryResponse,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdFeaturesNameTakenName400
      | GetServicesServiceVersionIdFeaturesNameTakenName401
      | GetServicesServiceVersionIdFeaturesNameTakenName500
    >,
    GetServicesServiceVersionIdFeaturesNameTakenNameQueryResponse,
    typeof queryKey
  >({
    enabled: !!(service_version_id && name),
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getServicesServiceVersionIdFeaturesNameTakenName(service_version_id, name, config)
    },
  })
}

/**
 * @description Check if feature name is taken
 * @summary Check if feature name is taken
 * {@link /services/:service_version_id/features/name-taken/:name}
 */
export function useGetServicesServiceVersionIdFeaturesNameTakenName<
  TData = GetServicesServiceVersionIdFeaturesNameTakenNameQueryResponse,
  TQueryData = GetServicesServiceVersionIdFeaturesNameTakenNameQueryResponse,
  TQueryKey extends QueryKey = GetServicesServiceVersionIdFeaturesNameTakenNameQueryKey,
>(
  service_version_id: GetServicesServiceVersionIdFeaturesNameTakenNamePathParams['service_version_id'],
  name: GetServicesServiceVersionIdFeaturesNameTakenNamePathParams['name'],
  options: {
    query?: Partial<
      QueryObserverOptions<
        GetServicesServiceVersionIdFeaturesNameTakenNameQueryResponse,
        ResponseErrorConfig<
          | GetServicesServiceVersionIdFeaturesNameTakenName400
          | GetServicesServiceVersionIdFeaturesNameTakenName401
          | GetServicesServiceVersionIdFeaturesNameTakenName500
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
  const queryKey = queryOptions?.queryKey ?? getServicesServiceVersionIdFeaturesNameTakenNameQueryKey(service_version_id, name)

  const query = useQuery(
    {
      ...(getServicesServiceVersionIdFeaturesNameTakenNameQueryOptions(service_version_id, name, config) as unknown as QueryObserverOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<QueryObserverOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseQueryResult<
    TData,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdFeaturesNameTakenName400
      | GetServicesServiceVersionIdFeaturesNameTakenName401
      | GetServicesServiceVersionIdFeaturesNameTakenName500
    >
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}