import { createContext } from 'react';

// User type based on backend API response
export interface User {
  uuid: string;
  email: string;
  first_name: string;
  last_name: string;
  is_admin: boolean;
}

export interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  tokenExpiresIn: number | null;
  login: (email: string, password: string, loginChallenge?: string) => Promise<void>;
  logout: () => Promise<void>;
  clearError: () => void;
}

export const AuthContext = createContext<AuthContextType | null>(null);
