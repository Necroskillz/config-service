/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import type { EchoHTTPError } from './echo/HTTPError.ts'
import type { HandlerCreateResponse } from './handler/CreateResponse.ts'
import type { HandlerCreateUserRequest } from './handler/CreateUserRequest.ts'

/**
 * @description OK
 */
export type PostMembershipUsers200 = HandlerCreateResponse

/**
 * @description Bad Request
 */
export type PostMembershipUsers400 = EchoHTTPError

/**
 * @description Unauthorized
 */
export type PostMembershipUsers401 = EchoHTTPError

/**
 * @description Forbidden
 */
export type PostMembershipUsers403 = EchoHTTPError

/**
 * @description Unprocessable Entity
 */
export type PostMembershipUsers422 = EchoHTTPError

/**
 * @description Internal Server Error
 */
export type PostMembershipUsers500 = EchoHTTPError

/**
 * @description User
 */
export type PostMembershipUsersMutationRequest = HandlerCreateUserRequest

export type PostMembershipUsersMutationResponse = PostMembershipUsers200

export type PostMembershipUsersMutation = {
  Response: PostMembershipUsers200
  Request: PostMembershipUsersMutationRequest
  Errors: PostMembershipUsers400 | PostMembershipUsers401 | PostMembershipUsers403 | PostMembershipUsers422 | PostMembershipUsers500
}