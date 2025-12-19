export interface LoaderContext {
  url: string;
  pathname: string;
  params: Record<string, string>;
}

export type LoaderFunction = (ctx: LoaderContext) => Promise<Record<string, unknown>>;

// Route-to-loader mapping
const loaders: Record<string, LoaderFunction> = {
  '/start': async () => {
    // Start page loader
    return { pageTitle: 'Start' };
  },
  '/home': async () => {
    return { pageTitle: 'Home' };
  },
  '/about': async () => {
    return {
      version: '0.0.1',
      pageTitle: 'About'
    };
  },
  '/documentation': async () => {
    // Could fetch documentation sections from API
    return {
      pageTitle: 'Documentation',
      sections: [
        { name: 'Getting Started', path: '/documentation/getting-started' },
        { name: 'Datasources', path: '/documentation/datasource' }
      ]
    };
  },
  '/documentation/datasource': async () => {
    return {
      pageTitle: 'Datasources',
      datasources: [
        { name: 'Gmail', path: '/documentation/datasource/gmail' },
        { name: 'Telegram', path: '/documentation/datasource/telegram' }
      ]
    };
  },
  '/documentation/datasource/gmail': async () => {
    return { pageTitle: 'Gmail Setup' };
  },
  '/documentation/datasource/telegram': async () => {
    return { pageTitle: 'Telegram Setup' };
  }
};

export async function fetchDataForRoute(url: string): Promise<Record<string, unknown>> {
  const urlObj = new URL(url, 'http://localhost');
  const pathname = urlObj.pathname;

  // Find exact match first
  if (loaders[pathname]) {
    return loaders[pathname]({ url, pathname, params: {} });
  }

  // Try to find a matching prefix loader
  const sortedPaths = Object.keys(loaders).sort((a, b) => b.length - a.length);
  for (const path of sortedPaths) {
    if (pathname.startsWith(path)) {
      return loaders[path]({ url, pathname, params: {} });
    }
  }

  // Default empty data
  return {};
}

// Register a loader for a route (useful for dynamic registration)
export function registerLoader(path: string, loader: LoaderFunction): void {
  loaders[path] = loader;
}
