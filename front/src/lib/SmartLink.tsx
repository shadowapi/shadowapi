import { Link } from 'react-router';
import { type ReactNode } from 'react';

interface SmartLinkProps {
  to: string;
  children: ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

/**
 * SmartLink provides unified navigation for the app.
 *
 * - External URLs (http://, https://) = Standard <a> tag
 * - Internal routes = React Router <Link> for SPA navigation
 *
 * This enables seamless SPA navigation between SSR (/page/*) and CSR (/) routes
 * when the user is already in the app. Direct URL access still uses SSR where configured.
 */
export function SmartLink({ to, children, className, style }: SmartLinkProps) {
  // External URLs use standard anchor tag
  if (to.startsWith('http://') || to.startsWith('https://')) {
    return (
      <a href={to} className={className} style={style} target="_blank" rel="noopener noreferrer">
        {children}
      </a>
    );
  }

  // All internal navigation uses React Router for SPA experience
  return (
    <Link to={to} className={className} style={style}>
      {children}
    </Link>
  );
}

export default SmartLink;
