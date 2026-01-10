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

// Policy set constants
export const PolicySets = {
  SUPER_ADMIN: 'super_admin',
  WORKSPACE_OWNER: 'workspace_owner',
  WORKSPACE_ADMIN: 'workspace_admin',
  WORKSPACE_MEMBER: 'workspace_member',
} as const;

// Helper function to check if user has admin privileges (super_admin policy set in global domain)
export function isAdmin(user: User | null): boolean {
  if (!user) return false;
  return user.policy_sets?.some(p => p.policy_set === PolicySets.SUPER_ADMIN && p.domain === 'global') ?? false;
}

// Alias for isAdmin for clarity
export function isSuperAdmin(user: User | null): boolean {
  return isAdmin(user);
}

// Check if user has any of the specified roles in a workspace
export function hasWorkspaceRole(user: User | null, workspaceSlug: string, roles: string[]): boolean {
  if (!user || !workspaceSlug) return false;
  return user.policy_sets?.some(
    p => roles.includes(p.policy_set) && p.domain === workspaceSlug
  ) ?? false;
}

// Check if user is workspace owner or super_admin
export function isWorkspaceOwnerOrAbove(user: User | null, workspaceSlug: string): boolean {
  if (isSuperAdmin(user)) return true;
  return hasWorkspaceRole(user, workspaceSlug, [PolicySets.WORKSPACE_OWNER]);
}

// Check if user is workspace admin, owner, or super_admin
export function isWorkspaceAdminOrAbove(user: User | null, workspaceSlug: string): boolean {
  if (isSuperAdmin(user)) return true;
  return hasWorkspaceRole(user, workspaceSlug, [PolicySets.WORKSPACE_OWNER, PolicySets.WORKSPACE_ADMIN]);
}

// Check if user is at least a workspace member (or higher)
export function isWorkspaceMemberOrAbove(user: User | null, workspaceSlug: string): boolean {
  if (isSuperAdmin(user)) return true;
  return hasWorkspaceRole(user, workspaceSlug, [
    PolicySets.WORKSPACE_OWNER,
    PolicySets.WORKSPACE_ADMIN,
    PolicySets.WORKSPACE_MEMBER
  ]);
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
