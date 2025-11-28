import { Routes, Route } from 'react-router';
import { SSRProvider } from './lib/ssr-context';
import { routes } from './routes';
import AppLayout from './layouts/AppLayout';
import PageLayout from './layouts/PageLayout';

interface AppProps {
  ssrData?: Record<string, unknown>;
}

function App({ ssrData }: AppProps) {
  return (
    <SSRProvider initialData={ssrData}>
      <Routes>
        {routes.map((route) => (
          <Route
            key={route.path}
            path={route.path}
            element={
              route.layout === 'app' ? (
                <AppLayout>{route.element}</AppLayout>
              ) : (
                <PageLayout>{route.element}</PageLayout>
              )
            }
          />
        ))}
      </Routes>
    </SSRProvider>
  );
}

export default App;
