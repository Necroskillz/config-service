/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import type { EchoHTTPError } from './echo/HTTPError.ts'

export type PutChangesetsChangesetIdReopenPathParams = {
  /**
   * @description Changeset ID
   * @type integer
   */
  changeset_id: number
}

/**
 * @description No Content
 */
export type PutChangesetsChangesetIdReopen204 = any

/**
 * @description Bad Request
 */
export type PutChangesetsChangesetIdReopen400 = EchoHTTPError

/**
 * @description Unauthorized
 */
export type PutChangesetsChangesetIdReopen401 = EchoHTTPError

/**
 * @description Forbidden
 */
export type PutChangesetsChangesetIdReopen403 = EchoHTTPError

/**
 * @description Not Found
 */
export type PutChangesetsChangesetIdReopen404 = EchoHTTPError

/**
 * @description Internal Server Error
 */
export type PutChangesetsChangesetIdReopen500 = EchoHTTPError

export type PutChangesetsChangesetIdReopenMutationResponse = PutChangesetsChangesetIdReopen204

export type PutChangesetsChangesetIdReopenMutation = {
  Response: PutChangesetsChangesetIdReopen204
  PathParams: PutChangesetsChangesetIdReopenPathParams
  Errors:
    | PutChangesetsChangesetIdReopen400
    | PutChangesetsChangesetIdReopen401
    | PutChangesetsChangesetIdReopen403
    | PutChangesetsChangesetIdReopen404
    | PutChangesetsChangesetIdReopen500
}