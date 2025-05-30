import { z, ZodType } from 'zod';
import { DbValueTypeKind, DbValueValidatorType } from '~/gen';
import { Validator } from 'jsonschema';

export function createDefaultValue(valueType: DbValueTypeKind) {
  switch (valueType) {
    case 'boolean':
      return 'FALSE';
    default:
      return '';
  }
}

type ValidatorRefiner = (parameter: string | undefined, errorText: string | undefined) => (zv: ZodType<string>) => ZodType<string>;
export type ValueValidatorDef = {
  validatorType: DbValueValidatorType;
  parameter?: string;
  errorText?: string;
};

function parseIntParameter(parameterName: string, parameter: string | undefined) {
  if (!parameter) {
    throw new Error(`Value validator ${parameterName}: parameter is required`);
  }

  const num = parseInt(parameter);
  if (isNaN(num)) {
    throw new Error(`Value validator ${parameterName}: parameter is not a number`);
  }

  return num;
}

function parseFloatParameter(parameterName: string, parameter: string | undefined) {
  if (!parameter) {
    throw new Error(`Value validator ${parameterName}: parameter is required`);
  }

  const num = parseFloat(parameter);
  if (isNaN(num)) {
    throw new Error(`Value validator ${parameterName}: parameter is not a number`);
  }

  return num;
}

function createRequiredValidatorRefiner(): ValidatorRefiner {
  return (_, errorText) => {
    return (zv: ZodType<string>) => {
      return zv.refine((value) => {
        return value != null && value !== '';
      }, errorText || 'Value is required');
    };
  };
}

function createMinValidatorRefiner(): ValidatorRefiner {
  return (parameter, errorText) => {
    const min = parseIntParameter('min', parameter);
    return (zv: ZodType<string>) => {
      return zv.refine((value) => {
        if (!value) {
          return true;
        }

        return parseIntParameter('value', value) >= min;
      }, errorText || `Value must be at least ${min}`);
    };
  };
}

function createMaxValidatorRefiner(): ValidatorRefiner {
  return (parameter, errorText) => {
    const max = parseIntParameter('max', parameter);
    return (zv: ZodType<string>) => {
      return zv.refine((value) => {
        if (!value) {
          return true;
        }

        return parseIntParameter('value', value) <= max;
      }, errorText || `Value must be at most ${max}`);
    };
  };
}

function createMinDecimalValidatorRefiner(): ValidatorRefiner {
  return (parameter, errorText) => {
    const min = parseFloatParameter('min_decimal', parameter);
    return (zv: ZodType<string>) => {
      return zv.refine((value) => {
        if (!value) {
          return true;
        }

        return parseFloatParameter('value', value) >= min;
      }, errorText || `Value must be at least ${min}`);
    };
  };
}

function createMaxDecimalValidatorRefiner(): ValidatorRefiner {
  return (parameter, errorText) => {
    const max = parseFloatParameter('max_decimal', parameter);
    return (zv: ZodType<string>) => {
      return zv.refine((value) => {
        if (!value) {
          return true;
        }

        return parseFloatParameter('value', value) <= max;
      }, errorText || `Value must be at most ${max}`);
    };
  };
}

function createMinLengthValidatorRefiner(): ValidatorRefiner {
  return (parameter, errorText) => {
    const minLength = parseIntParameter('min_length', parameter);
    return (zv: ZodType<string>) => {
      return zv.refine((value) => {
        if (!value) {
          return true;
        }

        return value.length >= minLength;
      }, errorText || `Value must be at least ${minLength} characters long`);
    };
  };
}

function createMaxLengthValidatorRefiner(): ValidatorRefiner {
  return (parameter, errorText) => {
    const maxLength = parseIntParameter('max_length', parameter);
    return (zv: ZodType<string>) => {
      return zv.refine((value) => {
        if (!value) {
          return true;
        }

        return value.length <= maxLength;
      }, errorText || `Value must have at most ${maxLength} characters`);
    };
  };
}

