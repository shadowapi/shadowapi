import { Routes, Route } from 'react-router';
import Dashboard from './Dashboard';
import OAuth2Credentials from './oauth2/OAuth2Credentials';
import OAuth2CredentialEdit from './oauth2/OAuth2CredentialEdit';

function AppRouter() {
  return (
    <Routes>
      <Route index element={<Dashboard />} />
      <Route path="oauth2/credentials" element={<OAuth2Credentials />} />
      <Route path="oauth2/credentials/new" element={<OAuth2CredentialEdit />} />
      <Route path="oauth2/credentials/:uuid" element={<OAuth2CredentialEdit />} />
    </Routes>
  );
}

export default AppRouter;
