/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

export const dbChangesetChangeType = {
  ChangesetChangeTypeCreate: 'create',
  ChangesetChangeTypeUpdate: 'update',
  ChangesetChangeTypeDelete: 'delete',
} as const

export type DbChangesetChangeTypeEnum = (typeof dbChangesetChangeType)[keyof typeof dbChangesetChangeType]

export type DbChangesetChangeType = DbChangesetChangeTypeEnum