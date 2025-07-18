/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { UseMutationOptions, QueryClient } from '@tanstack/react-query'
import type {
  PostServicesServiceVersionIdFeaturesMutationRequest,
  PostServicesServiceVersionIdFeaturesMutationResponse,
  PostServicesServiceVersionIdFeaturesPathParams,
  PostServicesServiceVersionIdFeatures400,
  PostServicesServiceVersionIdFeatures401,
  PostServicesServiceVersionIdFeatures403,
  PostServicesServiceVersionIdFeatures422,
  PostServicesServiceVersionIdFeatures500,
} from '../types/PostServicesServiceVersionIdFeatures.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { useMutation } from '@tanstack/react-query'

export const postServicesServiceVersionIdFeaturesMutationKey = () => [{ url: '/services/{service_version_id}/features' }] as const

export type PostServicesServiceVersionIdFeaturesMutationKey = ReturnType<typeof postServicesServiceVersionIdFeaturesMutationKey>

/**
 * @description Create feature
 * @summary Create feature
 * {@link /services/:service_version_id/features}
 */
export async function postServicesServiceVersionIdFeatures(
  service_version_id: PostServicesServiceVersionIdFeaturesPathParams['service_version_id'],
  data: PostServicesServiceVersionIdFeaturesMutationRequest,
  config: Partial<RequestConfig<PostServicesServiceVersionIdFeaturesMutationRequest>> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    PostServicesServiceVersionIdFeaturesMutationResponse,
    ResponseErrorConfig<
      | PostServicesServiceVersionIdFeatures400
      | PostServicesServiceVersionIdFeatures401
      | PostServicesServiceVersionIdFeatures403
      | PostServicesServiceVersionIdFeatures422
      | PostServicesServiceVersionIdFeatures500
    >,
    PostServicesServiceVersionIdFeaturesMutationRequest
  >({ method: 'POST', url: `/services/${service_version_id}/features`, data, ...requestConfig })
  return res.data
}

/**
 * @description Create feature
 * @summary Create feature
 * {@link /services/:service_version_id/features}
 */
export function usePostServicesServiceVersionIdFeatures<TContext>(
  options: {
    mutation?: UseMutationOptions<
      PostServicesServiceVersionIdFeaturesMutationResponse,
      ResponseErrorConfig<
        | PostServicesServiceVersionIdFeatures400
        | PostServicesServiceVersionIdFeatures401
        | PostServicesServiceVersionIdFeatures403
        | PostServicesServiceVersionIdFeatures422
        | PostServicesServiceVersionIdFeatures500
      >,
      { service_version_id: PostServicesServiceVersionIdFeaturesPathParams['service_version_id']; data: PostServicesServiceVersionIdFeaturesMutationRequest },
      TContext
    > & { client?: QueryClient }
    client?: Partial<RequestConfig<PostServicesServiceVersionIdFeaturesMutationRequest>> & { client?: typeof client }
  } = {},
) {
  const { mutation = {}, client: config = {} } = options ?? {}
  const { client: queryClient, ...mutationOptions } = mutation
  const mutationKey = mutationOptions.mutationKey ?? postServicesServiceVersionIdFeaturesMutationKey()

  return useMutation<
    PostServicesServiceVersionIdFeaturesMutationResponse,
    ResponseErrorConfig<
      | PostServicesServiceVersionIdFeatures400
      | PostServicesServiceVersionIdFeatures401
      | PostServicesServiceVersionIdFeatures403
      | PostServicesServiceVersionIdFeatures422
      | PostServicesServiceVersionIdFeatures500
    >,
    { service_version_id: PostServicesServiceVersionIdFeaturesPathParams['service_version_id']; data: PostServicesServiceVersionIdFeaturesMutationRequest },
    TContext
  >(
    {
      mutationFn: async ({ service_version_id, data }) => {
        return postServicesServiceVersionIdFeatures(service_version_id, data, config)
      },
      mutationKey,
      ...mutationOptions,
    },
    queryClient,
  )
}