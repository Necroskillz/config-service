/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import type { EchoHTTPError } from './echo/HTTPError.ts'
import type { HandlerCreateGroupRequest } from './handler/CreateGroupRequest.ts'
import type { HandlerCreateResponse } from './handler/CreateResponse.ts'

/**
 * @description OK
 */
export type PostMembershipGroups200 = HandlerCreateResponse

/**
 * @description Bad Request
 */
export type PostMembershipGroups400 = EchoHTTPError

/**
 * @description Unauthorized
 */
export type PostMembershipGroups401 = EchoHTTPError

/**
 * @description Forbidden
 */
export type PostMembershipGroups403 = EchoHTTPError

/**
 * @description Unprocessable Entity
 */
export type PostMembershipGroups422 = EchoHTTPError

/**
 * @description Internal Server Error
 */
export type PostMembershipGroups500 = EchoHTTPError

/**
 * @description Group
 */
export type PostMembershipGroupsMutationRequest = HandlerCreateGroupRequest

export type PostMembershipGroupsMutationResponse = PostMembershipGroups200

export type PostMembershipGroupsMutation = {
  Response: PostMembershipGroups200
  Request: PostMembershipGroupsMutationRequest
  Errors: PostMembershipGroups400 | PostMembershipGroups401 | PostMembershipGroups403 | PostMembershipGroups422 | PostMembershipGroups500
}