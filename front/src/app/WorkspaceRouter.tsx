import { Routes, Route } from 'react-router';
import { WorkspaceProvider } from '../lib/workspace/WorkspaceContext';
import Dashboard from './Dashboard';
import DataSources from './datasources/DataSources';
import DataSourceEdit from './datasources/DataSourceEdit';
import Storages from './storages/Storages';
import StorageEdit from './storages/StorageEdit';
import PostgresTablesList from './storages/tables/PostgresTablesList';
import PostgresTableEdit from './storages/tables/PostgresTableEdit';
import OAuth2Credentials from './oauth2/OAuth2Credentials';
import OAuth2CredentialEdit from './oauth2/OAuth2CredentialEdit';
import Users from './users/Users';
import UserEdit from './users/UserEdit';
import Roles from './rbac/Roles';
import RoleEdit from './rbac/RoleEdit';
import Pipelines from './pipelines/Pipelines';
import PipelineEdit from './pipelines/PipelineEdit';
import RegisteredWorkers from './workers/RegisteredWorkers';
import ActiveJobs from './workers/ActiveJobs';
import EnrollmentTokens from './workers/EnrollmentTokens';

/**
 * WorkspaceRouter handles all routes under /w/:slug/*
 * It provides workspace context to all child components.
 */
function WorkspaceRouter() {
  return (
    <WorkspaceProvider>
      <Routes>
        <Route index element={<Dashboard />} />
        <Route path="datasources" element={<DataSources />} />
        <Route path="datasources/new" element={<DataSourceEdit />} />
        <Route path="datasources/:uuid" element={<DataSourceEdit />} />
        <Route path="storages" element={<Storages />} />
        <Route path="storages/new" element={<StorageEdit />} />
        <Route path="storages/new/tables" element={<PostgresTablesList />} />
        <Route path="storages/new/tables/new" element={<PostgresTableEdit />} />
        <Route path="storages/new/tables/:index" element={<PostgresTableEdit />} />
        <Route path="storages/:uuid" element={<StorageEdit />} />
        <Route path="storages/:uuid/tables" element={<PostgresTablesList />} />
        <Route path="storages/:uuid/tables/new" element={<PostgresTableEdit />} />
        <Route path="storages/:uuid/tables/:index" element={<PostgresTableEdit />} />
        <Route path="pipelines" element={<Pipelines />} />
        <Route path="pipelines/new" element={<PipelineEdit />} />
        <Route path="pipelines/:uuid" element={<PipelineEdit />} />
        <Route path="workers" element={<RegisteredWorkers />} />
        <Route path="workers/jobs" element={<ActiveJobs />} />
        <Route path="workers/tokens" element={<EnrollmentTokens />} />
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
