import { createContext } from 'react';
import type { KratosIdentity, KratosSession } from './kratos-client';

export interface AuthContextType {
  user: KratosIdentity | null;
  session: KratosSession | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
}

export const AuthContext = createContext<AuthContextType | null>(null);
