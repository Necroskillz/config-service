/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { QueryKey, QueryClient, UseSuspenseQueryOptions, UseSuspenseQueryResult } from '@tanstack/react-query'
import type {
  GetChangeHistoryFeaturesFeatureIdVersionsQueryResponse,
  GetChangeHistoryFeaturesFeatureIdVersionsPathParams,
  GetChangeHistoryFeaturesFeatureIdVersions400,
  GetChangeHistoryFeaturesFeatureIdVersions401,
  GetChangeHistoryFeaturesFeatureIdVersions404,
  GetChangeHistoryFeaturesFeatureIdVersions500,
} from '../types/GetChangeHistoryFeaturesFeatureIdVersions.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { queryOptions, useSuspenseQuery } from '@tanstack/react-query'

export const getChangeHistoryFeaturesFeatureIdVersionsSuspenseQueryKey = (feature_id: GetChangeHistoryFeaturesFeatureIdVersionsPathParams['feature_id']) =>
  [{ url: '/change-history/features/:feature_id/versions', params: { feature_id: feature_id } }] as const

export type GetChangeHistoryFeaturesFeatureIdVersionsSuspenseQueryKey = ReturnType<typeof getChangeHistoryFeaturesFeatureIdVersionsSuspenseQueryKey>

/**
 * @description Get applied feature versions
 * @summary Get applied feature versions
 * {@link /change-history/features/:feature_id/versions}
 */
export async function getChangeHistoryFeaturesFeatureIdVersionsSuspense(
  feature_id: GetChangeHistoryFeaturesFeatureIdVersionsPathParams['feature_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    GetChangeHistoryFeaturesFeatureIdVersionsQueryResponse,
    ResponseErrorConfig<
      | GetChangeHistoryFeaturesFeatureIdVersions400
      | GetChangeHistoryFeaturesFeatureIdVersions401
      | GetChangeHistoryFeaturesFeatureIdVersions404
      | GetChangeHistoryFeaturesFeatureIdVersions500
    >,
    unknown
  >({ method: 'GET', url: `/change-history/features/${feature_id}/versions`, ...requestConfig })
  return res.data
}

export function getChangeHistoryFeaturesFeatureIdVersionsSuspenseQueryOptions(
  feature_id: GetChangeHistoryFeaturesFeatureIdVersionsPathParams['feature_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const queryKey = getChangeHistoryFeaturesFeatureIdVersionsSuspenseQueryKey(feature_id)
  return queryOptions<
    GetChangeHistoryFeaturesFeatureIdVersionsQueryResponse,
    ResponseErrorConfig<
      | GetChangeHistoryFeaturesFeatureIdVersions400
      | GetChangeHistoryFeaturesFeatureIdVersions401
      | GetChangeHistoryFeaturesFeatureIdVersions404
      | GetChangeHistoryFeaturesFeatureIdVersions500
    >,
    GetChangeHistoryFeaturesFeatureIdVersionsQueryResponse,
    typeof queryKey
  >({
    enabled: !!feature_id,
    queryKey,
    queryFn: async ({ signal }) => {
      config.signal = signal
      return getChangeHistoryFeaturesFeatureIdVersionsSuspense(feature_id, config)
    },
  })
}

/**
 * @description Get applied feature versions
 * @summary Get applied feature versions
 * {@link /change-history/features/:feature_id/versions}
 */
export function useGetChangeHistoryFeaturesFeatureIdVersionsSuspense<
  TData = GetChangeHistoryFeaturesFeatureIdVersionsQueryResponse,
  TQueryKey extends QueryKey = GetChangeHistoryFeaturesFeatureIdVersionsSuspenseQueryKey,
>(
  feature_id: GetChangeHistoryFeaturesFeatureIdVersionsPathParams['feature_id'],
  options: {
    query?: Partial<
      UseSuspenseQueryOptions<
        GetChangeHistoryFeaturesFeatureIdVersionsQueryResponse,
        ResponseErrorConfig<
          | GetChangeHistoryFeaturesFeatureIdVersions400
          | GetChangeHistoryFeaturesFeatureIdVersions401
          | GetChangeHistoryFeaturesFeatureIdVersions404
          | GetChangeHistoryFeaturesFeatureIdVersions500
        >,
        TData,
        TQueryKey
      >
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { query: { client: queryClient, ...queryOptions } = {}, client: config = {} } = options ?? {}
  const queryKey = queryOptions?.queryKey ?? getChangeHistoryFeaturesFeatureIdVersionsSuspenseQueryKey(feature_id)

  const query = useSuspenseQuery(
    {
      ...(getChangeHistoryFeaturesFeatureIdVersionsSuspenseQueryOptions(feature_id, config) as unknown as UseSuspenseQueryOptions),
      queryKey,
      ...(queryOptions as unknown as Omit<UseSuspenseQueryOptions, 'queryKey'>),
    },
    queryClient,
  ) as UseSuspenseQueryResult<
    TData,
    ResponseErrorConfig<
      | GetChangeHistoryFeaturesFeatureIdVersions400
      | GetChangeHistoryFeaturesFeatureIdVersions401
      | GetChangeHistoryFeaturesFeatureIdVersions404
      | GetChangeHistoryFeaturesFeatureIdVersions500
    >
  > & { queryKey: TQueryKey }

  query.queryKey = queryKey as TQueryKey

  return query
}