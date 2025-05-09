import { StandardSchemaV1Issue } from '@tanstack/react-form';
import { cn } from '~/lib/utils';

export function ZodErrorMessage({
  errors,
  pathFilter,
}: {
  errors: (Record<string, StandardSchemaV1Issue[]> | undefined)[];
  pathFilter?: string;
}) {
  if (!errors) {
    return null;
  }

  let schemaErrors = errors.flatMap((error) => Object.values(error ?? {}).flatMap((error) => error.map((error) => error)));

  if (pathFilter) {
    const pathFilterRegex = new RegExp(pathFilter);
    schemaErrors = schemaErrors.filter((error) => error.path && pathFilterRegex.test(error.path.join('.')));
  }

  if (schemaErrors.length === 0) {
    return null;
  }

  const errorMessages = schemaErrors.map((error) => error.message);

  return (
    <div className="flex flex-col">
      {errorMessages.map((error) => (
        <p key={error} className={cn('text-destructive text-sm')}>
          {error}
        </p>
      ))}
    </div>
  );
}
