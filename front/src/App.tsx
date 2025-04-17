import { type NavigateOptions, Route, Routes, useHref, useNavigate } from 'react-router-dom'
import { defaultTheme, Provider } from '@adobe/react-spectrum'

import {
  Dashboard,
  DataSourceEdit,
  DataSources,
  Files,
  Logs,
  Messages,
  OAuth2CredentialEdit,
  OAuth2Credentials,
  PipelineEdit,
  PipelineFlow,
  Pipelines,
  SchedulerEdit,
  Schedulers,
  StorageEdit,
  Storages,
  SyncPolicies,
  SyncPolicyEdit,
  UserEdit,
  Users,
  WorkerJobs,
} from '@/pages'
import { LoginPage, ProtectedRoute, SignupPage } from '@/shauth'

declare module '@adobe/react-spectrum' {
  interface RouterConfig {
    routerOptions: NavigateOptions
  }
}

function App() {
  const navigate = useNavigate()
  return (
    <Provider theme={defaultTheme} colorScheme="light" router={{ navigate, useHref }}>
      <Routes>
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Dashboard />
            </ProtectedRoute>
          }
        />
        <Route
          path="/datasources"
          element={
            <ProtectedRoute>
              <DataSources />
            </ProtectedRoute>
          }
        />
        <Route
          path="/users"
          element={
            <ProtectedRoute>
              <Users />
            </ProtectedRoute>
          }
        />
        <Route
          path="/users/:uuid"
          element={
            <ProtectedRoute>
              <UserEdit />
            </ProtectedRoute>
          }
        />
        <Route
          path="/schedulers"
          element={
            <ProtectedRoute>
              <Schedulers />
            </ProtectedRoute>
          }
        />
        <Route
          path="/schedulers/:uuid"
          element={
            <ProtectedRoute>
              <SchedulerEdit />
            </ProtectedRoute>
          }
        />
        <Route
          path="/datasources/:uuid"
          element={
            <ProtectedRoute>
              <DataSourceEdit />
            </ProtectedRoute>
          }
        />
        <Route path="/login" element={<LoginPage />} />
        <Route
          path="/logs"
          element={
            <ProtectedRoute>
              <Logs />
            </ProtectedRoute>
          }
        />
        <Route
          path="/oauth2/credentials"
          element={
            <ProtectedRoute>
              <OAuth2Credentials />
            </ProtectedRoute>
          }
        />
        <Route
          path="/oauth2/credentials/:clientID"
          element={
            <ProtectedRoute>
              <OAuth2CredentialEdit />
            </ProtectedRoute>
          }
        />
        <Route
          path="/pipelines/:uuid/flow"
          element={
            <ProtectedRoute>
              <PipelineFlow />
            </ProtectedRoute>
          }
        />
        <Route
          path="/pipelines/:uuid"
          element={
            <ProtectedRoute>
              <PipelineEdit />
            </ProtectedRoute>
          }
        />
        <Route
          path="/pipelines"
          element={
            <ProtectedRoute>
              <Pipelines />
            </ProtectedRoute>
          }
        />
        <Route path="/signup" element={<SignupPage />} />
        <Route
          path="/storages"
          element={
            <ProtectedRoute>
              <Storages />
            </ProtectedRoute>
          }
        />
        <Route
          path="/storages/:uuid/storageKind/:storageKind"
          element={
            <ProtectedRoute>
              <StorageEdit />
            </ProtectedRoute>
          }
        />
        <Route
          path="/storages/:uuid"
          element={
            <ProtectedRoute>
              <StorageEdit />
            </ProtectedRoute>
          }
        />
        <Route
          path="/syncpolicies"
          element={
            <ProtectedRoute>
              <SyncPolicies />
            </ProtectedRoute>
          }
        />
        <Route
          path="/syncpolicy/:uuid"
          element={
            <ProtectedRoute>
              <SyncPolicyEdit />
            </ProtectedRoute>
          }
        />
        <Route
          path="/messages"
          element={
            <ProtectedRoute>
              <Messages />
            </ProtectedRoute>
          }
        />

        <Route
          path="/files"
          element={
            <ProtectedRoute>
              <Files />
            </ProtectedRoute>
          }
        />

        <Route
          path="/workers"
          element={
            <ProtectedRoute>
              <WorkerJobs />
            </ProtectedRoute>
          }
        />
      </Routes>
    </Provider>
  )
}

export default App