function createRegexValidatorRefiner(): ValidatorRefiner {
  return (parameter, errorText) => {
    if (!parameter) {
      throw new Error(`Value validator regex: parameter is required`);
    }
    const regex = new RegExp(parameter);
    return (zv: ZodType<string>) => {
      return zv.refine((value) => {
        if (!value) {
          return true;
        }

        return regex.test(value);
      }, errorText || `Value must match the pattern ${parameter}`);
    };
  };
}

function createValidJsonValidatorRefiner(): ValidatorRefiner {
  return (_, errorText) => {
    return (zv: ZodType<string>) => {
      return zv.superRefine((value, ctx) => {
        if (!value) {
          return true;
        }

        try {
          JSON.parse(value);
          return true;
        } catch (error: any) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: (errorText || 'Value must be valid JSON: {0}').replace('{0}', error.message),
          });
        }
      });
    };
  };
}

function createJsonSchemaValidatorRefiner(): ValidatorRefiner {
  return (parameter, errorText) => {
    if (!parameter) {
      throw new Error(`Value validator json_schema: parameter is required`);
    }
    const schema = JSON.parse(parameter);
    const validator = new Validator();
    return (zv: ZodType<string>) => {
      return zv.superRefine((value, ctx) => {
        if (!value) {
          return true;
        }

        try {
          const o = JSON.parse(value);
          const result = validator.validate(o, schema);
          if (result.valid) {
            return true;
          }
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: (errorText || 'Value must match the JSON schema: {0}').replace('{0}', result.errors.map((e) => e.message).join(', ')),
          });
        } catch (error: any) {
          // This is validated by valid_json validator
          return true;
        }
      });
    };
  };
}

function createValidIntegerValidatorRefiner(): ValidatorRefiner {
  return (_, errorText) => {
    return (zv: ZodType<string>) => {
      return zv.refine((value) => {
        if (!value) {
          return true;
        }

        if (!/^\d+$/.test(value)) {
          return false;
        }
        return !isNaN(parseInt(value));
      }, errorText || 'Value must be a valid integer');
    };
  };
}

function createValidFloatValidatorRefiner(): ValidatorRefiner {
  return (_, errorText) => {
    return (zv: ZodType<string>) => {
      return zv.refine((value) => {
        if (!value) {
          return true;
        }

        if (!/^\d+(\.\d+)?$/.test(value)) {
          return false;
        }
        return !isNaN(parseFloat(value));
      }, errorText || 'Value must be a valid number with an optional decimal part');
    };
  };
}

function createValidRegexValidatorRefiner(): ValidatorRefiner {
  return (_, errorText) => {
    return (zv: ZodType<string>) =>
      zv.refine((value) => {
        if (!value) {
          return true;
        }

        try {
          new RegExp(value);
          return true;
        } catch (e: any) {
          return false;
        }
      }, errorText || 'Value must be a valid regex');
  };
}

const valueValidators: Record<DbValueValidatorType, ValidatorRefiner> = {
  required: createRequiredValidatorRefiner(),
  min_length: createMinLengthValidatorRefiner(),
  max_length: createMaxLengthValidatorRefiner(),
  min: createMinValidatorRefiner(),
  max: createMaxValidatorRefiner(),
  min_decimal: createMinDecimalValidatorRefiner(),
  max_decimal: createMaxDecimalValidatorRefiner(),
  regex: createRegexValidatorRefiner(),
  json_schema: createJsonSchemaValidatorRefiner(),
  valid_json: createValidJsonValidatorRefiner(),
  valid_integer: createValidIntegerValidatorRefiner(),
  valid_decimal: createValidFloatValidatorRefiner(),
  valid_regex: createValidRegexValidatorRefiner(),
};

export function createValueValidator(validators: ValueValidatorDef[]) {
  let zv: ZodType<string> = z.string();

  validators.forEach((validator) => {
    const refiner = valueValidators[validator.validatorType];
    if (!refiner) {
      throw new Error(`Value validator ${validator.validatorType}: validator not found`);
    }

    zv = refiner(validator.parameter, validator.errorText)(zv);
  });

  return zv;
}
