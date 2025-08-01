/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { UseMutationOptions, QueryClient } from '@tanstack/react-query'
import type {
  PutMembershipUsersUserIdMutationRequest,
  PutMembershipUsersUserIdMutationResponse,
  PutMembershipUsersUserIdPathParams,
  PutMembershipUsersUserId400,
  PutMembershipUsersUserId401,
  PutMembershipUsersUserId403,
  PutMembershipUsersUserId404,
  PutMembershipUsersUserId500,
} from '../types/PutMembershipUsersUserId.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { useMutation } from '@tanstack/react-query'

export const putMembershipUsersUserIdMutationKey = () => [{ url: '/membership/users/{user_id}' }] as const

export type PutMembershipUsersUserIdMutationKey = ReturnType<typeof putMembershipUsersUserIdMutationKey>

/**
 * @description Update a user by ID
 * @summary Update a user
 * {@link /membership/users/:user_id}
 */
export async function putMembershipUsersUserId(
  user_id: PutMembershipUsersUserIdPathParams['user_id'],
  data: PutMembershipUsersUserIdMutationRequest,
  config: Partial<RequestConfig<PutMembershipUsersUserIdMutationRequest>> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    PutMembershipUsersUserIdMutationResponse,
    ResponseErrorConfig<
      PutMembershipUsersUserId400 | PutMembershipUsersUserId401 | PutMembershipUsersUserId403 | PutMembershipUsersUserId404 | PutMembershipUsersUserId500
    >,
    PutMembershipUsersUserIdMutationRequest
  >({ method: 'PUT', url: `/membership/users/${user_id}`, data, ...requestConfig })
  return res.data
}

/**
 * @description Update a user by ID
 * @summary Update a user
 * {@link /membership/users/:user_id}
 */
export function usePutMembershipUsersUserId<TContext>(
  options: {
    mutation?: UseMutationOptions<
      PutMembershipUsersUserIdMutationResponse,
      ResponseErrorConfig<
        PutMembershipUsersUserId400 | PutMembershipUsersUserId401 | PutMembershipUsersUserId403 | PutMembershipUsersUserId404 | PutMembershipUsersUserId500
      >,
      { user_id: PutMembershipUsersUserIdPathParams['user_id']; data: PutMembershipUsersUserIdMutationRequest },
      TContext
    > & { client?: QueryClient }
    client?: Partial<RequestConfig<PutMembershipUsersUserIdMutationRequest>> & { client?: typeof client }
  } = {},
) {
  const { mutation = {}, client: config = {} } = options ?? {}
  const { client: queryClient, ...mutationOptions } = mutation
  const mutationKey = mutationOptions.mutationKey ?? putMembershipUsersUserIdMutationKey()

  return useMutation<
    PutMembershipUsersUserIdMutationResponse,
    ResponseErrorConfig<
      PutMembershipUsersUserId400 | PutMembershipUsersUserId401 | PutMembershipUsersUserId403 | PutMembershipUsersUserId404 | PutMembershipUsersUserId500
    >,
    { user_id: PutMembershipUsersUserIdPathParams['user_id']; data: PutMembershipUsersUserIdMutationRequest },
    TContext
  >(
    {
      mutationFn: async ({ user_id, data }) => {
        return putMembershipUsersUserId(user_id, data, config)
      },
      mutationKey,
      ...mutationOptions,
    },
    queryClient,
  )
}