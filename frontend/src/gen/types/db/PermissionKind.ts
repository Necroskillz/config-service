/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

export const dbPermissionKind = {
  PermissionKindService: 'service',
  PermissionKindFeature: 'feature',
  PermissionKindKey: 'key',
  PermissionKindVariation: 'variation',
} as const

export type DbPermissionKindEnum = (typeof dbPermissionKind)[keyof typeof dbPermissionKind]

export type DbPermissionKind = DbPermissionKindEnum