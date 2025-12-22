import { useState, useEffect, useCallback, useRef, type ReactNode } from 'react';
import {
  initiateOAuth2Flow,
  refreshToken as oauth2RefreshToken,
  checkSession as oauth2CheckSession,
  logout as oauth2Logout,
  OAuth2Error,
} from './oauth2-client';
import { AuthContext, type AuthContextType, type User } from './AuthContext';
import client from '../../api/client';

// API base URL - use environment variable or fallback to default
const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || 'http://api.localtest.me/api/v1';
const AUTH_LOGIN_URL = `${API_BASE_URL}/auth/login`;

interface LoginSubmitResponse {
  redirect_to: string;
}

export class AuthError extends Error {
  status?: number;
  details?: unknown;

  constructor(message: string, status?: number, details?: unknown) {
    super(message);
    this.name = 'AuthError';
    this.status = status;
    this.details = details;
  }
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tokenExpiresIn, setTokenExpiresIn] = useState<number | null>(null);
  const refreshTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Fetch user profile from the backend
  const fetchUserProfile = useCallback(async (): Promise<User | null> => {
    try {
      const { data, error } = await client.GET('/profile');
      if (error || !data) {
        console.error('Failed to fetch user profile:', error);
        return null;
      }
      return {
        uuid: data.uuid ?? '',
        email: data.email ?? '',
        first_name: data.first_name ?? '',
        last_name: data.last_name ?? '',
        is_admin: data.is_admin ?? false,
      };
    } catch (err) {
      console.error('Failed to fetch user profile:', err);
      return null;
    }
  }, []);

  // Check OAuth2 session on mount without triggering token refresh
  const checkSession = useCallback(async () => {
    try {
      const session = await oauth2CheckSession();
      if (session.authenticated) {
        setIsAuthenticated(true);
        setTokenExpiresIn(session.expires_in ?? null);
        // Fetch user profile to get user details including is_admin
        const userProfile = await fetchUserProfile();
        if (userProfile) {
          setUser(userProfile);
        }
      } else {
        setIsAuthenticated(false);
        setUser(null);
        setTokenExpiresIn(null);
      }
    } catch (err) {
      console.error('Session check failed:', err);
      setIsAuthenticated(false);
      setUser(null);
      setTokenExpiresIn(null);
    }
    setIsLoading(false);
  }, [fetchUserProfile]);

  // Auto-refresh token before it expires
  useEffect(() => {
    if (!isAuthenticated || !tokenExpiresIn) {
      return;
    }

    // Refresh 60 seconds before expiry
    const refreshIn = Math.max((tokenExpiresIn - 60) * 1000, 10000);

    refreshTimerRef.current = setTimeout(async () => {
      try {
        const response = await oauth2RefreshToken();
        setTokenExpiresIn(response.expires_in);
      } catch {
        // Refresh failed - user needs to re-login
        setIsAuthenticated(false);
        setUser(null);
        setTokenExpiresIn(null);
      }
    }, refreshIn);

    return () => {
      if (refreshTimerRef.current) {
        clearTimeout(refreshTimerRef.current);
      }
    };
  }, [isAuthenticated, tokenExpiresIn]);

  useEffect(() => {
    checkSession();
  }, [checkSession]);

  const login = useCallback(async (email: string, password: string, loginChallenge?: string) => {
    setError(null);
    setIsLoading(true);

    try {
      if (loginChallenge) {
        // OAuth2 flow: Submit credentials to backend's /auth/login endpoint
        const response = await fetch(AUTH_LOGIN_URL, {
          method: 'POST',
          credentials: 'include',
          headers: {
            'Content-Type': 'application/json',
            Accept: 'application/json',
          },
          body: JSON.stringify({
            login_challenge: loginChallenge,
            email,
            password,
            remember: false,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}));
          throw new AuthError(
            errorData.detail || 'Authentication failed',
            response.status,
            errorData
          );
        }

        const data: LoginSubmitResponse = await response.json();

        // Redirect to Hydra consent (which auto-approves and returns tokens)
        window.location.href = data.redirect_to;
        return;
      }

      // Direct login (not OAuth2 flow): Initiate OAuth2 flow first
      // This will redirect to /auth/login with a login_challenge
      const currentUrl = window.location.origin + '/workspaces';
      const oauth2Response = await initiateOAuth2Flow(currentUrl);
      window.location.href = oauth2Response.authorization_url;

      // Note: The page will reload after OAuth2 callback
    } catch (err) {
      setIsLoading(false);
      if (err instanceof AuthError) {
        setError(err.message);
      } else if (err instanceof OAuth2Error) {
        setError(err.message);
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('An unexpected error occurred');
      }
      throw err;
    }
    // Note: Don't set isLoading to false here because we're redirecting
  }, []);

  const logout = useCallback(async () => {
    setIsLoading(true);
    try {
      await oauth2Logout();
      setIsAuthenticated(false);
      setUser(null);
      setTokenExpiresIn(null);
    } catch (err) {
      if (err instanceof Error) {
        setError(err.message);
      }
    } finally {
      setIsLoading(false);
    }
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  const value: AuthContextType = {
    user,
    isAuthenticated,
    isLoading,
    error,
    tokenExpiresIn,
    login,
    logout,
    clearError,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
