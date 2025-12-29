import { useEffect, useState, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router';
import {
  Form,
  Input,
  Button,
  Space,
  Typography,
  message,
  Card,
  Select,
  Switch,
  Spin,
  Popconfirm,
  Row,
  Col,
  Divider,
  Tag,
  Alert,
} from 'antd';
import { DeleteOutlined, CheckCircleOutlined, CloseCircleOutlined, LoginOutlined } from '@ant-design/icons';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import type { components } from '../../api/v1';

const { Title, Paragraph } = Typography;

type DatasourceType = 'email' | 'email_oauth' | 'telegram' | 'whatsapp' | 'linkedin';

type OAuth2Client = components['schemas']['oauth2_client'];

interface FormValues {
  name: string;
  type: DatasourceType;
  is_enabled: boolean;
  // Email IMAP fields
  email?: string;
  provider?: string;
  imap_server?: string;
  smtp_server?: string;
  smtp_tls?: boolean;
  password?: string;
  // Email OAuth fields
  oauth2_client_uuid?: string;
  // Telegram fields
  phone_number?: string;
  api_id?: number;
  api_hash?: string;
  // WhatsApp fields
  device_name?: string;
  // LinkedIn fields
  username?: string;
}

const typeOptions = [
  { value: 'email', label: 'Email IMAP' },
  { value: 'email_oauth', label: 'Email OAuth' },
  { value: 'telegram', label: 'Telegram' },
  { value: 'whatsapp', label: 'WhatsApp' },
  { value: 'linkedin', label: 'LinkedIn' },
];

function DataSourceEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const [form] = Form.useForm<FormValues>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [datasourceType, setDatasourceType] = useState<DatasourceType>('email');
  const [oauth2Clients, setOauth2Clients] = useState<OAuth2Client[]>([]);
  const [loadedData, setLoadedData] = useState<FormValues | null>(null);
  const [hasOAuthTokens, setHasOAuthTokens] = useState<boolean | null>(null);
  const [authLoading, setAuthLoading] = useState(false);
  const isNew = !uuid;

  const selectedType = Form.useWatch('type', form) || datasourceType;

  // Load OAuth2 clients for dropdown
  const loadOAuth2Clients = useCallback(async () => {
    const { data } = await client.GET('/oauth2/client');
    if (data?.clients) {
      setOauth2Clients(data.clients);
    }
  }, []);

  // Check if the datasource has OAuth tokens (for email_oauth type)
  const checkOAuthStatus = useCallback(async (datasourceUuid: string) => {
    const { data: tokens, error } = await client.GET('/oauth2/client/{datasource_uuid}/token', {
      params: { path: { datasource_uuid: datasourceUuid } },
    });
    if (error) {
      console.log('[DataSourceEdit] Failed to check OAuth status:', error);
      setHasOAuthTokens(false);
      return;
    }
    setHasOAuthTokens(tokens && tokens.length > 0);
  }, []);

  // Handle OAuth login for email_oauth datasources
  const handleOAuthLogin = async () => {
    if (!uuid || datasourceType !== 'email_oauth') return;
    setAuthLoading(true);

    const { data, error } = await client.POST('/oauth2/login', {
      body: {
        query: { datasource_uuid: [uuid], workspace_slug: [slug] },
      },
    });

    setAuthLoading(false);

    if (error) {
      message.error(error.detail || 'Failed to initiate OAuth login');
      return;
    }

    if (data?.auth_code_url) {
      window.location.href = data.auth_code_url;
    }
  };

  // Load existing datasource for edit mode
  const loadDatasource = useCallback(async () => {
    if (isNew) return;

    setLoading(true);

    // First, get the generic datasource to determine type
    const { data: datasources, error: listError } = await client.GET('/datasource');
    if (listError) {
      message.error('Failed to load data source');
      setLoading(false);
      return;
    }

    const ds = datasources?.find((d) => d.uuid === uuid);
    if (!ds) {
      message.error('Data source not found');
      navigate(`/w/${slug}/datasources`);
      return;
    }

    const dsType = ds.type as DatasourceType;
    setDatasourceType(dsType);

    // Now fetch the full details based on type
    let fullData: FormValues | null = null;

    switch (dsType) {
      case 'email': {
        const { data } = await client.GET('/datasource/email/{uuid}', {
          params: { path: { uuid: uuid! } },
        });
        if (data) {
          fullData = {
            ...data,
            type: 'email',
          };
        }
        break;
      }
      case 'email_oauth': {
        const { data } = await client.GET('/datasource/email_oauth/{uuid}', {
          params: { path: { uuid: uuid! } },
        });
        if (data) {
          fullData = {
            ...data,
            type: 'email_oauth',
          };
          // Check OAuth authentication status
          checkOAuthStatus(uuid!);
        }
        break;
      }
      case 'telegram': {
        const { data } = await client.GET('/datasource/telegram/{uuid}', {
          params: { path: { uuid: uuid! } },
        });
        if (data) {
          fullData = {
            ...data,
            type: 'telegram',
          };
        }
        break;
      }
      case 'whatsapp': {
        const { data } = await client.GET('/datasource/whatsapp/{uuid}', {
          params: { path: { uuid: uuid! } },
        });
        if (data) {
          fullData = {
            ...data,
            type: 'whatsapp',
          };
        }
        break;
      }
      case 'linkedin': {
        const { data } = await client.GET('/datasource/linkedin/{uuid}', {
          params: { path: { uuid: uuid! } },
        });
        if (data) {
          fullData = {
            ...data,
            type: 'linkedin',
          };
        }
        break;
      }
    }

    // Store loaded data in state - form values will be set after re-render
    // when the correct type-specific fields are rendered
    setLoadedData(fullData);
    setLoading(false);
  }, [uuid, isNew, navigate, slug, checkOAuthStatus]);

  // Set form values after the component has rendered with the correct type
  // This ensures type-specific fields exist in the DOM before setting values
  useEffect(() => {
    if (loadedData) {
      // First, set just the type to trigger re-render with correct fields
      form.setFieldValue('type', loadedData.type);
    }
  }, [loadedData, form]);

  // Second effect: set all values after type is correctly set and fields are rendered
  useEffect(() => {
    if (loadedData && datasourceType === loadedData.type) {
      // Use setTimeout to ensure DOM has updated with type-specific fields
      const timer = setTimeout(() => {
        form.setFieldsValue(loadedData);
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [loadedData, datasourceType, form]);

  useEffect(() => {
    loadOAuth2Clients();
    loadDatasource();
  }, [loadOAuth2Clients, loadDatasource]);

  const handleSubmit = async (values: FormValues) => {
    setSaving(true);

    try {
      if (isNew) {
        // Create new datasource
        switch (values.type) {
          case 'email': {
            const { error } = await client.POST('/datasource/email', {
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                email: values.email || '',
                provider: values.provider || '',
                imap_server: values.imap_server || '',
                smtp_server: values.smtp_server || '',
                smtp_tls: values.smtp_tls,
                password: values.password || '',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'email_oauth': {
            const { error } = await client.POST('/datasource/email_oauth', {
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                email: values.email || '',
                provider: (values.provider || 'gmail') as 'gmail' | 'google',
                oauth2_client_uuid: values.oauth2_client_uuid || '',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'telegram': {
            const { error } = await client.POST('/datasource/telegram', {
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                phone_number: values.phone_number || '',
                provider: values.provider || 'telegram',
                api_id: values.api_id || 0,
                api_hash: values.api_hash || '',
                password: values.password,
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'whatsapp': {
            const { error } = await client.POST('/datasource/whatsapp', {
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                phone_number: values.phone_number || '',
                provider: values.provider || 'whatsapp',
                device_name: values.device_name || '',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'linkedin': {
            const { error } = await client.POST('/datasource/linkedin', {
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                username: values.username || '',
                password: values.password || '',
                provider: values.provider || 'linkedin',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
        }
        message.success('Data source created');
      } else {
        // Update existing datasource
        switch (datasourceType) {
          case 'email': {
            const { error } = await client.PUT('/datasource/email/{uuid}', {
              params: { path: { uuid: uuid! } },
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                email: values.email || '',
                provider: values.provider || '',
                imap_server: values.imap_server || '',
                smtp_server: values.smtp_server || '',
                smtp_tls: values.smtp_tls,
                password: values.password || '',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'email_oauth': {
            const { error } = await client.PUT('/datasource/email_oauth/{uuid}', {
              params: { path: { uuid: uuid! } },
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                email: values.email || '',
                provider: (values.provider || 'gmail') as 'gmail' | 'google',
                oauth2_client_uuid: values.oauth2_client_uuid || '',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'telegram': {
            const { error } = await client.PUT('/datasource/telegram/{uuid}', {
              params: { path: { uuid: uuid! } },
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                phone_number: values.phone_number || '',
                provider: values.provider || 'telegram',
                api_id: values.api_id || 0,
                api_hash: values.api_hash || '',
                password: values.password,
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'whatsapp': {
            const { error } = await client.PUT('/datasource/whatsapp/{uuid}', {
              params: { path: { uuid: uuid! } },
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                phone_number: values.phone_number || '',
                provider: values.provider || 'whatsapp',
                device_name: values.device_name || '',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'linkedin': {
            const { error } = await client.PUT('/datasource/linkedin/{uuid}', {
              params: { path: { uuid: uuid! } },
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                username: values.username || '',
                password: values.password || '',
                provider: values.provider || 'linkedin',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
        }
        message.success('Data source updated');
      }

      navigate(`/w/${slug}/datasources`);
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'Failed to save data source');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (isNew) return;

    let endpoint: string;
    switch (datasourceType) {
      case 'email':
        endpoint = '/datasource/email/{uuid}';
        break;
      case 'email_oauth':
        endpoint = '/datasource/email_oauth/{uuid}';
        break;
      case 'telegram':
        endpoint = '/datasource/telegram/{uuid}';
        break;
      case 'whatsapp':
        endpoint = '/datasource/whatsapp/{uuid}';
        break;
      case 'linkedin':
        endpoint = '/datasource/linkedin/{uuid}';
        break;
    }

    const { error } = await client.DELETE(endpoint as '/datasource/email/{uuid}', {
      params: { path: { uuid: uuid! } },
    });

    if (error) {
      message.error('Failed to delete data source');
      return;
    }

    message.success('Data source deleted');
    navigate(`/w/${slug}/datasources`);
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <Row gutter={24}>
      <Col xs={24} lg={16}>
        <Space style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <Title level={4} style={{ margin: 0 }}>
            {isNew ? 'Add Data Source' : 'Edit Data Source'}
          </Title>
        </Space>

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            type: 'email',
            is_enabled: true,
            smtp_tls: true,
          }}
        >
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: 'Name is required' }]}
          >
            <Input placeholder="My Email Account" />
          </Form.Item>

          <Form.Item
            name="type"
            label="Type"
            rules={[{ required: true, message: 'Type is required' }]}
          >
            <Select
              options={typeOptions}
              disabled={!isNew}
              onChange={(value) => setDatasourceType(value)}
            />
          </Form.Item>

          <Form.Item name="is_enabled" label="Enabled" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Divider />

          {/* Email IMAP fields */}
          {selectedType === 'email' && (
            <>
              <Form.Item name="email" label="Email Address" rules={[{ type: 'email' }]}>
                <Input placeholder="user@example.com" />
              </Form.Item>
              <Form.Item name="provider" label="Provider">
                <Input placeholder="gmail, outlook, etc." />
              </Form.Item>
              <Form.Item name="imap_server" label="IMAP Server">
                <Input placeholder="imap.gmail.com:993" />
              </Form.Item>
              <Form.Item name="smtp_server" label="SMTP Server">
                <Input placeholder="smtp.gmail.com:587" />
              </Form.Item>
              <Form.Item name="smtp_tls" label="SMTP TLS" valuePropName="checked">
                <Switch />
              </Form.Item>
              <Form.Item name="password" label="Password">
                <Input.Password placeholder="App password or account password" />
              </Form.Item>
            </>
          )}

          {/* Email OAuth fields */}
          {selectedType === 'email_oauth' && (
            <>
              {/* Authentication Status for existing datasources */}
              {!isNew && hasOAuthTokens !== null && (
                <Alert
                  style={{ marginBottom: 16 }}
                  type={hasOAuthTokens ? 'success' : 'warning'}
                  showIcon
                  icon={hasOAuthTokens ? <CheckCircleOutlined /> : <CloseCircleOutlined />}
                  message={
                    <Space>
                      <span>
                        {hasOAuthTokens
                          ? 'OAuth authenticated'
                          : 'Not authenticated - authorization required'}
                      </span>
                      {!hasOAuthTokens && (
                        <Button
                          type="primary"
                          size="small"
                          icon={<LoginOutlined />}
                          loading={authLoading}
                          onClick={handleOAuthLogin}
                        >
                          Authorize
                        </Button>
                      )}
                    </Space>
                  }
                />
              )}

              <Form.Item name="email" label="Email Address" rules={[{ type: 'email' }]}>
                <Input placeholder="user@gmail.com" />
              </Form.Item>
              <Form.Item
                name="provider"
                label="Provider"
                rules={[{ required: true, message: 'Provider is required' }]}
              >
                <Select
                  placeholder="Select provider"
                  options={[
                    { value: 'gmail', label: 'Gmail / Google Workspace' },
                    { value: 'google', label: 'Google (generic)' },
                  ]}
                />
              </Form.Item>
              <Form.Item
                name="oauth2_client_uuid"
                label="OAuth2 Client"
                rules={[{ required: true, message: 'OAuth2 Client is required' }]}
                extra="Create OAuth2 clients in Data Sources → OAuth2 Credentials"
              >
                <Select
                  placeholder="Select OAuth2 Client"
                  options={oauth2Clients.map((c) => ({
                    value: c.uuid,
                    label: `${c.name} (${c.client_id})`,
                  }))}
                />
              </Form.Item>
            </>
          )}

          {/* Telegram fields */}
          {selectedType === 'telegram' && (
            <>
              <Form.Item name="phone_number" label="Phone Number">
                <Input placeholder="+1234567890" />
              </Form.Item>
              <Form.Item name="provider" label="Provider">
                <Input placeholder="telegram" />
              </Form.Item>
              <Form.Item name="api_id" label="API ID">
                <Input type="number" placeholder="12345678" />
              </Form.Item>
              <Form.Item name="api_hash" label="API Hash">
                <Input placeholder="abcdef1234567890" />
              </Form.Item>
              <Form.Item name="password" label="2FA Password (optional)">
                <Input.Password placeholder="Two-factor authentication password" />
              </Form.Item>
            </>
          )}

          {/* WhatsApp fields */}
          {selectedType === 'whatsapp' && (
            <>
              <Form.Item name="phone_number" label="Phone Number">
                <Input placeholder="+1234567890" />
              </Form.Item>
              <Form.Item name="provider" label="Provider">
                <Input placeholder="whatsapp" />
              </Form.Item>
              <Form.Item name="device_name" label="Device Name">
                <Input placeholder="My Phone" />
              </Form.Item>
            </>
          )}

          {/* LinkedIn fields */}
          {selectedType === 'linkedin' && (
            <>
              <Form.Item name="username" label="Username">
                <Input placeholder="linkedin_username" />
              </Form.Item>
              <Form.Item name="password" label="Password">
                <Input.Password placeholder="Account password" />
              </Form.Item>
              <Form.Item name="provider" label="Provider">
                <Input placeholder="linkedin" />
              </Form.Item>
            </>
          )}

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={saving}>
                {isNew ? 'Create' : 'Save'}
              </Button>
              <Button onClick={() => navigate(`/w/${slug}/datasources`)}>Cancel</Button>
              {!isNew && (
                <Popconfirm
                  title="Delete data source"
                  description="Are you sure you want to delete this data source? This action cannot be undone."
                  onConfirm={handleDelete}
                  okButtonProps={{ danger: true }}
                  okText="Delete"
                >
                  <Button danger icon={<DeleteOutlined />}>
                    Delete
                  </Button>
                </Popconfirm>
              )}
            </Space>
          </Form.Item>
        </Form>
      </Col>

      <Col xs={24} lg={8}>
        <Card title="About Data Sources" size="small">
          <Paragraph>
            Data sources connect MeshPump to your communication channels. Each type has specific
            configuration requirements:
          </Paragraph>
          <Title level={5}>Email IMAP</Title>
          <Paragraph type="secondary">
            Traditional email access using IMAP/SMTP protocols. Requires server addresses and an
            app password.
          </Paragraph>
          <Title level={5}>Email OAuth</Title>
          <Paragraph type="secondary">
            Secure email access using OAuth2. Requires an OAuth2 client configured in OAuth2
            Credentials.
          </Paragraph>
          <Title level={5}>Telegram</Title>
          <Paragraph type="secondary">
            Telegram access using the MTProto API. Requires API ID and Hash from{' '}
            <a href="https://my.telegram.org" target="_blank" rel="noopener noreferrer">
              my.telegram.org
            </a>
            .
          </Paragraph>
          <Title level={5}>WhatsApp</Title>
          <Paragraph type="secondary">
            WhatsApp access via web pairing. Device pairing is required after creation.
          </Paragraph>
          <Title level={5}>LinkedIn</Title>
          <Paragraph type="secondary">
            LinkedIn access using account credentials.
          </Paragraph>
        </Card>
      </Col>
    </Row>
  );
}

export default DataSourceEdit;
