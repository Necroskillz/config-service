/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { UseMutationOptions, QueryClient } from '@tanstack/react-query'
import type {
  PutChangesetsChangesetIdApplyMutationResponse,
  PutChangesetsChangesetIdApplyPathParams,
  PutChangesetsChangesetIdApply400,
  PutChangesetsChangesetIdApply401,
  PutChangesetsChangesetIdApply403,
  PutChangesetsChangesetIdApply404,
  PutChangesetsChangesetIdApply500,
} from '../types/PutChangesetsChangesetIdApply.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { useMutation } from '@tanstack/react-query'

export const putChangesetsChangesetIdApplyMutationKey = () => [{ url: '/changesets/{changeset_id}/apply' }] as const

export type PutChangesetsChangesetIdApplyMutationKey = ReturnType<typeof putChangesetsChangesetIdApplyMutationKey>

/**
 * @description Apply a changeset by ID
 * @summary Apply a changeset
 * {@link /changesets/:changeset_id/apply}
 */
export async function putChangesetsChangesetIdApply(
  changeset_id: PutChangesetsChangesetIdApplyPathParams['changeset_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    PutChangesetsChangesetIdApplyMutationResponse,
    ResponseErrorConfig<
      | PutChangesetsChangesetIdApply400
      | PutChangesetsChangesetIdApply401
      | PutChangesetsChangesetIdApply403
      | PutChangesetsChangesetIdApply404
      | PutChangesetsChangesetIdApply500
    >,
    unknown
  >({ method: 'PUT', url: `/changesets/${changeset_id}/apply`, ...requestConfig })
  return res.data
}

/**
 * @description Apply a changeset by ID
 * @summary Apply a changeset
 * {@link /changesets/:changeset_id/apply}
 */
export function usePutChangesetsChangesetIdApply<TContext>(
  options: {
    mutation?: UseMutationOptions<
      PutChangesetsChangesetIdApplyMutationResponse,
      ResponseErrorConfig<
        | PutChangesetsChangesetIdApply400
        | PutChangesetsChangesetIdApply401
        | PutChangesetsChangesetIdApply403
        | PutChangesetsChangesetIdApply404
        | PutChangesetsChangesetIdApply500
      >,
      { changeset_id: PutChangesetsChangesetIdApplyPathParams['changeset_id'] },
      TContext
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { mutation = {}, client: config = {} } = options ?? {}
  const { client: queryClient, ...mutationOptions } = mutation
  const mutationKey = mutationOptions.mutationKey ?? putChangesetsChangesetIdApplyMutationKey()

  return useMutation<
    PutChangesetsChangesetIdApplyMutationResponse,
    ResponseErrorConfig<
      | PutChangesetsChangesetIdApply400
      | PutChangesetsChangesetIdApply401
      | PutChangesetsChangesetIdApply403
      | PutChangesetsChangesetIdApply404
      | PutChangesetsChangesetIdApply500
    >,
    { changeset_id: PutChangesetsChangesetIdApplyPathParams['changeset_id'] },
    TContext
  >(
    {
      mutationFn: async ({ changeset_id }) => {
        return putChangesetsChangesetIdApply(changeset_id, config)
      },
      mutationKey,
      ...mutationOptions,
    },
    queryClient,
  )
}