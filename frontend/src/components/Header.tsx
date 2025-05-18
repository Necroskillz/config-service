import { Link } from '@tanstack/react-router';
import { useAuth } from '~/auth';
import { Badge } from './ui/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuPortal,
  DropdownMenuShortcut,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
  DropdownMenuTriggerLabel,
} from './ui/dropdown-menu';
import { useChangeset } from '~/hooks/useChangeset';
import { useGetChangesetsApprovableCount } from '~/gen';
import { useTheme } from '~/ThemeProvider';
import { Sun, Moon, Monitor, Check } from 'lucide-react';

export function Header() {
  const { user } = useAuth();
  const { id: changesetId, numberOfChanges } = useChangeset();
  const { data: approvableCount } = useGetChangesetsApprovableCount({
    query: {
      refetchInterval: 60000,
      enabled: user.isAuthenticated,
    },
  });

  return (
    <div className="p-2 flex gap-2 text-lg items-center">
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
            className="flex items-center gap-1"
          >
            Changeset <Badge variant="outline">{numberOfChanges}</Badge>
          </Link>
          <Link
            to="/changesets"
            search={{ mode: approvableCount?.count && approvableCount.count > 0 ? 'approvable' : 'my' }}
            activeOptions={{ exact: true, includeSearch: false }}
            className="flex items-center gap-1"
          >
            Changesets
            {approvableCount?.count && approvableCount.count > 0 ? <Badge variant="outline">{approvableCount.count}</Badge> : null}
          </Link>
          {user.isGlobalAdmin && <Link to="/admin">Admin</Link>}
        </>
      )}
      <div className="ml-auto">{user.isAuthenticated ? <UserMenu /> : <Link to="/login">Login</Link>}</div>
    </div>
  );
}

function UserMenu() {
  const { user, logout } = useAuth();
  const { setTheme, theme } = useTheme();

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <DropdownMenuTriggerLabel>{user.username}</DropdownMenuTriggerLabel>
      </DropdownMenuTrigger>

      <DropdownMenuContent>
        <DropdownMenuSub>
          <DropdownMenuSubTrigger>Theme</DropdownMenuSubTrigger>
          <DropdownMenuPortal>
            <DropdownMenuSubContent>
              <DropdownMenuItem onClick={() => setTheme('system')}>
                <Monitor className="w-4 h-4" />
                System
                {(!theme || theme === 'system') && (
                  <DropdownMenuShortcut>
                    <Check className="w-4 h-4" />
                  </DropdownMenuShortcut>
                )}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTheme('light')}>
                <Sun className="w-4 h-4" />
                Light
                {theme === 'light' && (
                  <DropdownMenuShortcut>
                    <Check className="w-4 h-4" />
                  </DropdownMenuShortcut>
                )}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => setTheme('dark')}>
                <Moon className="w-4 h-4" />
                Dark
                {theme === 'dark' && (
                  <DropdownMenuShortcut>
                    <Check className="w-4 h-4" />
                  </DropdownMenuShortcut>
                )}
              </DropdownMenuItem>
            </DropdownMenuSubContent>
          </DropdownMenuPortal>
        </DropdownMenuSub>
        <DropdownMenuItem onClick={logout}>Logout</DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
