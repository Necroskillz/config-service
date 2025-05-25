import { z } from 'zod';
import Ajv from 'ajv';
import { ValidationValueValidatorParameterType } from '~/gen';

export function createParameterValidator(parameterType: ValidationValueValidatorParameterType): z.ZodType<string | undefined> {
  switch (parameterType) {
    case 'none':
      return z
        .string()
        .max(0, 'Parameter is must not be specified')
        .transform((value) => value || undefined);
    case 'integer':
      return z
        .string()
        .min(1, 'Parameter is required')
        .regex(/^\d+$/, 'Parameter must be a valid integer')
        .refine((value) => !isNaN(parseInt(value)), 'Parameter must be a valid integer');
    case 'float':
      return z
        .string()
        .min(1, 'Parameter is required')
        .regex(/^\d+(\.\d+)?$/, 'Parameter must be a number with an optional decimal part')
        .refine((value) => !isNaN(parseFloat(value)), 'Parameter must be a valid number with an optional decimal part');
    case 'regex':
      return z
        .string()
        .min(1, 'Parameter is required')
        .refine((value) => {
          try {
            new RegExp(value);
            return true;
          } catch (error) {
            return false;
          }
        }, 'Must be a valid regex');
    case 'json_schema':
      return z.string().superRefine((value, ctx) => {
        try {
          const schema = JSON.parse(value);
          const ajv = new Ajv();
          const valid = ajv.validateSchema(schema);
          if (!valid) {
            const errors = ajv.errors?.map((e: any) => e.message).join(', ') || 'Unknown error';
            ctx.addIssue({
              code: z.ZodIssueCode.custom,
              message: `Must be a valid JSON schema: ${errors}`,
            });
          }
        } catch (error: any) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: `Must be a valid JSON schema: ${error.message}`,
          });
        }
      });
    default:
      throw new Error(`Unknown parameter type: ${parameterType}`);
  }
}
