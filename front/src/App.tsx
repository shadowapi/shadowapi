import {
  defaultTheme,
  Provider,
} from '@adobe/react-spectrum';
import {
  type NavigateOptions,
  useHref,
  useNavigate,
  Routes,
  Route,
} from "react-router-dom"

import {
  Dashboard,
  DataSourceAuth,
  DataSourceEdit,
  DataSources,
  Logs,
  OAuth2CredentialEdit,
  OAuth2Credentials,
  PipelineFlow,
  PipelineEdit,
  Pipelines,
  StorageEdit,
  Storages,
  Workers,
} from '@/pages'
import { ProtectedRoute, LoginPage, SignupPage } from '@/shauth'

declare module '@adobe/react-spectrum' {
  interface RouterConfig {
    routerOptions: NavigateOptions;
  }
}

function App() {
  const navigate = useNavigate();
  return (
    (
      <Provider theme={defaultTheme} colorScheme="light" router={{ navigate, useHref }}>
        <Routes>
          <Route path="/" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
          <Route path="/datasources" element={<ProtectedRoute><DataSources /></ProtectedRoute>} />
          <Route path="/datasources/:uuid" element={<ProtectedRoute><DataSourceEdit /></ProtectedRoute>} />
          <Route path="/datasources/:uuid/auth" element={<ProtectedRoute><DataSourceAuth /></ProtectedRoute>} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/logs" element={<ProtectedRoute><Logs /></ProtectedRoute>} />
          <Route path="/oauth2/credentials" element={<ProtectedRoute><OAuth2Credentials /></ProtectedRoute>} />
          <Route path="/oauth2/credentials/:clientID" element={<ProtectedRoute><OAuth2CredentialEdit /></ProtectedRoute>} />
          <Route path="/pipelines/:uuid/flow" element={<ProtectedRoute><PipelineFlow /></ProtectedRoute>} />
          <Route path="/pipelines/:uuid" element={<ProtectedRoute><PipelineEdit /></ProtectedRoute>} />
          <Route path="/pipelines" element={<ProtectedRoute><Pipelines /></ProtectedRoute>} />
          <Route path="/signup" element={<SignupPage />} />
          <Route path="/storages" element={<ProtectedRoute><Storages /></ProtectedRoute>} />
          <Route path="/storages/:uuid" element={<ProtectedRoute><StorageEdit /></ProtectedRoute>} />
          <Route path="/workers" element={<ProtectedRoute><Workers /></ProtectedRoute>} />
        </Routes>
      </Provider>
    )
  );
}

export default App
