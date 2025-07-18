/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import type { EchoHTTPError } from './echo/HTTPError.ts'

export type DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdPathParams = {
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
  /**
   * @description Key ID
   * @type integer
   */
  key_id: number
}

/**
 * @description No Content
 */
export type DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId204 = any

/**
 * @description Bad Request
 */
export type DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId400 = EchoHTTPError

/**
 * @description Unauthorized
 */
export type DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId401 = EchoHTTPError

/**
 * @description Forbidden
 */
export type DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId403 = EchoHTTPError

/**
 * @description Not Found
 */
export type DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId404 = EchoHTTPError

/**
 * @description Internal Server Error
 */
export type DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId500 = EchoHTTPError

export type DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdMutationResponse = DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId204

export type DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdMutation = {
  Response: DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId204
  PathParams: DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdPathParams
  Errors:
    | DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId400
    | DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId401
    | DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId403
    | DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId404
    | DeleteServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId500
}