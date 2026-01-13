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
  Alert,
} from 'antd';
import { DeleteOutlined, SettingOutlined, NodeIndexOutlined } from '@ant-design/icons';
import { Link } from 'react-router';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import type { components } from '../../api/v1';
import TestConnectionModal from './components/TestConnectionModal';
import { useConnectionTest } from './components/useConnectionTest';

const { Title, Paragraph } = Typography;

type Datasource = components['schemas']['datasource'];
type Storage = components['schemas']['storage'];
type MapperFieldMapping = components['schemas']['mapper_field_mapping'];
type MapperConfig = components['schemas']['mapper_config'];
type Pipeline = components['schemas']['pipeline'];
type RegisteredWorker = components['schemas']['registered_worker'];

interface FormValues {
  name: string;
  datasource_uuid: string;
  storage_uuid: string;
  worker_uuid: string | null;
  is_enabled: boolean;
}

function PipelineEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const [form] = Form.useForm<FormValues>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [datasources, setDatasources] = useState<Datasource[]>([]);
  const [storages, setStorages] = useState<Storage[]>([]);
  const [workers, setWorkers] = useState<RegisteredWorker[]>([]);
  const [mapperMappings, setMapperMappings] = useState<MapperFieldMapping[]>([]);
  const [testModalOpen, setTestModalOpen] = useState(false);
  const [pendingFormValues, setPendingFormValues] = useState<FormValues | null>(null);
  const { state: testState, startTest, cancelTest, reset: resetTest } = useConnectionTest();
  const isNew = !uuid;
  const hasWorkers = workers.length > 0;


  // Load datasources, storages, and workers for dropdowns
  const loadOptions = useCallback(async () => {
    const [dsRes, stRes, workersRes] = await Promise.all([
      client.GET('/datasource'),
      client.GET('/storage'),
      client.GET('/workers'),
    ]);

    if (dsRes.data) {
      setDatasources(dsRes.data);
    }
    if (stRes.data) {
      setStorages(stRes.data);
    }
    if (workersRes.data) {
      setWorkers(workersRes.data);
    }
  }, []);

  // Load existing pipeline for edit mode
  const loadPipeline = useCallback(async () => {
    if (isNew) return;

    setLoading(true);

    const { data, error } = await client.GET('/pipeline/{uuid}', {
      params: { path: { uuid: uuid! } },
    });

    if (error) {
      message.error('Failed to load pipeline');
      setLoading(false);
      return;
    }

    if (!data) {
      message.error('Pipeline not found');
      navigate(`/w/${slug}/pipelines`);
      return;
    }

    form.setFieldsValue({
      name: data.name,
      datasource_uuid: data.datasource_uuid,
      storage_uuid: data.storage_uuid,
      worker_uuid: data.worker_uuid ?? null,
      is_enabled: data.is_enabled ?? false,
    });

    // Extract mapper config from flow nodes if present
    const flow = data.flow as Pipeline['flow'];
    if (flow?.nodes) {
      const mapperNode = flow.nodes.find((n) => n.type === 'mapper');
      if (mapperNode?.data?.config) {
        const config = mapperNode.data.config as MapperConfig;
        setMapperMappings(config.mappings || []);
      }
    }

    setLoading(false);
  }, [uuid, isNew, navigate, slug, form]);

  useEffect(() => {
    loadOptions();
    loadPipeline();
  }, [loadOptions, loadPipeline]);

  // Actual save logic extracted to allow calling after test
  const savePipeline = async (values: FormValues) => {
    setSaving(true);

    // Build mapper config
    const mapperConfig: MapperConfig = {
      version: '1.0',
      mappings: mapperMappings,
    };

    // Build flow with mapper node
    const flow: Pipeline['flow'] = {
      nodes: [
        {
          id: 'mapper-1',
          type: 'mapper',
          position: { x: 0, y: 0 },
          data: {
            label: 'Mapper',
            config: mapperConfig,
          },
        },
      ],
      edges: [],
    };

    try {
      if (isNew) {
        const { error } = await client.POST('/pipeline', {
          body: {
            name: values.name,
            datasource_uuid: values.datasource_uuid,
            storage_uuid: values.storage_uuid,
            worker_uuid: values.worker_uuid,
            is_enabled: values.is_enabled,
            flow,
          },
        });
        if (error) throw new Error((error as { detail?: string }).detail);
        message.success('Pipeline created');
      } else {
        const { error } = await client.PUT('/pipeline/{uuid}', {
          params: { path: { uuid: uuid! } },
          body: {
            name: values.name,
            datasource_uuid: values.datasource_uuid,
            storage_uuid: values.storage_uuid,
            worker_uuid: values.worker_uuid,
            is_enabled: values.is_enabled,
            flow,
          },
        });
        if (error) throw new Error((error as { detail?: string }).detail);
        message.success('Pipeline updated');
      }

      navigate(`/w/${slug}/pipelines`);
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'Failed to save pipeline');
    } finally {
      setSaving(false);
    }
  };

  const handleSubmit = async (values: FormValues) => {
    const ds = datasources.find((d) => d.uuid === values.datasource_uuid);
    const st = storages.find((s) => s.uuid === values.storage_uuid);

    // Determine if tests are needed
    const needsDatasourceTest = ds?.type === 'email_oauth';
    const needsStorageTest = st?.type === 'postgres';

    if (needsDatasourceTest || needsStorageTest) {
      // Store form values for later save
      setPendingFormValues(values);

      // Open modal and start tests
      setTestModalOpen(true);
      await startTest(
        {
          uuid: values.datasource_uuid,
          type: ds?.type || '',
          name: ds?.name || '',
        },
        {
          uuid: values.storage_uuid,
          type: st?.type || '',
          name: st?.name || '',
        }
      );
    } else {
      // No tests needed - save directly
      await savePipeline(values);
    }
  };

  // Modal handlers
  const handleTestProceed = () => {
    setTestModalOpen(false);
    if (pendingFormValues) {
      savePipeline(pendingFormValues);
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
      const ds = datasources.find((d) => d.uuid === pendingFormValues.datasource_uuid);
      const st = storages.find((s) => s.uuid === pendingFormValues.storage_uuid);
      await startTest(
        {
          uuid: pendingFormValues.datasource_uuid,
          type: ds?.type || '',
          name: ds?.name || '',
        },
        {
          uuid: pendingFormValues.storage_uuid,
          type: st?.type || '',
          name: st?.name || '',
        }
      );
    }
  };

  const handleDelete = async () => {
    if (isNew) return;

    const { error } = await client.DELETE('/pipeline/{uuid}', {
      params: { path: { uuid: uuid! } },
    });

    if (error) {
      message.error('Failed to delete pipeline');
      return;
    }

    message.success('Pipeline deleted');
    navigate(`/w/${slug}/pipelines`);
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  const datasourceOptions = datasources.map((ds) => ({
    value: ds.uuid,
    label: `${ds.name} (${ds.type})`,
  }));

  const storageOptions = storages.map((st) => ({
    value: st.uuid,
    label: `${st.name} (${st.type})`,
  }));

  const workerOptions = [
    { value: '', label: 'Auto (any available worker)' },
    ...workers.map((w) => ({
      value: w.uuid,
      label: `${w.name} (${w.status})${w.is_global ? ' - Global' : ''}`,
      disabled: w.status === 'offline',
    })),
  ];

  return (
    <Row gutter={24}>
      <Col xs={24} lg={16}>
        <Space style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <Title level={4} style={{ margin: 0 }}>
            {isNew ? 'Add Pipeline' : 'Edit Pipeline'}
          </Title>
        </Space>

        {!hasWorkers && isNew && (
          <Alert
            message="No workers available"
            description={
              <>
                You need to register at least one worker before creating a pipeline. Workers execute
                your data sync jobs.{' '}
                <Link to={`/w/${slug}/workers`}>
                  <Button type="link" size="small" icon={<SettingOutlined />} style={{ padding: 0 }}>
                    Go to Workers
                  </Button>
                </Link>
              </>
            }
            type="warning"
            showIcon
            style={{ marginBottom: 16 }}
          />
        )}

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            worker_uuid: null,
            is_enabled: true,
          }}
        >
          <Form.Item
            name="name"
            label="Name"
            rules={[{ required: true, message: 'Name is required' }]}
          >
            <Input placeholder="My Pipeline" />
          </Form.Item>

          <Form.Item
            name="datasource_uuid"
            label="Data Source"
            rules={[{ required: true, message: 'Data source is required' }]}
          >
            <Select
              options={datasourceOptions}
              placeholder="Select a data source"
              showSearch
              filterOption={(input, option) =>
                (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
              }
              onChange={() => {
                // Clear mappings when datasource changes to avoid invalid field selections
                setMapperMappings([]);
              }}
            />
          </Form.Item>

          <Form.Item
            name="storage_uuid"
            label="Storage"
            rules={[{ required: true, message: 'Storage is required' }]}
          >
            <Select
              options={storageOptions}
              placeholder="Select a storage"
              showSearch
              filterOption={(input, option) =>
                (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
              }
            />
          </Form.Item>

          <Form.Item
            name="worker_uuid"
            label="Worker"
            tooltip="Assign a specific worker to execute this pipeline, or leave as Auto for any available worker"
          >
            <Select
              options={workerOptions}
              placeholder="Auto (any available worker)"
              allowClear
              showSearch
              filterOption={(input, option) =>
                (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
              }
            />
          </Form.Item>

          {!isNew && (
            <Form.Item label="Pipeline Flow">
              <Space>
                <Button
                  type="default"
                  icon={<NodeIndexOutlined />}
                  onClick={() => navigate(`/w/${slug}/pipelines/${uuid}/flow`)}
                >
                  Edit Flow
                </Button>
                {mapperMappings.length > 0 && (
                  <Typography.Text type="secondary">
                    {mapperMappings.length} mapping{mapperMappings.length !== 1 ? 's' : ''} configured
                  </Typography.Text>
                )}
              </Space>
            </Form.Item>
          )}

          <Divider />

          <Form.Item name="is_enabled" label="Enabled" valuePropName="checked">
            <Switch />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={saving}
                disabled={isNew && !hasWorkers}
              >
                {isNew ? 'Create' : 'Save'}
              </Button>
              <Button onClick={() => navigate(`/w/${slug}/pipelines`)}>Cancel</Button>
              {!isNew && (
                <Popconfirm
                  title="Delete pipeline"
                  description="Are you sure you want to delete this pipeline? This action cannot be undone."
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
        <Card title="About Pipelines" size="small">
          <Paragraph>
            Pipelines connect a data source to a storage destination, defining how data flows
            through MeshPump.
          </Paragraph>
          <Title level={5}>Data Source</Title>
          <Paragraph type="secondary">
            The source of messages and data. Can be an email account, Telegram, WhatsApp, or
            LinkedIn connection.
          </Paragraph>
          <Title level={5}>Storage</Title>
          <Paragraph type="secondary">
            Where extracted data will be saved. Choose from S3, PostgreSQL, or host filesystem
            storage.
          </Paragraph>
          <Title level={5}>Worker</Title>
          <Paragraph type="secondary">
            Assign a specific worker to execute this pipeline, or leave as &quot;Auto&quot; to let
            any available worker process it.
          </Paragraph>
          <Title level={5}>Field Mappings</Title>
          <Paragraph type="secondary">
            Define how source fields from messages and contacts map to your storage tables. Each
            mapping can include a transformation like lowercase or trim.
          </Paragraph>
          <Title level={5}>Enabled</Title>
          <Paragraph type="secondary">
            When enabled, the pipeline can be triggered by schedulers or run manually. Disabled
            pipelines are paused.
          </Paragraph>
        </Card>
      </Col>

      <TestConnectionModal
        open={testModalOpen}
        state={testState}
        onProceed={handleTestProceed}
        onCancel={handleTestCancel}
        onRetry={handleTestRetry}
      />
    </Row>
  );
}

export default PipelineEdit;
