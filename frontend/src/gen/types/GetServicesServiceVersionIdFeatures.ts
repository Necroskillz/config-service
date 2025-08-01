/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import type { EchoHTTPError } from './echo/HTTPError.ts'
import type { FeatureFeatureVersionItemDto } from './feature/FeatureVersionItemDto.ts'

export type GetServicesServiceVersionIdFeaturesPathParams = {
  /**
   * @description Service version ID
   * @type integer
   */
  service_version_id: number
}

/**
 * @description OK
 */
export type GetServicesServiceVersionIdFeatures200 = FeatureFeatureVersionItemDto[]

/**
 * @description Bad Request
 */
export type GetServicesServiceVersionIdFeatures400 = EchoHTTPError

/**
 * @description Unauthorized
 */
export type GetServicesServiceVersionIdFeatures401 = EchoHTTPError

/**
 * @description Not Found
 */
export type GetServicesServiceVersionIdFeatures404 = EchoHTTPError

/**
 * @description Internal Server Error
 */
export type GetServicesServiceVersionIdFeatures500 = EchoHTTPError

export type GetServicesServiceVersionIdFeaturesQueryResponse = GetServicesServiceVersionIdFeatures200

export type GetServicesServiceVersionIdFeaturesQuery = {
  Response: GetServicesServiceVersionIdFeatures200
  PathParams: GetServicesServiceVersionIdFeaturesPathParams
  Errors:
    | GetServicesServiceVersionIdFeatures400
    | GetServicesServiceVersionIdFeatures401
    | GetServicesServiceVersionIdFeatures404
    | GetServicesServiceVersionIdFeatures500
}