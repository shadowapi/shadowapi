import { Routes, Route } from 'react-router';
import { WorkspaceProvider } from '../lib/workspace/WorkspaceContext';
import Dashboard from './Dashboard';
import OAuth2Credentials from './oauth2/OAuth2Credentials';
import OAuth2CredentialEdit from './oauth2/OAuth2CredentialEdit';
import Users from './users/Users';
import UserEdit from './users/UserEdit';
import Roles from './rbac/Roles';
import RoleEdit from './rbac/RoleEdit';

/**
 * WorkspaceRouter handles all routes under /w/:slug/*
 * It provides workspace context to all child components.
 */
function WorkspaceRouter() {
  return (
    <WorkspaceProvider>
      <Routes>
        <Route index element={<Dashboard />} />
        <Route path="oauth2/credentials" element={<OAuth2Credentials />} />
        <Route path="oauth2/credentials/new" element={<OAuth2CredentialEdit />} />
        <Route path="oauth2/credentials/:uuid" element={<OAuth2CredentialEdit />} />
        <Route path="users" element={<Users />} />
        <Route path="users/new" element={<UserEdit />} />
        <Route path="users/:uuid" element={<UserEdit />} />
        <Route path="rbac/roles" element={<Roles />} />
        <Route path="rbac/roles/new" element={<RoleEdit />} />
        <Route path="rbac/roles/:uuid" element={<RoleEdit />} />
      </Routes>
    </WorkspaceProvider>
  );
}

export default WorkspaceRouter;
