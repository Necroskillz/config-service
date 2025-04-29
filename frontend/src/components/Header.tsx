import { Link } from '@tanstack/react-router';
import { useAuth } from '~/auth';
import { Badge } from './ui/badge';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger, DropdownMenuTriggerLabel } from './ui/dropdown-menu';
import { useChangeset } from '~/hooks/useChangeset';

export function Header() {
  const { user } = useAuth();
  const { id: changesetId, numberOfChanges } = useChangeset();

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
