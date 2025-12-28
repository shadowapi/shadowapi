import { useState, useEffect } from 'react';

/**
 * Hook to detect responsive breakpoints.
 * Uses Ant Design's lg breakpoint (992px) as the mobile/desktop threshold.
 */
export function useResponsive() {
  // Default to false (desktop) for SSR hydration compatibility
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const checkBreakpoint = () => {
      setIsMobile(window.innerWidth < 992);
    };

    // Check immediately on mount
    checkBreakpoint();

    // Listen for resize events
    window.addEventListener('resize', checkBreakpoint);
    return () => window.removeEventListener('resize', checkBreakpoint);
  }, []);

  return { isMobile };
}
