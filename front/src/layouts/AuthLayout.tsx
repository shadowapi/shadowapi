import { type ReactNode } from 'react';
import { Layout, theme } from 'antd';

const { Content } = Layout;

interface AuthLayoutProps {
  children: ReactNode;
}

function AuthLayout({ children }: AuthLayoutProps) {
  const {
    token: { colorBgContainer },
  } = theme.useToken();

  return (
    <Layout style={{ minHeight: '100vh', background: colorBgContainer }}>
      <Content
        style={{
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          padding: 24,
        }}
      >
        {children}
      </Content>
    </Layout>
  );
}

export default AuthLayout;
