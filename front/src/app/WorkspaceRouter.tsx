import { Routes, Route } from 'react-router';
import { WorkspaceProvider } from '../lib/workspace/WorkspaceContext';
import Dashboard from './Dashboard';
import DataSources from './datasources/DataSources';
import DataSourceEdit from './datasources/DataSourceEdit';
import Storages from './storages/Storages';
import StorageEdit from './storages/StorageEdit';
import LastMessages from './storages/LastMessages';
import PostgresTablesList from './storages/tables/PostgresTablesList';
import PostgresTableEdit from './storages/tables/PostgresTableEdit';
import OAuth2Credentials from './oauth2/OAuth2Credentials';
import OAuth2CredentialEdit from './oauth2/OAuth2CredentialEdit';
import Users from './users/Users';
import UserEdit from './users/UserEdit';
import Invites from './invites/Invites';
import PolicySets from './access/PolicySets';
import PolicySetEdit from './access/PolicySetEdit';
import UsageLimits from './access/UsageLimits';
import UsageLimitEdit from './access/UsageLimitEdit';
import UserUsageLimits from './access/UserUsageLimits';
import UserUsageLimitEdit from './access/UserUsageLimitEdit';
import WorkerUsageLimits from './access/WorkerUsageLimits';
import WorkerUsageLimitEdit from './access/WorkerUsageLimitEdit';
import UsageOverview from './access/UsageOverview';
import Pipelines from './pipelines/Pipelines';
import PipelineEdit from './pipelines/PipelineEdit';
import RegisteredWorkers from './workers/RegisteredWorkers';
import ActiveJobs from './workers/ActiveJobs';
import EnrollmentTokens from './workers/EnrollmentTokens';
import Schedulers from './schedulers/Schedulers';
import SchedulerEdit from './schedulers/SchedulerEdit';

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
        <Route path="storages/messages" element={<LastMessages />} />
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
        <Route path="invites" element={<Invites />} />
        <Route path="access/policy-sets" element={<PolicySets />} />
        <Route path="access/policy-sets/new" element={<PolicySetEdit />} />
        <Route path="access/policy-sets/:uuid" element={<PolicySetEdit />} />
        <Route path="access/usage-overview" element={<UsageOverview />} />
        <Route path="access/usage-limits" element={<UsageLimits />} />
        <Route path="access/usage-limits/new" element={<UsageLimitEdit />} />
        <Route path="access/usage-limits/:uuid" element={<UsageLimitEdit />} />
        <Route path="access/user-usage-limits" element={<UserUsageLimits />} />
        <Route path="access/user-usage-limits/new" element={<UserUsageLimitEdit />} />
        <Route path="access/user-usage-limits/:uuid" element={<UserUsageLimitEdit />} />
        <Route path="access/worker-usage-limits" element={<WorkerUsageLimits />} />
        <Route path="access/worker-usage-limits/new" element={<WorkerUsageLimitEdit />} />
        <Route path="access/worker-usage-limits/:uuid" element={<WorkerUsageLimitEdit />} />
        <Route path="schedulers" element={<Schedulers />} />
        <Route path="schedulers/new" element={<SchedulerEdit />} />
        <Route path="schedulers/:uuid" element={<SchedulerEdit />} />
      </Routes>
    </WorkspaceProvider>
  );
}

export default WorkspaceRouter;
