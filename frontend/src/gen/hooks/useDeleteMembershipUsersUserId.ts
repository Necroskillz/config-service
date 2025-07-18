/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { UseMutationOptions, QueryClient } from '@tanstack/react-query'
import type {
  DeleteMembershipUsersUserIdMutationResponse,
  DeleteMembershipUsersUserIdPathParams,
  DeleteMembershipUsersUserId400,
  DeleteMembershipUsersUserId401,
  DeleteMembershipUsersUserId403,
  DeleteMembershipUsersUserId404,
  DeleteMembershipUsersUserId500,
} from '../types/DeleteMembershipUsersUserId.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { useMutation } from '@tanstack/react-query'

export const deleteMembershipUsersUserIdMutationKey = () => [{ url: '/membership/users/{user_id}' }] as const

export type DeleteMembershipUsersUserIdMutationKey = ReturnType<typeof deleteMembershipUsersUserIdMutationKey>

/**
 * @description Delete a user by ID
 * @summary Delete a user
 * {@link /membership/users/:user_id}
 */
export async function deleteMembershipUsersUserId(
  user_id: DeleteMembershipUsersUserIdPathParams['user_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    DeleteMembershipUsersUserIdMutationResponse,
    ResponseErrorConfig<
      | DeleteMembershipUsersUserId400
      | DeleteMembershipUsersUserId401
      | DeleteMembershipUsersUserId403
      | DeleteMembershipUsersUserId404
      | DeleteMembershipUsersUserId500
    >,
    unknown
  >({ method: 'DELETE', url: `/membership/users/${user_id}`, ...requestConfig })
  return res.data
}

/**
 * @description Delete a user by ID
 * @summary Delete a user
 * {@link /membership/users/:user_id}
 */
export function useDeleteMembershipUsersUserId<TContext>(
  options: {
    mutation?: UseMutationOptions<
      DeleteMembershipUsersUserIdMutationResponse,
      ResponseErrorConfig<
        | DeleteMembershipUsersUserId400
        | DeleteMembershipUsersUserId401
        | DeleteMembershipUsersUserId403
        | DeleteMembershipUsersUserId404
        | DeleteMembershipUsersUserId500
      >,
      { user_id: DeleteMembershipUsersUserIdPathParams['user_id'] },
      TContext
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { mutation = {}, client: config = {} } = options ?? {}
  const { client: queryClient, ...mutationOptions } = mutation
  const mutationKey = mutationOptions.mutationKey ?? deleteMembershipUsersUserIdMutationKey()

  return useMutation<
    DeleteMembershipUsersUserIdMutationResponse,
    ResponseErrorConfig<
      | DeleteMembershipUsersUserId400
      | DeleteMembershipUsersUserId401
      | DeleteMembershipUsersUserId403
      | DeleteMembershipUsersUserId404
      | DeleteMembershipUsersUserId500
    >,
    { user_id: DeleteMembershipUsersUserIdPathParams['user_id'] },
    TContext
  >(
    {
      mutationFn: async ({ user_id }) => {
        return deleteMembershipUsersUserId(user_id, config)
      },
      mutationKey,
      ...mutationOptions,
    },
    queryClient,
  )
}