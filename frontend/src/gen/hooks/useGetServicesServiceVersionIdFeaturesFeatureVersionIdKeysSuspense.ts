/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, UseSuspenseQueryOptions, UseSuspenseQueryResult } from '@tanstack/react-query'
import type {
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysQueryResponse,
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysPathParams,
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys400,
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys401,
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys404,
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys500,
} from '../types/GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useSuspenseQuery } from '@tanstack/react-query'

export const getServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspenseQueryKey = (
  service_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysPathParams['service_version_id'],
  feature_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysPathParams['feature_version_id'],
) =>
  [
    {
      url: '/services/:service_version_id/features/:feature_version_id/keys',
      params: { service_version_id: service_version_id, feature_version_id: feature_version_id },
    },
  ] as const

export type GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspenseQueryKey = ReturnType<
  typeof getServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspenseQueryKey
>

/**
 * @description Get keys for a feature
 * @summary Get keys for a feature
 * {@link /services/:service_version_id/features/:feature_version_id/keys}
 */
export async function getServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspense(
  service_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysPathParams['service_version_id'],
  feature_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysPathParams['feature_version_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysQueryResponse,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys400
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys401
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys404
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys500
    >,
    unknown
  >({ method: 'GET', url: `/services/${service_version_id}/features/${feature_version_id}/keys`, ...requestConfig })
  return res.data
}

export function getServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspenseQueryOptions(
  service_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysPathParams['service_version_id'],
  feature_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysPathParams['feature_version_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspenseQueryKey(service_version_id, feature_version_id)
  return queryOptions<
    GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysQueryResponse,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys400
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys401
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys404
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys500
    >,
    GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysQueryResponse,
    typeof queryKey
  >({
    enabled: !!(service_version_id && feature_version_id),
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspense(service_version_id, feature_version_id, config)
    },
  })
}

/**
 * @description Get keys for a feature
 * @summary Get keys for a feature
 * {@link /services/:service_version_id/features/:feature_version_id/keys}
 */
export function useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspense<
  TData = GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysQueryResponse,
  TQueryKey extends QueryKey = GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspenseQueryKey,
>(
  service_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysPathParams['service_version_id'],
  feature_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysPathParams['feature_version_id'],
  options: {
    query?: Partial<
      UseSuspenseQueryOptions<
        GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysQueryResponse,
        ResponseErrorConfig<
          | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys400
          | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys401
          | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys404
          | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys500
        >,
        TData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspenseQueryKey(service_version_id, feature_version_id)

  const query = useSuspenseQuery(
    {
      ...(getServicesServiceVersionIdFeaturesFeatureVersionIdKeysSuspenseQueryOptions(
        service_version_id,
        feature_version_id,
        config,
      ) as unknown as UseSuspenseQueryOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<UseSuspenseQueryOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseSuspenseQueryResult<
    TData,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys400
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys401
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys404
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeys500
    >
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}