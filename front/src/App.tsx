import { Routes, Route } from 'react-router';
import { SSRProvider } from './lib/ssr-context';
import { AuthProvider, ProtectedRoute } from './lib/auth';
import { routes } from './routes';
import AppLayout from './layouts/AppLayout';
import PageLayout from './layouts/PageLayout';
import AuthLayout from './layouts/AuthLayout';

interface AppProps {
  ssrData?: Record<string, unknown>;
}

function App({ ssrData }: AppProps) {
  return (
    <AuthProvider>
      <SSRProvider initialData={ssrData}>
        <Routes>
          {routes.map((route) => {
            let element = route.element;

            // Wrap in appropriate layout
            if (route.layout === 'app') {
              element = <AppLayout>{element}</AppLayout>;
            } else if (route.layout === 'page') {
              element = <PageLayout>{element}</PageLayout>;
            } else if (route.layout === 'auth') {
              element = <AuthLayout>{element}</AuthLayout>;
            }

            // Wrap protected routes (app layout routes are protected by default)
            if (route.protected !== false && route.layout === 'app') {
              element = <ProtectedRoute>{element}</ProtectedRoute>;
            }

            return <Route key={route.path} path={route.path} element={element} />;
          })}
        </Routes>
      </SSRProvider>
    </AuthProvider>
  );
}

export default App;
