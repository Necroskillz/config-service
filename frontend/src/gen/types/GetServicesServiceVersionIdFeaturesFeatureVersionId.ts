/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import type { EchoHTTPError } from './echo/HTTPError.ts'
import type { FeatureFeatureVersionDto } from './feature/FeatureVersionDto.ts'

export type GetServicesServiceVersionIdFeaturesFeatureVersionIdPathParams = {
  /**
   * @description Service version ID
   * @type integer
   */
  service_version_id: number
  /**
   * @description Feature version ID
   * @type integer
   */
  feature_version_id: number
}

/**
 * @description OK
 */
export type GetServicesServiceVersionIdFeaturesFeatureVersionId200 = FeatureFeatureVersionDto

/**
 * @description Bad Request
 */
export type GetServicesServiceVersionIdFeaturesFeatureVersionId400 = EchoHTTPError

/**
 * @description Unauthorized
 */
export type GetServicesServiceVersionIdFeaturesFeatureVersionId401 = EchoHTTPError

/**
 * @description Not Found
 */
export type GetServicesServiceVersionIdFeaturesFeatureVersionId404 = EchoHTTPError

/**
 * @description Internal Server Error
 */
export type GetServicesServiceVersionIdFeaturesFeatureVersionId500 = EchoHTTPError

export type GetServicesServiceVersionIdFeaturesFeatureVersionIdQueryResponse = GetServicesServiceVersionIdFeaturesFeatureVersionId200

export type GetServicesServiceVersionIdFeaturesFeatureVersionIdQuery = {
  Response: GetServicesServiceVersionIdFeaturesFeatureVersionId200
  PathParams: GetServicesServiceVersionIdFeaturesFeatureVersionIdPathParams
  Errors:
    | GetServicesServiceVersionIdFeaturesFeatureVersionId400
    | GetServicesServiceVersionIdFeaturesFeatureVersionId401
    | GetServicesServiceVersionIdFeaturesFeatureVersionId404
    | GetServicesServiceVersionIdFeaturesFeatureVersionId500
}