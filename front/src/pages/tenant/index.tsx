import { useState, useEffect } from 'react';
import { Typography, Card, Form, Input, Button, List, Alert, Spin, Space, Divider } from 'antd';
import { GlobalOutlined, LoginOutlined, TeamOutlined } from '@ant-design/icons';

const { Title, Text, Paragraph } = Typography;

interface AuthenticatedTenant {
  tenant_name: string;
  tenant_display_name: string;
  user_email: string;
}

function TenantSelectionPage() {
  const [tenantName, setTenantName] = useState('');
  const [authenticatedTenants, setAuthenticatedTenants] = useState<AuthenticatedTenant[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [checkingTenant, setCheckingTenant] = useState(false);
  const [tenantError, setTenantError] = useState<string | null>(null);

  // Get base domain from current hostname (e.g., localtest.me)
  const getBaseDomain = () => {
    const hostname = window.location.hostname;
    // If we're on a subdomain, extract the base domain
    const parts = hostname.split('.');
    if (parts.length >= 2) {
      return parts.slice(-2).join('.');
    }
    return hostname;
  };

  // Build tenant URL
  const buildTenantUrl = (tenant: string) => {
    const baseDomain = getBaseDomain();
    const protocol = window.location.protocol;
    const port = window.location.port ? `:${window.location.port}` : '';
    return `${protocol}//${tenant}.${baseDomain}${port}`;
  };

  // Fetch authenticated tenants on mount
  useEffect(() => {
    const fetchAuthenticatedTenants = async () => {
      try {
        const response = await fetch('/api/v1/auth/tenants', {
          credentials: 'include', // Include cookies for shared session
        });

        if (response.ok) {
          const data = await response.json();
          setAuthenticatedTenants(Array.isArray(data) ? data : []);
        } else if (response.status !== 401) {
          // 401 is expected if no shared session exists
          console.error('Failed to fetch authenticated tenants');
        }
      } catch (err) {
        console.error('Error fetching authenticated tenants:', err);
      } finally {
        setLoading(false);
      }
    };

    fetchAuthenticatedTenants();
  }, []);

  // Check if tenant exists and redirect
  const handleTenantSubmit = async () => {
    if (!tenantName.trim()) {
      setTenantError('Please enter a tenant name');
      return;
    }

    // Validate tenant name format (subdomain-safe)
    const tenantNameLower = tenantName.toLowerCase().trim();
    if (!/^[a-z0-9]([a-z0-9-]*[a-z0-9])?$/.test(tenantNameLower)) {
      setTenantError('Tenant name must be lowercase letters, numbers, and hyphens only');
      return;
    }

    setCheckingTenant(true);
    setTenantError(null);

    try {
      const response = await fetch(`/api/v1/tenant/check?name=${encodeURIComponent(tenantNameLower)}`);
      const data = await response.json();

      if (data.exists) {
        // Tenant exists, redirect to it
        window.location.href = buildTenantUrl(tenantNameLower);
      } else {
        setTenantError('Tenant not found. Please check the name and try again.');
      }
    } catch (err) {
      setTenantError('Failed to check tenant. Please try again.');
      console.error('Error checking tenant:', err);
    } finally {
      setCheckingTenant(false);
    }
  };

  // Navigate directly to an authenticated tenant
  const handleTenantSelect = (tenantName: string) => {
    window.location.href = buildTenantUrl(tenantName);
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: '60vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <div style={{ maxWidth: 600, margin: '0 auto', padding: '24px' }}>
      <Card>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <GlobalOutlined style={{ fontSize: 48, color: '#1890ff', marginBottom: 16 }} />
          <Title level={2} style={{ margin: 0 }}>Select Tenant</Title>
          <Paragraph type="secondary">
            Enter your organization's tenant name or select from your authenticated tenants.
          </Paragraph>
        </div>

        {error && (
          <Alert
            message={error}
            type="error"
            showIcon
            closable
            onClose={() => setError(null)}
            style={{ marginBottom: 24 }}
          />
        )}

        {/* Tenant name input */}
        <Form layout="vertical" onFinish={handleTenantSubmit}>
          <Form.Item
            label="Tenant Name"
            help={tenantError}
            validateStatus={tenantError ? 'error' : undefined}
          >
            <Space.Compact style={{ width: '100%' }}>
              <Input
                placeholder="your-organization"
                value={tenantName}
                onChange={(e) => {
                  setTenantName(e.target.value);
                  setTenantError(null);
                }}
                onPressEnter={handleTenantSubmit}
                suffix={<Text type="secondary">.{getBaseDomain()}</Text>}
                size="large"
              />
              <Button
                type="primary"
                icon={<LoginOutlined />}
                size="large"
                loading={checkingTenant}
                onClick={handleTenantSubmit}
              >
                Go
              </Button>
            </Space.Compact>
          </Form.Item>
        </Form>

        {/* Authenticated tenants list */}
        {authenticatedTenants.length > 0 && (
          <>
            <Divider>
              <Text type="secondary">
                <TeamOutlined /> Your Authenticated Tenants
              </Text>
            </Divider>

            <List
              dataSource={authenticatedTenants}
              renderItem={(tenant) => (
                <List.Item style={{ padding: '8px 0' }}>
                  <Button
                    type="default"
                    block
                    size="large"
                    onClick={() => handleTenantSelect(tenant.tenant_name)}
                    style={{ textAlign: 'left', height: 'auto', padding: '12px 16px' }}
                  >
                    <div>
                      <Text strong>{tenant.tenant_display_name}</Text>
                      <br />
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        {tenant.tenant_name}.{getBaseDomain()} - {tenant.user_email}
                      </Text>
                    </div>
                  </Button>
                </List.Item>
              )}
            />
          </>
        )}

        {authenticatedTenants.length === 0 && (
          <div style={{ textAlign: 'center', marginTop: 24 }}>
            <Text type="secondary">
              No authenticated sessions found. Enter a tenant name above to get started.
            </Text>
          </div>
        )}
      </Card>
    </div>
  );
}

export default TenantSelectionPage;
