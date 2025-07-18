/**
 * Generated by Kubb (https://kubb.dev/).
 * Do not edit manually.
 */

import type { DbValueTypeKind } from '../db/ValueTypeKind.ts'
import type { ValidationValidatorDto } from '../validation/ValidatorDto.ts'

export type KeyKeyDto = {
  /**
   * @type boolean
   */
  canEdit: boolean
  /**
   * @type string
   */
  description: string
  /**
   * @type integer
   */
  id: number
  /**
   * @type string
   */
  name: string
  /**
   * @type array
   */
  validators: ValidationValidatorDto[]
  /**
   * @type string
   */
  valueType: DbValueTypeKind
  /**
   * @type integer
   */
  valueTypeId: number
  /**
   * @type string
   */
  valueTypeName: string
}