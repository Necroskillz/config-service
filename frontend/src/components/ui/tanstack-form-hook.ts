import { createFormHook } from '@tanstack/react-form';
import { fieldContext, formContext, useFieldContext, useFormContext } from './tanstack-form-context';
import { FormLabel, FormControl, FormDescription, FormMessage } from './tanstack-form';

const { useAppForm, withForm } = createFormHook({
    fieldContext,
    formContext,
    fieldComponents: {
      FormLabel,
      FormControl,
      FormDescription,
      FormMessage,
    },
    formComponents: {},
  });

  
export { useAppForm, useFormContext, useFieldContext, withForm };