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
  Spin,
  Popconfirm,
  Row,
  Col,
  Divider,
  Collapse,
  Switch,
} from 'antd';
import { DeleteOutlined, SettingOutlined, TableOutlined } from '@ant-design/icons';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useStorageConnectionTest } from './hooks/useStorageConnectionTest';
import StorageTestModal from './components/StorageTestModal';

const { Title, Paragraph } = Typography;

type StorageType = 's3' | 'postgres';

interface FormValues {
  name: string;
  type: StorageType;
  is_enabled: boolean;
  // S3 fields
  provider?: string;
  region?: string;
  bucket?: string;
  access_key_id?: string;
  secret_access_key?: string;
  // PostgreSQL fields
  connection_string?: string;
  user?: string;
  password?: string;
  host?: string;
  port?: string;
  database?: string;
  options?: string;
}

// PostgreSQL connection string utilities
interface PostgresConnectionParams {
  host?: string;
  port?: string;
  user?: string;
  password?: string;
  database?: string;
  options?: string;
}

function parsePostgresUrl(url: string): PostgresConnectionParams | null {
  if (!url) return null;

  try {
    // Handle postgres:// and postgresql:// schemes
    const normalized = url.replace(/^postgresql:\/\//, 'postgres://');
    if (!normalized.startsWith('postgres://')) return null;

    const urlObj = new URL(normalized);

    // Extract path (database name) - remove leading slash
    const database = urlObj.pathname.slice(1) || undefined;

    return {
      host: urlObj.hostname || undefined,
      port: urlObj.port || undefined,
      user: urlObj.username ? decodeURIComponent(urlObj.username) : undefined,
      password: urlObj.password ? decodeURIComponent(urlObj.password) : undefined,
      database,
      options: urlObj.search ? urlObj.search.slice(1) : undefined, // Remove leading ?
    };
  } catch {
    return null;
  }
}

function buildPostgresUrl(params: PostgresConnectionParams): string {
  const { host, port, user, password, database, options } = params;

  if (!host) return '';

  let url = 'postgres://';

  if (user) {
    url += encodeURIComponent(user);
    if (password) {
      url += ':' + encodeURIComponent(password);
    }
    url += '@';
  }

  url += host;

  if (port) {
    url += ':' + port;
  }

  url += '/' + (database || '');

  if (options) {
    url += '?' + options;
  }

  return url;
}

const typeOptions = [
  { value: 's3', label: 'S3' },
  { value: 'postgres', label: 'PostgreSQL' },
];

function StorageEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const [form] = Form.useForm<FormValues>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [storageType, setStorageType] = useState<StorageType>('s3');
  const [loadedData, setLoadedData] = useState<FormValues | null>(null);
  const [testModalOpen, setTestModalOpen] = useState(false);
  const [pendingFormValues, setPendingFormValues] = useState<FormValues | null>(null);
  const { state: testState, testPostgresConnection, cancel: cancelTest, reset: resetTest } = useStorageConnectionTest();
  const isNew = !uuid;

  const selectedType = Form.useWatch('type', form) || storageType;

  // Track if we're syncing to prevent loops
  const [isSyncing, setIsSyncing] = useState(false);

  // Handle connection string changes - parse and update individual fields
  const handleConnectionStringChange = useCallback(
    (value: string) => {
      if (isSyncing) return;

      const parsed = parsePostgresUrl(value);
      if (parsed) {
        setIsSyncing(true);
        form.setFieldsValue({
          host: parsed.host || '',
          port: parsed.port || '',
          user: parsed.user || '',
          password: parsed.password || '',
          database: parsed.database || '',
          options: parsed.options || '',
        });
        // Use setTimeout to ensure state updates complete
        setTimeout(() => setIsSyncing(false), 0);
      }
    },
    [form, isSyncing]
  );

  // Handle individual field changes - rebuild connection string
  const handleFieldChange = useCallback(() => {
    if (isSyncing) return;

    setIsSyncing(true);
    const values = form.getFieldsValue();
    const connStr = buildPostgresUrl({
      host: values.host,
      port: values.port,
      user: values.user,
      password: values.password,
      database: values.database,
      options: values.options,
    });
    form.setFieldValue('connection_string', connStr);
    setTimeout(() => setIsSyncing(false), 0);
  }, [form, isSyncing]);

  // Load existing storage for edit mode
  const loadStorage = useCallback(async () => {
    if (isNew) return;

    setLoading(true);

    // First, get the generic storage list to determine type
    const { data: storages, error: listError } = await client.GET('/storage');
    if (listError) {
      message.error('Failed to load storage');
      setLoading(false);
      return;
    }

    const storage = storages?.find((s) => s.uuid === uuid);
    if (!storage) {
      message.error('Storage not found');
      navigate(`/w/${slug}/storages`);
      return;
    }

    const sType = storage.type as StorageType;
    setStorageType(sType);

    // Now fetch the full details based on type
    let fullData: FormValues | null = null;

    switch (sType) {
      case 's3': {
        const { data } = await client.GET('/storage/s3/{uuid}', {
          params: { path: { uuid: uuid! } },
        });
        if (data) {
          fullData = {
            ...data,
            type: 's3',
            is_enabled: data.is_enabled ?? true,
          };
        }
        break;
      }
      case 'postgres': {
        const { data } = await client.GET('/storage/postgres/{uuid}', {
          params: { path: { uuid: uuid! } },
        });
        if (data) {
          // Build connection string from individual fields
          const connStr = buildPostgresUrl({
            host: data.host,
            port: data.port,
            user: data.user,
            password: data.password,
            database: data.database,
            options: data.options,
          });
          fullData = {
            ...data,
            type: 'postgres',
            is_enabled: data.is_enabled ?? true,
            connection_string: connStr,
          };
        }
        break;
      }
    }

    // Store loaded data in state - form values will be set after re-render
    setLoadedData(fullData);
    setLoading(false);
  }, [uuid, isNew, navigate, slug]);

  // Set form values after the component has rendered with the correct type
  useEffect(() => {
    if (loadedData) {
      // First, set just the type to trigger re-render with correct fields
      form.setFieldValue('type', loadedData.type);
    }
  }, [loadedData, form]);

  // Second effect: set all values after type is correctly set
  useEffect(() => {
    if (loadedData && storageType === loadedData.type) {
      const timer = setTimeout(() => {
        form.setFieldsValue(loadedData);
      }, 0);
      return () => clearTimeout(timer);
    }
  }, [loadedData, storageType, form]);

  useEffect(() => {
    loadStorage();
  }, [loadStorage]);

  // Extract save logic to separate function for reuse after test
  const saveStorage = async (values: FormValues) => {
    setSaving(true);

    try {
      if (isNew) {
        // Create new storage
        switch (values.type) {
          case 's3': {
            const { error } = await client.POST('/storage/s3', {
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                provider: values.provider || '',
                region: values.region || '',
                bucket: values.bucket || '',
                access_key_id: values.access_key_id || '',
                secret_access_key: values.secret_access_key || '',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'postgres': {
            const { data, error } = await client.POST('/storage/postgres', {
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                user: values.user,
                password: values.password,
                host: values.host,
                port: values.port,
                database: values.database,
                options: values.options,
                tables: [], // Tables are configured separately
              },
            });
            if (error) throw new Error(error.detail);
            // Redirect to tables page for newly created postgres storage
            if (data?.uuid) {
              message.success('Connection saved. Now configure target tables.');
              navigate(`/w/${slug}/storages/${data.uuid}/tables`);
              return;
            }
            break;
          }
        }
        message.success('Storage created');
      } else {
        // Update existing storage
        switch (storageType) {
          case 's3': {
            const { error } = await client.PUT('/storage/s3/{uuid}', {
              params: { path: { uuid: uuid! } },
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                provider: values.provider || '',
                region: values.region || '',
                bucket: values.bucket || '',
                access_key_id: values.access_key_id || '',
                secret_access_key: values.secret_access_key || '',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'postgres': {
            // Note: tables are updated separately via the tables page
            // We need to preserve existing tables when updating connection settings
            const { data: existingData } = await client.GET('/storage/postgres/{uuid}', {
              params: { path: { uuid: uuid! } },
            });
            const { error } = await client.PUT('/storage/postgres/{uuid}', {
              params: { path: { uuid: uuid! } },
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                user: values.user,
                password: values.password,
                host: values.host,
                port: values.port,
                database: values.database,
                options: values.options,
                tables: existingData?.tables || [],
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
        }
        message.success('Storage updated');
      }

      navigate(`/w/${slug}/storages`);
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'Failed to save storage');
    } finally {
      setSaving(false);
    }
  };

  const handleSubmit = async (values: FormValues) => {
    // Check if this is postgres that needs testing
    const needsTest = values.type === 'postgres';

    if (needsTest) {
      // Parse connection string if individual fields are empty
      // This handles the case where the user fills connection_string and immediately clicks Create
      let testParams = {
        user: values.user,
        password: values.password,
        host: values.host,
        port: values.port,
        database: values.database,
        options: values.options,
      };

      if (!values.host && values.connection_string) {
        const parsed = parsePostgresUrl(values.connection_string);
        if (parsed) {
          testParams = {
            user: parsed.user,
            password: parsed.password,
            host: parsed.host,
            port: parsed.port,
            database: parsed.database,
            options: parsed.options,
          };
          // Also update form values for saving later
          const updatedValues = { ...values, ...parsed };
          setPendingFormValues(updatedValues);
        } else {
          setPendingFormValues(values);
        }
      } else {
        setPendingFormValues(values);
      }

      // Open test modal and start test
      setTestModalOpen(true);
      await testPostgresConnection(testParams);
    } else {
      // No test needed (s3, hostfiles)
      await saveStorage(values);
    }
  };

  // Modal handlers
  const handleTestProceed = () => {
    setTestModalOpen(false);
    if (pendingFormValues) {
      saveStorage(pendingFormValues);
    }
    setPendingFormValues(null);
    resetTest();
  };

  const handleTestCancel = () => {
    cancelTest();
    setTestModalOpen(false);
    setPendingFormValues(null);
    resetTest();
  };

  const handleTestRetry = async () => {
    resetTest();
    if (pendingFormValues) {
      await testPostgresConnection({
        user: pendingFormValues.user,
        password: pendingFormValues.password,
        host: pendingFormValues.host,
        port: pendingFormValues.port,
        database: pendingFormValues.database,
        options: pendingFormValues.options,
      });
    }
  };

  const handleDelete = async () => {
    if (isNew) return;

    let endpoint: '/storage/s3/{uuid}' | '/storage/postgres/{uuid}';
    switch (storageType) {
      case 's3':
        endpoint = '/storage/s3/{uuid}';
        break;
      case 'postgres':
        endpoint = '/storage/postgres/{uuid}';
        break;
    }

    const { error } = await client.DELETE(endpoint, {
      params: { path: { uuid: uuid! } },
    });

    if (error) {
      message.error('Failed to delete storage');
      return;
    }

    message.success('Storage deleted');
    navigate(`/w/${slug}/storages`);
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
            {isNew ? 'Add Storage' : 'Edit Storage'}
          </Title>
        </Space>

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            type: 's3',
            is_enabled: true,
          }}
        >
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: 'Name is required' }]}
          >
            <Input placeholder="My Storage" />
          </Form.Item>

          <Form.Item
            name="type"
            label="Type"
            rules={[{ required: true, message: 'Type is required' }]}
          >
            <Select
              options={typeOptions}
              disabled={!isNew}
              onChange={(value) => setStorageType(value)}
            />
          </Form.Item>

          <Form.Item name="is_enabled" label="Enabled" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Divider />

          {/* S3 fields */}
          {selectedType === 's3' && (
            <>
              <Form.Item
                name="provider"
                label="Provider"
                rules={[{ required: true, message: 'Provider is required' }]}
              >
                <Input placeholder="AWS, Azure, MinIO endpoint, etc." />
              </Form.Item>
              <Form.Item
                name="region"
                label="Region"
                rules={[{ required: true, message: 'Region is required' }]}
              >
                <Input placeholder="us-east-1" />
              </Form.Item>
              <Form.Item
                name="bucket"
                label="Bucket"
                rules={[{ required: true, message: 'Bucket is required' }]}
              >
                <Input placeholder="my-bucket" />
              </Form.Item>
              <Form.Item
                name="access_key_id"
                label="Access Key ID"
                rules={[{ required: true, message: 'Access Key ID is required' }]}
              >
                <Input placeholder="AKIAIOSFODNN7EXAMPLE" />
              </Form.Item>
              <Form.Item
                name="secret_access_key"
                label="Secret Access Key"
                rules={[{ required: true, message: 'Secret Access Key is required' }]}
              >
                <Input.Password placeholder="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" />
              </Form.Item>
            </>
          )}

          {/* PostgreSQL fields */}
          {selectedType === 'postgres' && (
            <>
              <Form.Item
                name="connection_string"
                label="Connection String"
                extra="Example: postgres://user:password@localhost:5432/database?sslmode=require"
              >
                <Input
                  placeholder="postgres://user:password@host:5432/database"
                  onChange={(e) => handleConnectionStringChange(e.target.value)}
                />
              </Form.Item>

              <Collapse
                ghost
                items={[
                  {
                    key: 'manual',
                    label: (
                      <Space>
                        <SettingOutlined />
                        Manual Configuration
                      </Space>
                    ),
                    children: (
                      <>
                        <Form.Item name="host" label="Host">
                          <Input placeholder="localhost" onChange={handleFieldChange} />
                        </Form.Item>
                        <Form.Item name="port" label="Port">
                          <Input placeholder="5432" onChange={handleFieldChange} />
                        </Form.Item>
                        <Form.Item
                          name="database"
                          label="Database"
                          extra="Defaults to 'postgres' if not specified"
                        >
                          <Input placeholder="postgres" onChange={handleFieldChange} />
                        </Form.Item>
                        <Form.Item name="user" label="Username">
                          <Input placeholder="postgres" onChange={handleFieldChange} />
                        </Form.Item>
                        <Form.Item name="password" label="Password">
                          <Input.Password
                            placeholder="Database password"
                            onChange={handleFieldChange}
                          />
                        </Form.Item>
                        <Form.Item
                          name="options"
                          label="Connection Options"
                          extra="Additional connection options in URL query format"
                        >
                          <Input
                            placeholder="sslmode=require&connect_timeout=10"
                            onChange={handleFieldChange}
                          />
                        </Form.Item>
                      </>
                    ),
                  },
                ]}
              />

            </>
          )}

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={saving}>
                {isNew ? 'Create' : 'Save'}
              </Button>
              <Button onClick={() => navigate(`/w/${slug}/storages`)}>Cancel</Button>
              {!isNew && storageType === 'postgres' && (
                <Button
                  icon={<TableOutlined />}
                  onClick={() => navigate(`/w/${slug}/storages/${uuid}/tables`)}
                >
                  Configure Tables
                </Button>
              )}
              {!isNew && (
                <Popconfirm
                  title="Delete storage"
                  description="Are you sure you want to delete this storage? This action cannot be undone."
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
        <Card title="About Storages" size="small">
          <Paragraph>
            Storages define where MeshPump stores extracted data. Choose the type that best fits
            your infrastructure:
          </Paragraph>
          <Title level={5}>S3</Title>
          <Paragraph type="secondary">
            Store files in any S3-compatible object storage (AWS S3, MinIO, Azure Blob, etc.).
            Ideal for large volumes of files and attachments.
          </Paragraph>
          <Title level={5}>PostgreSQL</Title>
          <Paragraph type="secondary">
            Store data directly in a PostgreSQL database. Connect to any PostgreSQL instance
            to store structured data.
          </Paragraph>
        </Card>
      </Col>

      <StorageTestModal
        open={testModalOpen}
        state={testState}
        storageName={form.getFieldValue('name') || 'Storage'}
        onProceed={handleTestProceed}
        onCancel={handleTestCancel}
        onRetry={handleTestRetry}
      />
    </Row>
  );
}

export default StorageEdit;
