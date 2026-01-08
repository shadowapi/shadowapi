import { lazy, Suspense, type ReactNode } from 'react';
import { Spin } from 'antd';

// SSR Pages - imported directly (bundled for server)
import LandingPage from './pages/landing';
import AboutPage from './pages/about';
import DocumentationIndex from './pages/documentation';
import DatasourceIndex from './pages/documentation/datasource';
import GmailDocumentation from './pages/documentation/datasource/gmail';
import TelegramDocumentation from './pages/documentation/datasource/telegram';

// CSR Pages - lazy loaded (only on client)
const WorkspaceRouter = lazy(() => import('./app/WorkspaceRouter'));
const WorkspaceSelectionPage = lazy(() => import('./pages/workspaces'));
const LoginPage = lazy(() => import('./pages/auth/LoginPage'));
const AcceptInvitePage = lazy(() => import('./pages/auth/AcceptInvitePage'));
const ForgotPasswordPage = lazy(() => import('./pages/auth/ForgotPasswordPage'));
const ResetPasswordPage = lazy(() => import('./pages/auth/ResetPasswordPage'));
const RootRedirect = lazy(() => import('./pages/RootRedirect'));

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
  layout: 'app' | 'auth' | 'landing';
  ssr: boolean;
  protected?: boolean;
  showBreadcrumb?: boolean;
  showSidebar?: boolean;
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
  {
    path: '/invite/:token',
    element: withSuspense(AcceptInvitePage),
    layout: 'auth',
    ssr: false,
    protected: false
  },
  {
    path: '/forgot-password',
    element: withSuspense(ForgotPasswordPage),
    layout: 'auth',
    ssr: false,
    protected: false
  },
  {
    path: '/reset-password/:token',
    element: withSuspense(ResetPasswordPage),
    layout: 'auth',
    ssr: false,
    protected: false
  },

  // Workspace selection - protected, centered layout without sidebar
  {
    path: '/workspaces',
    element: withSuspense(WorkspaceSelectionPage),
    layout: 'auth',
    ssr: false,
    protected: true,
  },

  // Workspace routes - protected, with workspace context
  // Using /* to allow nested routes in WorkspaceRouter
  {
    path: '/w/:slug/*',
    element: withSuspense(WorkspaceRouter),
    layout: 'app',
    ssr: false,
    protected: true,
  },

  // Root redirect on app subdomain - authenticated users go to /workspaces, others to /login
  {
    path: '/',
    element: withSuspense(RootRedirect),
    layout: 'auth',
    ssr: false,
    protected: false,
  },

  // SSR routes (public pages on root domain)
  {
    path: '/start',
    element: <LandingPage />,
    layout: 'landing',
    ssr: true
  },
  {
    path: '/about',
    element: <AboutPage />,
    layout: 'app',
    ssr: true,
    showSidebar: false,
  },
  {
    path: '/documentation',
    element: <DocumentationIndex />,
    layout: 'app',
    ssr: true,
    showSidebar: false,
  },
  {
    path: '/documentation/datasource',
    element: <DatasourceIndex />,
    layout: 'app',
    ssr: true,
    showSidebar: false,
  },
  {
    path: '/documentation/datasource/gmail',
    element: <GmailDocumentation />,
    layout: 'app',
    ssr: true,
    showSidebar: false,
  },
  {
    path: '/documentation/datasource/telegram',
    element: <TelegramDocumentation />,
    layout: 'app',
    ssr: true,
    showSidebar: false,
  }
];

// Helper to check if a path should be SSR rendered
// SSR routes are on root domain (not app subdomain)
export function isSSRRoute(pathname: string): boolean {
  const ssrPaths = ['/start', '/about', '/documentation'];
  return ssrPaths.some(p => pathname === p || pathname.startsWith(p + '/'));
}

// Get route config by path
export function getRouteConfig(pathname: string): RouteConfig | undefined {
  return routes.find(route => {
    if (route.path.endsWith('/*')) {
      const basePath = route.path.slice(0, -2);
      return pathname.startsWith(basePath);
    }
    // Handle parameterized routes like /w/:slug
    if (route.path.includes(':')) {
      const pattern = route.path.replace(/:[\w]+/g, '[^/]+');
      const regex = new RegExp(`^${pattern}$`);
      return regex.test(pathname);
    }
    return route.path === pathname;
  });
}
