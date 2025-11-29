import { lazy, Suspense, type ReactNode } from 'react';
import { Spin } from 'antd';

// SSR Pages - imported directly (bundled for server)
import AboutPage from './pages/about';
import DocumentationIndex from './pages/documentation';
import DatasourceIndex from './pages/documentation/datasource';
import GmailDocumentation from './pages/documentation/datasource/gmail';
import TelegramDocumentation from './pages/documentation/datasource/telegram';

// CSR Pages - lazy loaded (only on client)
const Dashboard = lazy(() => import('./app/Dashboard'));
const LoginPage = lazy(() => import('./pages/auth/LoginPage'));

// Loading fallback for lazy components
function LoadingFallback() {
  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 200 }}>
      <Spin size="large" />
    </div>
  );
}

export interface RouteConfig {
  path: string;
  element: ReactNode;
  layout: 'page' | 'app' | 'auth';
  ssr: boolean;
  protected?: boolean;
}

// Wrap lazy components with Suspense
function withSuspense(Component: React.LazyExoticComponent<React.ComponentType>) {
  return (
    <Suspense fallback={<LoadingFallback />}>
      <Component />
    </Suspense>
  );
}

export const routes: RouteConfig[] = [
  // Auth routes
  {
    path: '/login',
    element: withSuspense(LoginPage),
    layout: 'auth',
    ssr: false,
    protected: false
  },

  // CSR routes (app dashboard) - protected
  {
    path: '/',
    element: withSuspense(Dashboard),
    layout: 'app',
    ssr: false
  },
  {
    path: '/app/*',
    element: withSuspense(Dashboard),
    layout: 'app',
    ssr: false
  },

  // SSR routes (public pages)
  {
    path: '/page',
    element: <AboutPage />,
    layout: 'page',
    ssr: true
  },
  {
    path: '/page/about',
    element: <AboutPage />,
    layout: 'page',
    ssr: true
  },
  {
    path: '/page/documentation',
    element: <DocumentationIndex />,
    layout: 'page',
    ssr: true
  },
  {
    path: '/page/documentation/datasource',
    element: <DatasourceIndex />,
    layout: 'page',
    ssr: true
  },
  {
    path: '/page/documentation/datasource/gmail',
    element: <GmailDocumentation />,
    layout: 'page',
    ssr: true
  },
  {
    path: '/page/documentation/datasource/telegram',
    element: <TelegramDocumentation />,
    layout: 'page',
    ssr: true
  }
];

// Helper to check if a path should be SSR rendered
export function isSSRRoute(pathname: string): boolean {
  return pathname.startsWith('/page');
}

// Get route config by path
export function getRouteConfig(pathname: string): RouteConfig | undefined {
  return routes.find(route => {
    if (route.path.endsWith('/*')) {
      const basePath = route.path.slice(0, -2);
      return pathname.startsWith(basePath);
    }
    return route.path === pathname;
  });
}
