/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import type { EchoHTTPError } from './echo/HTTPError.ts'
import type { VariationpropertyServiceTypeVariationPropertyDto } from './variationproperty/ServiceTypeVariationPropertyDto.ts'

export type GetServiceTypesServiceTypeIdVariationPropertiesPathParams = {
  /**
   * @description Service type ID
   * @type integer
   */
  service_type_id: number
}

/**
 * @description OK
 */
export type GetServiceTypesServiceTypeIdVariationProperties200 = VariationpropertyServiceTypeVariationPropertyDto[]

/**
 * @description Bad Request
 */
export type GetServiceTypesServiceTypeIdVariationProperties400 = EchoHTTPError

/**
 * @description Unauthorized
 */
export type GetServiceTypesServiceTypeIdVariationProperties401 = EchoHTTPError

/**
 * @description Not Found
 */
export type GetServiceTypesServiceTypeIdVariationProperties404 = EchoHTTPError

/**
 * @description Internal Server Error
 */
export type GetServiceTypesServiceTypeIdVariationProperties500 = EchoHTTPError

export type GetServiceTypesServiceTypeIdVariationPropertiesQueryResponse = GetServiceTypesServiceTypeIdVariationProperties200

export type GetServiceTypesServiceTypeIdVariationPropertiesQuery = {
  Response: GetServiceTypesServiceTypeIdVariationProperties200
  PathParams: GetServiceTypesServiceTypeIdVariationPropertiesPathParams
  Errors:
    | GetServiceTypesServiceTypeIdVariationProperties400
    | GetServiceTypesServiceTypeIdVariationProperties401
    | GetServiceTypesServiceTypeIdVariationProperties404
    | GetServiceTypesServiceTypeIdVariationProperties500
}