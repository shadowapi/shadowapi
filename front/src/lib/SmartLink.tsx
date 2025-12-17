import { Link } from 'react-router';
import { type ReactNode } from 'react';

// Subdomain URLs from environment
const WWW_BASE_URL =
  import.meta.env.VITE_WWW_BASE_URL || 'http://www.localtest.me'
const APP_BASE_URL = import.meta.env.VITE_APP_BASE_URL || 'http://localtest.me'

// SSR routes that live on www subdomain
const SSR_PATHS = ['/start', '/about', '/documentation']

// Check if a path is an SSR route (www subdomain)
function isSSRPath(path: string): boolean {
  return SSR_PATHS.some((p) => path === p || path.startsWith(p + '/'))
}

// Check if a path is an app route (root domain)
function isAppPath(path: string): boolean {
  return (
    path === '/' ||
    path.startsWith('/workspaces') ||
    path.startsWith('/w/') ||
    path.startsWith('/login')
  )
}

// Get current subdomain context
function getCurrentContext(): 'www' | 'app' {
  if (typeof window === 'undefined') return 'app'
  const hostname = window.location.hostname
  return hostname.startsWith('www.') ? 'www' : 'app'
}

interface SmartLinkProps {
  to: string;
  children: ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

/**
 * SmartLink provides unified navigation across subdomains.
 *
 * - External URLs (http://, https://) = Standard <a> tag with target="_blank"
 * - Cross-subdomain routes = Standard <a> tag for full page navigation
 * - Same-subdomain routes = React Router <Link> for SPA navigation
 *
 * Subdomain architecture:
 * - www.{domain} - SSR pages (/start, /about, /documentation)
 * - {domain} - App routes (/workspaces, /w/*, /login)
 */
export function SmartLink({ to, children, className, style }: SmartLinkProps) {
  // External URLs use standard anchor tag
  if (to.startsWith('http://') || to.startsWith('https://')) {
    return (
      <a
        href={to}
        className={className}
        style={style}
        target="_blank"
        rel="noopener noreferrer"
      >
        {children}
      </a>
    )
  }

  const currentContext = getCurrentContext()
  const targetIsSSR = isSSRPath(to)
  const targetIsApp = isAppPath(to)

  // Cross-subdomain navigation requires full page redirect
  if (currentContext === 'app' && targetIsSSR) {
    // From app to www subdomain
    return (
      <a href={`${WWW_BASE_URL}${to}`} className={className} style={style}>
        {children}
      </a>
    )
  }

  if (currentContext === 'www' && targetIsApp) {
    // From www to app subdomain
    return (
      <a href={`${APP_BASE_URL}${to}`} className={className} style={style}>
        {children}
      </a>
    )
  }

  // Same-subdomain navigation uses React Router for SPA experience
  return (
    <Link to={to} className={className} style={style}>
      {children}
    </Link>
  )
}

export default SmartLink
