import { useState, useEffect, useCallback, type ReactNode } from 'react';
import {
  type KratosSession,
  getSession,
  createLoginFlow,
  submitLogin,
  executeLogout,
  KratosAuthError,
} from './kratos-client';
import { AuthContext, type AuthContextType } from './AuthContext';

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [session, setSession] = useState<KratosSession | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const checkSession = useCallback(async () => {
    try {
      const currentSession = await getSession();
      setSession(currentSession);
    } catch {
      setSession(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    checkSession();
  }, [checkSession]);

  const login = useCallback(async (email: string, password: string) => {
    setError(null);
    setIsLoading(true);

    try {
      const flow = await createLoginFlow();

      const csrfNode = flow.ui.nodes.find(
        (node) => node.attributes.name === 'csrf_token'
      );
      const csrfToken = csrfNode?.attributes.value;

      const newSession = await submitLogin(flow.id, email, password, csrfToken);
      setSession(newSession);
    } catch (err) {
      if (err instanceof KratosAuthError) {
        setError(err.message);
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('An unexpected error occurred');
      }
      throw err;
    } finally {
      setIsLoading(false);
    }
  }, []);

  const logout = useCallback(async () => {
    setIsLoading(true);
    try {
      await executeLogout();
      setSession(null);
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
    user: session?.identity ?? null,
    session,
    isAuthenticated: session?.active ?? false,
    isLoading,
    error,
    login,
    logout,
    clearError,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
