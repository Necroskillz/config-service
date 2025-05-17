import { createContext, use, useEffect, useState } from 'react';
import { getChangesetsCurrentQueryKey, HandlerChangesetInfoResponse, useGetChangesetsCurrent, useGetChangesetsCurrentSuspense } from '~/gen';
import { useQueryClient } from '@tanstack/react-query';
import { useAuth } from '~/auth';

export type ChangesetContext = {
  id: number;
  numberOfChanges: number;
  refresh: () => Promise<void>;
};

const ChangesetContext = createContext<ChangesetContext>(undefined as unknown as ChangesetContext);

export function ChangesetProvider({ children }: { children: React.ReactNode }) {
  const { user } = useAuth();
  const { data: changesetData } = useGetChangesetsCurrent({
    query: {
      enabled: user.isAuthenticated,
    },
  });
  const queryClient = useQueryClient();
  const [changeset, setChangeset] = useState<HandlerChangesetInfoResponse>({ id: 0, numberOfChanges: 0 });

  useEffect(() => {
    if (changesetData) {
      setChangeset(changesetData);
    } else {
      setChangeset({ id: 0, numberOfChanges: 0 });
    }
  }, [changesetData]);

  async function refresh() {
    await queryClient.refetchQueries({ queryKey: getChangesetsCurrentQueryKey() });
  }

  return (
    <ChangesetContext.Provider value={{ id: changeset.id, numberOfChanges: changeset.numberOfChanges, refresh }}>
      {children}
    </ChangesetContext.Provider>
  );
}

export function useChangeset() {
  const changeset = use(ChangesetContext);
  if (!changeset) {
    throw new Error('useChangeset must be used within a ChangesetProvider');
  }

  return changeset;
}
