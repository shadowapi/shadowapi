import { createContext } from 'react';

// UserRole represents a role assignment for a user
export interface UserRole {
  role: string;
  domain: string;
}

// User type based on backend API response
export interface User {
  uuid: string;
  email: string;
  first_name: string;
  last_name: string;
  roles: UserRole[];
}

// Helper function to check if user has admin privileges (super_admin role in global domain)
export function isAdmin(user: User | null): boolean {
  if (!user) return false;
  return user.roles?.some(r => r.role === 'super_admin' && r.domain === 'global') ?? false;
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
