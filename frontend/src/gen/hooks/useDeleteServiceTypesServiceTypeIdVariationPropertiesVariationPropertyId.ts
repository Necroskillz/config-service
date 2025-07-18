/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { UseMutationOptions, QueryClient } from '@tanstack/react-query'
import type {
  DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdMutationResponse,
  DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdPathParams,
  DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId400,
  DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId401,
  DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId403,
  DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId404,
  DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId500,
} from '../types/DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { useMutation } from '@tanstack/react-query'

export const deleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdMutationKey = () =>
  [{ url: '/service-types/{service_type_id}/variation-properties/{variation_property_id}' }] as const

export type DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdMutationKey = ReturnType<
  typeof deleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdMutationKey
>

/**
 * @description Unlink a variation property from a service type
 * @summary Unlink variation property from service type
 * {@link /service-types/:service_type_id/variation-properties/:variation_property_id}
 */
export async function deleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId(
  service_type_id: DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdPathParams['service_type_id'],
  variation_property_id: DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdPathParams['variation_property_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdMutationResponse,
    ResponseErrorConfig<
      | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId400
      | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId401
      | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId403
      | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId404
      | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId500
    >,
    unknown
  >({ method: 'DELETE', url: `/service-types/${service_type_id}/variation-properties/${variation_property_id}`, ...requestConfig })
  return res.data
}

/**
 * @description Unlink a variation property from a service type
 * @summary Unlink variation property from service type
 * {@link /service-types/:service_type_id/variation-properties/:variation_property_id}
 */
export function useDeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId<TContext>(
  options: {
    mutation?: UseMutationOptions<
      DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdMutationResponse,
      ResponseErrorConfig<
        | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId400
        | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId401
        | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId403
        | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId404
        | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId500
      >,
      {
        service_type_id: DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdPathParams['service_type_id']
        variation_property_id: DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdPathParams['variation_property_id']
      },
      TContext
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { mutation = {}, client: config = {} } = options ?? {}
  const { client: queryClient, ...mutationOptions } = mutation
  const mutationKey = mutationOptions.mutationKey ?? deleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdMutationKey()

  return useMutation<
    DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdMutationResponse,
    ResponseErrorConfig<
      | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId400
      | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId401
      | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId403
      | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId404
      | DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId500
    >,
    {
      service_type_id: DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdPathParams['service_type_id']
      variation_property_id: DeleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyIdPathParams['variation_property_id']
    },
    TContext
  >(
    {
      mutationFn: async ({ service_type_id, variation_property_id }) => {
        return deleteServiceTypesServiceTypeIdVariationPropertiesVariationPropertyId(service_type_id, variation_property_id, config)
      },
      mutationKey,
      ...mutationOptions,
    },
    queryClient,
  )
}