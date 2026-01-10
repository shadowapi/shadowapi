import { createContext } from 'react';

// UserPolicySet represents a policy set assignment for a user
export interface UserPolicySet {
  policy_set: string;
  domain: string;
}

// CurrentWorkspace represents the workspace the JWT is scoped to
export interface CurrentWorkspace {
  uuid: string;
  slug: string;
}

// User type based on backend API response
export interface User {
  uuid: string;
  email: string;
  first_name: string;
  last_name: string;
  policy_sets: UserPolicySet[];
  current_workspace?: CurrentWorkspace;
}

// Helper function to check if user has admin privileges (super_admin policy set in global domain)
export function isAdmin(user: User | null): boolean {
  if (!user) return false;
  return user.policy_sets?.some(p => p.policy_set === 'super_admin' && p.domain === 'global') ?? false;
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
