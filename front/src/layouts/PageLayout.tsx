import type { ReactNode } from 'react';
import { Layout, theme, Breadcrumb } from 'antd';
import { Link, useLocation } from 'react-router';

const { Content } = Layout;

import BaseLayout from './BaseLayout';
import { getRouteConfig } from '../routes';
import { SmartLink } from '../lib/SmartLink';

const breadcrumbNameMap: Record<string, string> = {
  '/tenant': 'Select Tenant',
  '/documentation': 'Documentation',
  '/documentation/datasource': 'Datasources',
  '/documentation/datasource/gmail': 'Gmail',
  '/documentation/datasource/telegram': 'Telegram',
  '/about': 'About',
};

interface PageLayoutProps {
  children: ReactNode;
}

function PageLayout({ children }: PageLayoutProps) {
  const location = useLocation();
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  const routeConfig = getRouteConfig(location.pathname);
  const showBreadcrumb = routeConfig?.showBreadcrumb !== false;

  const pathSnippets = location.pathname.split('/').filter((i) => i && i !== 'page' && i !== 'start');

  const breadcrumbItems = [
    {
      title: <SmartLink to="/">Home</SmartLink>,
    },
    ...pathSnippets.map((_, index) => {
      const url = `/${pathSnippets.slice(0, index + 1).join('/')}`;
      const isLast = index === pathSnippets.length - 1;
      const name = breadcrumbNameMap[url] || pathSnippets[index];

      return {
        title: isLast ? name : <Link to={url}>{name}</Link>,
      };
    }),
  ];

  return (
    <BaseLayout>
      <div
        style={{
          padding: '0 48px',
          display: 'flex',
          flexDirection: 'column',
          flex: 1,
        }}
      >
        {showBreadcrumb && (
          <Breadcrumb
            style={{ margin: '16px 0', flexShrink: 0 }}
            items={breadcrumbItems}
          />
        )}
        <div
          style={{
            background: colorBgContainer,
            padding: 24,
            borderRadius: borderRadiusLG,
            flex: 1,
            marginBottom: 24,
          }}
        >
          <Content style={{ height: '100%' }}>
            {children}
          </Content>
        </div>
      </div>
    </BaseLayout>
  );
}

export default PageLayout;
