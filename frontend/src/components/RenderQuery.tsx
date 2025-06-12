import { UseQueryResult } from '@tanstack/react-query';
import { ResponseErrorConfig } from '~/axios';
import { EchoHTTPError } from '~/gen';
import { Spinner } from './ui/spinner';
import { Alert, AlertDescription } from './ui/alert';
import { useEffect, useState } from 'react';
import { useDebouncer } from '@tanstack/react-pacer/debouncer';

export function RenderQuery<T>({
  query,
  emptyMessage,
  children,
  disabledMessage,
}: {
  query: UseQueryResult<T, ResponseErrorConfig<EchoHTTPError>>;
  emptyMessage?: string;
  disabledMessage?: string;
  children: (data: T) => React.ReactNode;
}) {
  const { status, error, data, fetchStatus } = query;
  const [showSpinner, setShowSpinner] = useState(false);

  const spinnerDebouncer = useDebouncer(setShowSpinner, { wait: 200 });

  useEffect(() => {
    if (status === 'pending') {
      spinnerDebouncer.maybeExecute(true);
    } else {
      spinnerDebouncer.cancel();
      setShowSpinner(false);
    }
  }, [status]);

  if (status === 'pending') {
    if (fetchStatus === 'idle' && disabledMessage) {
      return <div className="text-muted-foreground">{disabledMessage}</div>;
    }

    if (showSpinner) {
      return <Spinner />;
    }

    return null;
  }

  if (status === 'error') {
    return (
      <Alert variant="destructive">
        <AlertDescription>{error.message}</AlertDescription>
      </Alert>
    );
  }

  if (status === 'success' && Array.isArray(data) && data.length === 0 && emptyMessage) {
    return <div className="text-muted-foreground">{emptyMessage}</div>;
  }

  return children(data);
}
