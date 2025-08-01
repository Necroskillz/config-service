/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import type { EchoHTTPError } from './echo/HTTPError.ts'
import type { HandlerUpdateKeyRequest } from './handler/UpdateKeyRequest.ts'

export type PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdPathParams = {
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
export type PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId204 = any

/**
 * @description Bad Request
 */
export type PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId400 = EchoHTTPError

/**
 * @description Unauthorized
 */
export type PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId401 = EchoHTTPError

/**
 * @description Forbidden
 */
export type PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId403 = EchoHTTPError

/**
 * @description Not Found
 */
export type PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId404 = EchoHTTPError

/**
 * @description Internal Server Error
 */
export type PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId500 = EchoHTTPError

/**
 * @description Update key request
 */
export type PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdMutationRequest = HandlerUpdateKeyRequest

export type PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdMutationResponse = PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId204

export type PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdMutation = {
  Response: PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId204
  Request: PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdMutationRequest
  PathParams: PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyIdPathParams
  Errors:
    | PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId400
    | PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId401
    | PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId403
    | PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId404
    | PutServicesServiceVersionIdFeaturesFeatureVersionIdKeysKeyId500
}