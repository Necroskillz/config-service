/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, UseSuspenseQueryOptions, UseSuspenseQueryResult } from '@tanstack/react-query'
import type {
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryResponse,
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams,
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues400,
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues401,
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues404,
  GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues500,
} from '../types/GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useSuspenseQuery } from '@tanstack/react-query'

export const getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspenseQueryKey = (
  service_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['service_version_id'],
  feature_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['feature_version_id'],
  key_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['key_id'],
) =>
  [
    {
      url: '/services/:service_version_id/features/:feature_version_id/keys/:key_id/values',
      params: { service_version_id: service_version_id, feature_version_id: feature_version_id, key_id: key_id },
    },
  ] as const

export type GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspenseQueryKey = ReturnType<
  typeof getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspenseQueryKey
>

/**
 * @description Get values for a key
 * @summary Get values for a key
 * {@link /services/:service_version_id/features/:feature_version_id/keys/:key_id/values}
 */
export async function getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspense(
  service_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['service_version_id'],
  feature_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['feature_version_id'],
  key_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['key_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryResponse,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues400
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues401
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues404
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues500
    >,
    unknown
  >({ method: 'GET', url: `/services/${service_version_id}/features/${feature_version_id}/keys/${key_id}/values`, ...requestConfig })
  return res.data
}

export function getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspenseQueryOptions(
  service_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['service_version_id'],
  feature_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['feature_version_id'],
  key_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['key_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspenseQueryKey(service_version_id, feature_version_id, key_id)
  return queryOptions<
    GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryResponse,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues400
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues401
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues404
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues500
    >,
    GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryResponse,
    typeof queryKey
  >({
    enabled: !!(service_version_id && feature_version_id && key_id),
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspense(service_version_id, feature_version_id, key_id, config)
    },
  })
}

/**
 * @description Get values for a key
 * @summary Get values for a key
 * {@link /services/:service_version_id/features/:feature_version_id/keys/:key_id/values}
 */
export function useGetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspense<
  TData = GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryResponse,
  TQueryKey extends QueryKey = GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspenseQueryKey,
>(
  service_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['service_version_id'],
  feature_version_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['feature_version_id'],
  key_id: GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesPathParams['key_id'],
  options: {
    query?: Partial<
      UseSuspenseQueryOptions<
        GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesQueryResponse,
        ResponseErrorConfig<
          | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues400
          | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues401
          | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues404
          | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues500
        >,
        TData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey =
    queryOptions?.queryKey ?? getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspenseQueryKey(service_version_id, feature_version_id, key_id)

  const query = useSuspenseQuery(
    {
      ...(getServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValuesSuspenseQueryOptions(
        service_version_id,
        feature_version_id,
        key_id,
        config,
      ) as unknown as UseSuspenseQueryOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<UseSuspenseQueryOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseSuspenseQueryResult<
    TData,
    ResponseErrorConfig<
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues400
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues401
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues404
      | GetServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdValues500
    >
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}