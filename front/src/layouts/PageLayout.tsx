import type { ReactNode } from 'react';
import { Layout, theme, Breadcrumb } from 'antd';
import { Link, useLocation } from 'react-router';

const { Content } = Layout;

import BaseLayout from './BaseLayout';

const breadcrumbNameMap: Record<string, string> = {
  '/page/tenant': 'Select Tenant',
  '/page/documentation': 'Documentation',
  '/page/documentation/datasource': 'Datasources',
  '/page/documentation/datasource/gmail': 'Gmail',
  '/page/documentation/datasource/telegram': 'Telegram',
  '/page/about': 'About',
};

interface PageLayoutProps {
  children: ReactNode;
}

function PageLayout({ children }: PageLayoutProps) {
  const location = useLocation();
  const {
    token: { colorBgContainer, borderRadiusLG },
  } = theme.useToken();

  const pathSnippets = location.pathname.split('/').filter((i) => i && i !== 'page');

  const breadcrumbItems = [
    {
      title: <Link to="/">Dashboard</Link>,
    },
    ...pathSnippets.map((_, index) => {
      const url = `/page/${pathSnippets.slice(0, index + 1).join('/')}`;
      const isLast = index === pathSnippets.length - 1;
      const name = breadcrumbNameMap[url] || pathSnippets[index];

      return {
        title: isLast ? name : <Link to={url}>{name}</Link>,
      };
    }),
  ];

  return (
    <BaseLayout>
      <div style={{ padding: '0 48px' }}>
        <Breadcrumb
          style={{ margin: '16px 0' }}
          items={breadcrumbItems}
        />
        <div
          style={{
            background: colorBgContainer,
            minHeight: 280,
            padding: 24,
            borderRadius: borderRadiusLG,
          }}
        >
          <Content>
            {children}
          </Content>
        </div>
      </div>
    </BaseLayout>
  );
}

export default PageLayout;
