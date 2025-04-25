import { ResponseErrorConfig } from '~/axios';
import { EchoHTTPError } from '~/gen';
import { Alert, AlertDescription } from './ui/alert';

export function MutationErrors({ mutations }: { mutations: { isError: boolean; error: ResponseErrorConfig<EchoHTTPError> | null }[] }) {
  return (
    <>
      {mutations
        .filter((mutation) => mutation.isError)
        .map((mutation) => (
          <Alert key={mutation.error!.message} variant="destructive">
            <AlertDescription>{mutation.error!.message}</AlertDescription>
          </Alert>
        ))}
    </>
  );
}
