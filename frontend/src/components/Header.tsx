import { Link } from '@tanstack/react-router';
import { useAuth } from '~/auth';
import { Badge } from './ui/badge';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger, DropdownMenuTriggerLabel } from './ui/dropdown-menu';
import { useChangeset } from '~/hooks/useChangeset';
import { useGetChangesetsApprovableCount } from '~/gen';

export function Header() {
  const { user } = useAuth();
  const { id: changesetId, numberOfChanges } = useChangeset();
  const { data: approvableCount } = useGetChangesetsApprovableCount({
    query: {
      refetchInterval: 60000,
    },
  });

  return (
    <div className="p-2 flex gap-2 text-lg">
      <Link to="/" activeOptions={{ exact: true }}>
        <img src="/logo-no-background.png" alt="Config Service" className="w-25" />
      </Link>
      {user.isAuthenticated && (
        <>
          <Link to="/services">Services</Link>
          <Link
            to={changesetId > 0 ? '/changesets/$changesetId' : '/changesets/empty'}
            params={{ changesetId }}
            activeOptions={{ exact: true }}
          >
            Changeset <Badge variant="outline">{numberOfChanges}</Badge>
          </Link>
          <Link
            to="/changesets"
            search={{ mode: approvableCount?.count && approvableCount.count > 0 ? 'approvable' : 'my' }}
            activeOptions={{ exact: true, includeSearch: false }}
          >
            Changesets
            {approvableCount?.count && approvableCount.count > 0 && <Badge variant="outline">{approvableCount.count}</Badge>}
          </Link>
        </>
      )}
      <div className="ml-auto">{user.isAuthenticated ? <UserMenu /> : <Link to="/login">Login</Link>}</div>
    </div>
  );
}

function UserMenu() {
  const { user, logout } = useAuth();
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <DropdownMenuTriggerLabel>{user.username}</DropdownMenuTriggerLabel>
      </DropdownMenuTrigger>
      <DropdownMenuContent>
        <DropdownMenuItem onClick={logout}>Logout</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
