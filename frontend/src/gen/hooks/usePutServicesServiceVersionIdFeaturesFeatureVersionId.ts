/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { UseMutationOptions, QueryClient } from '@tanstack/react-query'
import type {
  PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationRequest,
  PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationResponse,
  PutServicesServiceVersionIdFeaturesFeatureVersionIdPathParams,
  PutServicesServiceVersionIdFeaturesFeatureVersionId400,
  PutServicesServiceVersionIdFeaturesFeatureVersionId401,
  PutServicesServiceVersionIdFeaturesFeatureVersionId403,
  PutServicesServiceVersionIdFeaturesFeatureVersionId422,
  PutServicesServiceVersionIdFeaturesFeatureVersionId500,
} from '../types/PutServicesServiceVersionIdFeaturesFeatureVersionId.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { useMutation } from '@tanstack/react-query'

export const putServicesServiceVersionIdFeaturesFeatureVersionIdMutationKey = () =>
  [{ url: '/services/{service_version_id}/features/{feature_version_id}' }] as const

export type PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationKey = ReturnType<typeof putServicesServiceVersionIdFeaturesFeatureVersionIdMutationKey>

/**
 * @description Create feature
 * @summary Create feature
 * {@link /services/:service_version_id/features/:feature_version_id}
 */
export async function putServicesServiceVersionIdFeaturesFeatureVersionId(
  service_version_id: PutServicesServiceVersionIdFeaturesFeatureVersionIdPathParams['service_version_id'],
  feature_version_id: PutServicesServiceVersionIdFeaturesFeatureVersionIdPathParams['feature_version_id'],
  data: PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationRequest,
  config: Partial<RequestConfig<PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationRequest>> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationResponse,
    ResponseErrorConfig<
      | PutServicesServiceVersionIdFeaturesFeatureVersionId400
      | PutServicesServiceVersionIdFeaturesFeatureVersionId401
      | PutServicesServiceVersionIdFeaturesFeatureVersionId403
      | PutServicesServiceVersionIdFeaturesFeatureVersionId422
      | PutServicesServiceVersionIdFeaturesFeatureVersionId500
    >,
    PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationRequest
  >({ method: 'PUT', url: `/services/${service_version_id}/features/${feature_version_id}`, data, ...requestConfig })
  return res.data
}

/**
 * @description Create feature
 * @summary Create feature
 * {@link /services/:service_version_id/features/:feature_version_id}
 */
export function usePutServicesServiceVersionIdFeaturesFeatureVersionId<TContext>(
  options: {
    mutation?: UseMutationOptions<
      PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationResponse,
      ResponseErrorConfig<
        | PutServicesServiceVersionIdFeaturesFeatureVersionId400
        | PutServicesServiceVersionIdFeaturesFeatureVersionId401
        | PutServicesServiceVersionIdFeaturesFeatureVersionId403
        | PutServicesServiceVersionIdFeaturesFeatureVersionId422
        | PutServicesServiceVersionIdFeaturesFeatureVersionId500
      >,
      {
        service_version_id: PutServicesServiceVersionIdFeaturesFeatureVersionIdPathParams['service_version_id']
        feature_version_id: PutServicesServiceVersionIdFeaturesFeatureVersionIdPathParams['feature_version_id']
        data: PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationRequest
      },
      TContext
    > & { client?: QueryClient }
    client?: Partial<RequestConfig<PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationRequest>> & { client?: typeof client }
  } = {},
) {
  const { mutation = {}, client: config = {} } = options ?? {}
  const { client: queryClient, ...mutationOptions } = mutation
  const mutationKey = mutationOptions.mutationKey ?? putServicesServiceVersionIdFeaturesFeatureVersionIdMutationKey()

  return useMutation<
    PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationResponse,
    ResponseErrorConfig<
      | PutServicesServiceVersionIdFeaturesFeatureVersionId400
      | PutServicesServiceVersionIdFeaturesFeatureVersionId401
      | PutServicesServiceVersionIdFeaturesFeatureVersionId403
      | PutServicesServiceVersionIdFeaturesFeatureVersionId422
      | PutServicesServiceVersionIdFeaturesFeatureVersionId500
    >,
    {
      service_version_id: PutServicesServiceVersionIdFeaturesFeatureVersionIdPathParams['service_version_id']
      feature_version_id: PutServicesServiceVersionIdFeaturesFeatureVersionIdPathParams['feature_version_id']
      data: PutServicesServiceVersionIdFeaturesFeatureVersionIdMutationRequest
    },
    TContext
  >(
    {
      mutationFn: async ({ service_version_id, feature_version_id, data }) => {
        return putServicesServiceVersionIdFeaturesFeatureVersionId(service_version_id, feature_version_id, data, config)
      },
      mutationKey,
      ...mutationOptions,
    },
    queryClient,
  )
}