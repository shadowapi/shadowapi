import { useState, useEffect, useCallback, useRef, type ReactNode } from 'react';
import {
  type KratosSession,
  type KratosIdentity,
  createLoginFlow,
  submitLogin,
  KratosAuthError,
} from './kratos-client';
import {
  initiateOAuth2Flow,
  refreshToken as oauth2RefreshToken,
  logout as oauth2Logout,
  OAuth2Error,
} from './oauth2-client';
import { AuthContext, type AuthContextType } from './AuthContext';

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [session, setSession] = useState<KratosSession | null>(null);
  const [user, setUser] = useState<KratosIdentity | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tokenExpiresIn, setTokenExpiresIn] = useState<number | null>(null);
  const refreshTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Check OAuth2 session on mount by attempting token refresh
  const checkSession = useCallback(async () => {
    try {
      // Try to refresh the token - if successful, we have a valid session
      const response = await oauth2RefreshToken();
      setIsAuthenticated(true);
      setTokenExpiresIn(response.expires_in);
      // Note: User info would typically come from the JWT claims or a /userinfo endpoint
      // For now, we mark as authenticated without user details
    } catch {
      // No valid session
      setIsAuthenticated(false);
      setUser(null);
      setSession(null);
      setTokenExpiresIn(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

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
        setSession(null);
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

  const login = useCallback(async (email: string, password: string) => {
    setError(null);
    setIsLoading(true);

    try {
      // Step 1: Authenticate with Kratos
      const flow = await createLoginFlow();

      const csrfNode = flow.ui.nodes.find(
        (node) => node.attributes.name === 'csrf_token'
      );
      const csrfToken = csrfNode?.attributes.value;

      const kratosSession = await submitLogin(flow.id, email, password, csrfToken);

      // Store user info from Kratos session
      setSession(kratosSession);
      setUser(kratosSession.identity);

      // Step 2: Initiate OAuth2 flow to get tokens
      // The redirect_uri should be the current page or the app root
      const currentUrl = window.location.origin + window.location.pathname;
      const oauth2Response = await initiateOAuth2Flow(currentUrl);

      // Step 3: Redirect to Hydra for authorization
      // After Hydra consent (auto-approved since user is authenticated),
      // the backend will set cookies and redirect back
      window.location.href = oauth2Response.authorization_url;

      // Note: The page will reload after OAuth2 callback,
      // so we don't need to set isAuthenticated here
    } catch (err) {
      setIsLoading(false);
      if (err instanceof KratosAuthError) {
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
      setSession(null);
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
    session,
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
