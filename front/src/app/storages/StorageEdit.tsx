import { useEffect, useState, useCallback } from 'react';
import { useNavigate, useParams, useLocation } from 'react-router';
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
} from 'antd';
import { DeleteOutlined, RightOutlined, ExclamationCircleOutlined } from '@ant-design/icons';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { getStorageKey, getTables, setTables, clearTables, hasTables } from '../../lib/storage/storageTablesStore';
import type { components } from '../../api/v1';

type PostgresTable = components['schemas']['storage_postgres_table'];

const { Title, Paragraph } = Typography;

type StorageType = 's3' | 'postgres' | 'hostfiles';

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
  is_same_database?: boolean;
  user?: string;
  password?: string;
  host?: string;
  port?: string;
  options?: string;
  tables?: PostgresTable[];
  // Host Files fields
  path?: string;
}

const typeOptions = [
  { value: 's3', label: 'S3' },
  { value: 'postgres', label: 'PostgreSQL' },
  { value: 'hostfiles', label: 'Host Files' },
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
  const [tablesCount, setTablesCount] = useState(0);
  const [originalTablesJson, setOriginalTablesJson] = useState<string>('[]');
  const isNew = !uuid;
  const storageKey = getStorageKey(uuid);

  // Check if tables have unsaved changes
  const hasTableChanges = (() => {
    const currentTables = getTables(storageKey);
    const currentJson = JSON.stringify(currentTables);
    return currentJson !== originalTablesJson;
  })();

  const selectedType = Form.useWatch('type', form) || storageType;
  const isSameDatabase = Form.useWatch('is_same_database', form);

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
          fullData = {
            ...data,
            type: 'postgres',
            is_enabled: data.is_enabled ?? true,
          };
          // Save tables to sessionStorage (only if not already present from previous edit)
          const tables = data.tables || [];
          setOriginalTablesJson(JSON.stringify(tables));
          if (!hasTables(storageKey)) {
            setTables(storageKey, tables);
          }
          setTablesCount(getTables(storageKey).length);
        }
        break;
      }
      case 'hostfiles': {
        const { data } = await client.GET('/storage/hostfiles/{uuid}', {
          params: { path: { uuid: uuid! } },
        });
        if (data) {
          fullData = {
            ...data,
            type: 'hostfiles',
            is_enabled: data.is_enabled ?? true,
          };
        }
        break;
      }
    }

    // Store loaded data in state - form values will be set after re-render
    setLoadedData(fullData);
    setLoading(false);
  }, [uuid, isNew, navigate, slug, storageKey]);

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

  // Update tables count when returning from tables page or on initial render
  // Location.key changes on each navigation, triggering refresh when returning from tables page
  const location = useLocation();

  useEffect(() => {
    if (selectedType === 'postgres') {
      setTablesCount(getTables(storageKey).length);
    }
  }, [selectedType, storageKey, location.key]);

  const handleSubmit = async (values: FormValues) => {
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
            const { error } = await client.POST('/storage/postgres', {
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                is_same_database: values.is_same_database,
                user: values.user,
                password: values.password,
                host: values.host,
                port: values.port,
                options: values.options,
                tables: getTables(storageKey),
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'hostfiles': {
            const { error } = await client.POST('/storage/hostfiles', {
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                path: values.path || '',
              },
            });
            if (error) throw new Error(error.detail);
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
            const { error } = await client.PUT('/storage/postgres/{uuid}', {
              params: { path: { uuid: uuid! } },
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                is_same_database: values.is_same_database,
                user: values.user,
                password: values.password,
                host: values.host,
                port: values.port,
                options: values.options,
                tables: getTables(storageKey),
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
          case 'hostfiles': {
            const { error } = await client.PUT('/storage/hostfiles/{uuid}', {
              params: { path: { uuid: uuid! } },
              body: {
                name: values.name,
                is_enabled: values.is_enabled,
                path: values.path || '',
              },
            });
            if (error) throw new Error(error.detail);
            break;
          }
        }
        message.success('Storage updated');
      }

      clearTables(storageKey);
      navigate(`/w/${slug}/storages`);
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'Failed to save storage');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (isNew) return;

    let endpoint: '/storage/s3/{uuid}' | '/storage/postgres/{uuid}' | '/storage/hostfiles/{uuid}';
    switch (storageType) {
      case 's3':
        endpoint = '/storage/s3/{uuid}';
        break;
      case 'postgres':
        endpoint = '/storage/postgres/{uuid}';
        break;
      case 'hostfiles':
        endpoint = '/storage/hostfiles/{uuid}';
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
            is_same_database: false,
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
                name="is_same_database"
                label="Use Application Database"
                valuePropName="checked"
                extra="Use the same PostgreSQL database as the application"
              >
                <Switch />
              </Form.Item>

              {!isSameDatabase && (
                <>
                  <Form.Item name="host" label="Host">
                    <Input placeholder="localhost" />
                  </Form.Item>
                  <Form.Item name="port" label="Port">
                    <Input placeholder="5432" />
                  </Form.Item>
                  <Form.Item name="user" label="Username">
                    <Input placeholder="postgres" />
                  </Form.Item>
                  <Form.Item name="password" label="Password">
                    <Input.Password placeholder="Database password" />
                  </Form.Item>
                  <Form.Item
                    name="options"
                    label="Connection Options"
                    extra="Additional connection options in URL query format"
                  >
                    <Input placeholder="sslmode=require&connect_timeout=10" />
                  </Form.Item>
                </>
              )}

              <Divider />

              <Form.Item
                label={
                  <Space>
                    Target Tables
                    {hasTableChanges && (
                      <Typography.Text type="warning" style={{ fontSize: 12, fontWeight: 'normal' }}>
                        <ExclamationCircleOutlined /> unsaved changes
                      </Typography.Text>
                    )}
                  </Space>
                }
              >
                <Button
                  onClick={() => navigate(`/w/${slug}/storages/${uuid || 'new'}/tables`)}
                  style={{ width: '100%', textAlign: 'left' }}
                >
                  <Space style={{ width: '100%', justifyContent: 'space-between' }}>
                    <span>{tablesCount} table{tablesCount !== 1 ? 's' : ''} defined</span>
                    <RightOutlined />
                  </Space>
                </Button>
              </Form.Item>
            </>
          )}

          {/* Host Files fields */}
          {selectedType === 'hostfiles' && (
            <>
              <Form.Item
                name="path"
                label="Path"
                rules={[{ required: true, message: 'Path is required' }]}
                extra="Absolute or relative filesystem path for file storage"
              >
                <Input placeholder="/var/data/files" />
              </Form.Item>
            </>
          )}

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={saving}>
                {isNew ? 'Create' : 'Save'}
              </Button>
              <Button onClick={() => {
                clearTables(storageKey);
                navigate(`/w/${slug}/storages`);
              }}>Cancel</Button>
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
            Store data directly in a PostgreSQL database. Can use the application's own database
            or connect to an external instance.
          </Paragraph>
          <Title level={5}>Host Files</Title>
          <Paragraph type="secondary">
            Store files directly on the host filesystem. Simple setup but requires local storage
            access.
          </Paragraph>
        </Card>
      </Col>
    </Row>
  );
}

export default StorageEdit;
