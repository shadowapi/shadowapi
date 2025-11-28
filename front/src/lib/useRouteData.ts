import { useState, useEffect, useRef } from 'react';
import { useLocation } from 'react-router';
import { fetchDataForRoute } from './data-fetching';

// Note: Window.__SSR_DATA__ type is declared in ssr-context.tsx

// Check if we have SSR data available
function getSSRData(): Record<string, unknown> | null {
  if (typeof window !== 'undefined' && window.__SSR_DATA__) {
    return window.__SSR_DATA__;
  }
  return null;
}

interface UseRouteDataResult<T> {
  data: T | undefined;
  loading: boolean;
}

/**
 * Hook for accessing route data with SSR/CSR support.
 *
 * - During SSR hydration: returns data from window.__SSR_DATA__
 * - During CSR navigation: fetches data client-side
 *
 * @returns { data, loading } - The route data and loading state
 */
export function useRouteData<T = Record<string, unknown>>(): UseRouteDataResult<T> {
  const location = useLocation();
  const [data, setData] = useState<T | undefined>(() => {
    // Initialize with SSR data if available and path matches
    const ssrData = getSSRData();
    if (ssrData) {
      return ssrData as T;
    }
    return undefined;
  });
  const [loading, setLoading] = useState(false);

  // Track if we've consumed the SSR data for the initial route
  const consumedSSRData = useRef(false);
  const initialPathname = useRef(location.pathname);

  useEffect(() => {
    // If this is the initial render and we have SSR data, don't fetch
    const ssrData = getSSRData();
    if (!consumedSSRData.current && ssrData && location.pathname === initialPathname.current) {
      consumedSSRData.current = true;
      setData(ssrData as T);
      return;
    }

    // Mark SSR data as consumed after first navigation
    consumedSSRData.current = true;

    // Fetch data client-side for this route
    let cancelled = false;

    async function loadData() {
      setLoading(true);
      try {
        const result = await fetchDataForRoute(location.pathname);
        if (!cancelled) {
          setData(result as T);
        }
      } catch (error) {
        console.error('Failed to fetch route data:', error);
        if (!cancelled) {
          setData(undefined);
        }
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    }

    loadData();

    return () => {
      cancelled = true;
    };
  }, [location.pathname]);

  return { data, loading };
}

export default useRouteData;
