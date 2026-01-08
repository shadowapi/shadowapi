import { Link } from 'react-router';
import { type ReactNode } from 'react';

interface SmartLinkProps {
  to: string;
  children: ReactNode;
  className?: string;
  style?: React.CSSProperties;
}

/**
 * SmartLink provides unified navigation.
 *
 * - External URLs (http://, https://) = Standard <a> tag with target="_blank"
 * - Internal paths = React Router <Link> for SPA navigation
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

  // All internal paths use React Router for SPA navigation
  return (
    <Link to={to} className={className} style={style}>
      {children}
    </Link>
  )
}

export default SmartLink
