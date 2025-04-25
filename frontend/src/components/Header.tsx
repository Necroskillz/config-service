import { Link } from '@tanstack/react-router';
import { useAuth } from '~/auth';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger, DropdownMenuTriggerLabel } from './ui/dropdown-menu';

export function Header() {
  const { user } = useAuth();
  return (
    <div className="p-2 flex gap-2 text-lg">
      <Link to="/" activeOptions={{ exact: true }}>
        <img src="/logo-no-background.png" alt="Config Service" className="w-25" />
      </Link>
      {user.isAuthenticated && (
        <>
          <Link to="/services">Services</Link>
          {user.changesetId > 0 ? (
            <Link to="/changesets/$changesetId" params={{ changesetId: user.changesetId }}>
              Changeset
            </Link>
          ) : (
            <span>No Changeset</span>
          )}
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
