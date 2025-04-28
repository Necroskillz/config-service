import { Slot } from '@radix-ui/react-slot';
import { Label } from '~/components/ui/label';
import { cn } from '~/lib/utils';
import { useStore } from '@tanstack/react-form';
import { useFieldContext } from './tanstack-form-context';

export function FormLabel({ className, ...props }: React.ComponentPropsWithoutRef<typeof Label>) {
  const { name, store } = useFieldContext();
  const errors = useStore(store, (state) => state.meta.errors);
  const isTouched = useStore(store, (state) => state.meta.isTouched);
  const hasError = errors.length > 0 && isTouched;

  return (
    <Label
      data-slot="form-label"
      data-error={hasError}
      className={cn('data-[error=true]:text-destructive', className)}
      htmlFor={name}
      {...props}
    />
  );
}

export function FormControl({ ...props }: React.ComponentPropsWithoutRef<typeof Slot>) {
  const { name, store } = useFieldContext();
  const errors = useStore(store, (state) => state.meta.errors);
  const isTouched = useStore(store, (state) => state.meta.isTouched);
  const hasError = errors.length > 0 && isTouched;

  return (
    <Slot
      data-slot="form-control"
      id={name}
      aria-describedby={hasError ? `${name}-description ${name}-message` : `${name}-description`}
      aria-invalid={hasError}
      {...props}
    />
  );
}

export function FormDescription({ className, ...props }: React.ComponentPropsWithoutRef<'p'>) {
  const { name } = useFieldContext();

  return (
    <p data-slot="form-description" id={`${name}-description`} className={cn('text-muted-foreground text-sm', className)} {...props} />
  );
}

export function FormMessage({ className, ...props }: React.ComponentPropsWithoutRef<'p'>) {
  const { name, store } = useFieldContext();
  const errors = useStore(store, (state) => state.meta.errors);
  const isTouched = useStore(store, (state) => state.meta.isTouched);
  const hasError = errors.length > 0 && isTouched;

  const body = hasError ? String(errors.at(0)?.message ?? errors.at(0) ?? '') : props.children;
  if (!body) return null;

  return (
    <p data-slot="form-message" id={`${name}-message`} className={cn('text-destructive text-sm', className)} {...props}>
      {body}
    </p>
  );
}
