import { createServerFn } from '@tanstack/react-start';
import { getCookie, setCookie, deleteCookie, getContext, setContext } from '@tanstack/react-start/server';
import { createContext, useState, use, useEffect } from 'react';
import { z } from 'zod';
import { AuthUser, postAuthLogin, postAuthRefreshToken, getAuthUser } from './gen';
import { isServer } from './utils/is-server';

export type User = AuthUser;

interface AuthContextType {
  user: User;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

export const AnonymousUser: User = {
  id: 0,
  username: 'Anonymous',
  isAuthenticated: false,
  isGlobalAdmin: false,
};

function setTokenCookies(accessToken: string, refreshToken: string) {
  setCookie('access_token', accessToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    maxAge: 60 * 15, // 15 minutes
  });

  setCookie('refresh_token', refreshToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    path: '/',
    maxAge: 60 * 60 * 24 * 30, // 30 days
  });
}

export function setRequestAccessToken(accessToken: string | null) {
  if (accessToken) {
    setContext('access_token', accessToken);
  }
}

export function getRequestAccessToken(): string | null {
  return getContext('access_token') ?? null;
}

const loginFn = createServerFn({ method: 'POST' })
  .validator(
    z.object({
      username: z.string(),
      password: z.string(),
    })
  )
  .handler(async (ctx) => {
    try {
      const { username, password } = ctx.data;
      const { access_token, refresh_token } = await postAuthLogin({ username, password });
      setTokenCookies(access_token, refresh_token);

      const user = await getAuthUser({ headers: { Authorization: `Bearer ${access_token}` } });

      return { accessToken: access_token, user };
    } catch (error: any) {
      throw new Error((error.response?.data?.message ?? error.message) || 'An unexpected error occurred. Please try again later.');
    }
  });

const logoutFn = createServerFn({ method: 'POST' }).handler(async () => {
  deleteCookie('access_token');
  deleteCookie('refresh_token');
});

export const refreshFn = createServerFn({ method: 'POST' }).handler(async () => {
  const refresh_token = getCookie('refresh_token');

  if (!refresh_token) {
    throw new Error('No refresh token found');
  }

  const { access_token: accessToken, refresh_token: refreshToken } = await postAuthRefreshToken({ refresh_token });
  setTokenCookies(accessToken, refreshToken);

  return { accessToken };
});

export function getAccessToken(): string | null {
  if (isServer) {
    return getRequestAccessToken() ?? getCookie('access_token') ?? null;
  }
  return localStorage.getItem('access_token');
}

export function getRefreshToken(): string | null {
  if (isServer) {
    return getCookie('refresh_token') ?? null;
  }

  throw new Error('refresh_token is not available on the client');
}

export function setAccessToken(accessToken: string | null) {
  if (isServer) {
    return;
  }

  
  if (accessToken) {
    localStorage.setItem('access_token', accessToken);
  } else {
    localStorage.removeItem('access_token');
  }
}
const AuthContext = createContext<AuthContextType>(undefined as unknown as AuthContextType);

export const AuthProvider = ({
  children,
  accessToken,
  initialUser,
}: {
  children: React.ReactNode;
  accessToken: string | null;
  initialUser: AuthUser;
}) => {
  useEffect(() => {
    setAccessToken(accessToken);
  }, []);
  const [userState, setUserState] = useState<User>(initialUser);

  async function login(username: string, password: string) {
    const { accessToken, user } = await loginFn({ data: { username, password } });
    setAccessToken(accessToken);
    setUserState(user);
  }

  async function logout() {
    await logoutFn();
    setUserState(AnonymousUser);
    setAccessToken(null);
  }

  return <AuthContext value={{ user: userState, login, logout }}>{children}</AuthContext>;
};

export const useAuth = () => {
  const context = use(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
