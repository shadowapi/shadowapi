export interface LoaderContext {
  url: string;
  pathname: string;
  params: Record<string, string>;
}

export type LoaderFunction = (ctx: LoaderContext) => Promise<Record<string, unknown>>;

// Route-to-loader mapping
const loaders: Record<string, LoaderFunction> = {
  '/page': async () => {
    // Home page loader
    return { pageTitle: 'Home' };
  },
  '/page/home': async () => {
    return { pageTitle: 'Home' };
  },
  '/page/about': async () => {
    return {
      version: '0.0.1',
      pageTitle: 'About'
    };
  },
  '/page/documentation': async () => {
    // Could fetch documentation sections from API
    return {
      pageTitle: 'Documentation',
      sections: [
        { name: 'Getting Started', path: '/page/documentation/getting-started' },
        { name: 'Datasources', path: '/page/documentation/datasource' }
      ]
    };
  },
  '/page/documentation/datasource': async () => {
    return {
      pageTitle: 'Datasources',
      datasources: [
        { name: 'Gmail', path: '/page/documentation/datasource/gmail' },
        { name: 'Telegram', path: '/page/documentation/datasource/telegram' }
      ]
    };
  },
  '/page/documentation/datasource/gmail': async () => {
    return { pageTitle: 'Gmail Setup' };
  },
  '/page/documentation/datasource/telegram': async () => {
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
