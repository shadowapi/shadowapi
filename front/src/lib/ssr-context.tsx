import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';

// Extend Window interface for SSR data
declare global {
  interface Window {
    __SSR_DATA__?: Record<string, unknown>;
  }
}

interface SSRContextType {
  data: Record<string, unknown>;
  isHydrating: boolean;
}

const SSRContext = createContext<SSRContextType>({
  data: {},
  isHydrating: false
});

interface SSRProviderProps {
  children: ReactNode;
  initialData?: Record<string, unknown>;
}

export function SSRProvider({ children, initialData }: SSRProviderProps) {
  const [data] = useState<Record<string, unknown>>(() => {
    // On client, read from window if available (hydration)
    if (typeof window !== 'undefined' && window.__SSR_DATA__) {
      return window.__SSR_DATA__;
    }
    // On server or client without SSR data
    return initialData || {};
  });

  const [isHydrating, setIsHydrating] = useState(true);

  useEffect(() => {
    // After first render, we're no longer hydrating
    setIsHydrating(false);
  }, []);

  return (
    <SSRContext.Provider value={{ data, isHydrating }}>
      {children}
    </SSRContext.Provider>
  );
}

export function useSSRData<T = unknown>(key: string): T | undefined {
  const { data } = useContext(SSRContext);
  return data[key] as T | undefined;
}

export function useIsHydrating(): boolean {
  const { isHydrating } = useContext(SSRContext);
  return isHydrating;
}
