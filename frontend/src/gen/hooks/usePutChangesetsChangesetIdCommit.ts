/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import client from '~/axios'
import type { UseMutationOptions, QueryClient } from '@tanstack/react-query'
import type {
  PutChangesetsChangesetIdCommitMutationResponse,
  PutChangesetsChangesetIdCommitPathParams,
  PutChangesetsChangesetIdCommit400,
  PutChangesetsChangesetIdCommit401,
  PutChangesetsChangesetIdCommit403,
  PutChangesetsChangesetIdCommit404,
  PutChangesetsChangesetIdCommit500,
} from '../types/PutChangesetsChangesetIdCommit.ts'
import type { RequestConfig, ResponseErrorConfig } from '~/axios'
import { useMutation } from '@tanstack/react-query'

export const putChangesetsChangesetIdCommitMutationKey = () => [{ url: '/changesets/{changeset_id}/commit' }] as const

export type PutChangesetsChangesetIdCommitMutationKey = ReturnType<typeof putChangesetsChangesetIdCommitMutationKey>

/**
 * @description Commit a changeset by ID
 * @summary Commit a changeset
 * {@link /changesets/:changeset_id/commit}
 */
export async function putChangesetsChangesetIdCommit(
  changeset_id: PutChangesetsChangesetIdCommitPathParams['changeset_id'],
  config: Partial<RequestConfig> & { client?: typeof client } = {},
) {
  const { client: request = client, ...requestConfig } = config

  const res = await request<
    PutChangesetsChangesetIdCommitMutationResponse,
    ResponseErrorConfig<
      | PutChangesetsChangesetIdCommit400
      | PutChangesetsChangesetIdCommit401
      | PutChangesetsChangesetIdCommit403
      | PutChangesetsChangesetIdCommit404
      | PutChangesetsChangesetIdCommit500
    >,
    unknown
  >({ method: 'PUT', url: `/changesets/${changeset_id}/commit`, ...requestConfig })
  return res.data
}

/**
 * @description Commit a changeset by ID
 * @summary Commit a changeset
 * {@link /changesets/:changeset_id/commit}
 */
export function usePutChangesetsChangesetIdCommit<TContext>(
  options: {
    mutation?: UseMutationOptions<
      PutChangesetsChangesetIdCommitMutationResponse,
      ResponseErrorConfig<
        | PutChangesetsChangesetIdCommit400
        | PutChangesetsChangesetIdCommit401
        | PutChangesetsChangesetIdCommit403
        | PutChangesetsChangesetIdCommit404
        | PutChangesetsChangesetIdCommit500
      >,
      { changeset_id: PutChangesetsChangesetIdCommitPathParams['changeset_id'] },
      TContext
    > & { client?: QueryClient }
    client?: Partial<RequestConfig> & { client?: typeof client }
  } = {},
) {
  const { mutation = {}, client: config = {} } = options ?? {}
  const { client: queryClient, ...mutationOptions } = mutation
  const mutationKey = mutationOptions.mutationKey ?? putChangesetsChangesetIdCommitMutationKey()

  return useMutation<
    PutChangesetsChangesetIdCommitMutationResponse,
    ResponseErrorConfig<
      | PutChangesetsChangesetIdCommit400
      | PutChangesetsChangesetIdCommit401
      | PutChangesetsChangesetIdCommit403
      | PutChangesetsChangesetIdCommit404
      | PutChangesetsChangesetIdCommit500
    >,
    { changeset_id: PutChangesetsChangesetIdCommitPathParams['changeset_id'] },
    TContext
  >(
    {
      mutationFn: async ({ changeset_id }) => {
        return putChangesetsChangesetIdCommit(changeset_id, config)
      },
      mutationKey,
      ...mutationOptions,
    },
    queryClient,
  )
}