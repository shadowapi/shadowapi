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
  DatePicker,
  Radio,
  Descriptions,
} from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import type { components } from '../../api/v1';

const { Title, Paragraph, Text } = Typography;

type Pipeline = components['schemas']['pipeline'];
type Scheduler = components['schemas']['scheduler'];

// Common timezones
const TIMEZONES = [
  { value: 'UTC', label: 'UTC' },
  { value: 'America/New_York', label: 'Eastern Time (US)' },
  { value: 'America/Chicago', label: 'Central Time (US)' },
  { value: 'America/Denver', label: 'Mountain Time (US)' },
  { value: 'America/Los_Angeles', label: 'Pacific Time (US)' },
  { value: 'Europe/London', label: 'London' },
  { value: 'Europe/Paris', label: 'Paris' },
  { value: 'Europe/Berlin', label: 'Berlin' },
  { value: 'Europe/Moscow', label: 'Moscow' },
  { value: 'Asia/Tokyo', label: 'Tokyo' },
  { value: 'Asia/Shanghai', label: 'Shanghai' },
  { value: 'Asia/Singapore', label: 'Singapore' },
  { value: 'Australia/Sydney', label: 'Sydney' },
];

// Cron expression presets
const CRON_PRESETS = [
  { label: 'Every 5 minutes', value: '*/5 * * * *' },
  { label: 'Every 10 minutes', value: '*/10 * * * *' },
  { label: 'Every 30 minutes', value: '*/30 * * * *' },
  { label: 'Every hour', value: '0 * * * *' },
  { label: 'Every 6 hours', value: '0 */6 * * *' },
  { label: 'Every day at midnight', value: '0 0 * * *' },
  { label: 'Every day at 2 AM', value: '0 2 * * *' },
  { label: 'Every Sunday at 2 AM', value: '0 2 * * 0' },
  { label: 'Every Monday at 9 AM', value: '0 9 * * 1' },
  { label: 'First day of month', value: '0 0 1 * *' },
];

interface FormValues {
  pipeline_uuid: string;
  schedule_type: 'cron' | 'one_time';
  cron_expression: string;
  run_at: dayjs.Dayjs | null;
  timezone: string;
  is_enabled: boolean;
  is_paused: boolean;
}

function SchedulerEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const [form] = Form.useForm<FormValues>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [scheduler, setScheduler] = useState<Scheduler | null>(null);
  const [scheduleType, setScheduleType] = useState<'cron' | 'one_time'>('cron');
  const isNew = !uuid;

  // Load pipelines for dropdown
  const loadPipelines = useCallback(async () => {
    const { data, error } = await client.GET('/pipeline');
    if (error) {
      message.error('Failed to load pipelines');
      return;
    }
    const pipelinesData = data as { pipelines?: Pipeline[] } | undefined;
    setPipelines(pipelinesData?.pipelines || []);
  }, []);

  // Load existing scheduler for edit mode
  const loadScheduler = useCallback(async () => {
    if (isNew) return;

    setLoading(true);

    const { data, error } = await client.GET('/scheduler/{uuid}', {
      params: { path: { uuid: uuid! } },
    });

    if (error) {
      message.error('Failed to load scheduler');
      setLoading(false);
      return;
    }

    if (!data) {
      message.error('Scheduler not found');
      navigate(`/w/${slug}/schedulers`);
      return;
    }

    setScheduler(data);
    setScheduleType(data.schedule_type as 'cron' | 'one_time');

    form.setFieldsValue({
      pipeline_uuid: data.pipeline_uuid,
      schedule_type: data.schedule_type as 'cron' | 'one_time',
      cron_expression: data.cron_expression || '',
      run_at: data.run_at ? dayjs(data.run_at) : null,
      timezone: data.timezone || 'UTC',
      is_enabled: data.is_enabled ?? false,
      is_paused: data.is_paused ?? false,
    });

    setLoading(false);
  }, [uuid, isNew, navigate, slug, form]);

  useEffect(() => {
    loadPipelines();
    loadScheduler();
  }, [loadPipelines, loadScheduler]);

  const handleSubmit = async (values: FormValues) => {
    setSaving(true);

    const body: Scheduler = {
      pipeline_uuid: values.pipeline_uuid,
      schedule_type: values.schedule_type,
      cron_expression: values.schedule_type === 'cron' ? values.cron_expression : null,
      run_at: values.schedule_type === 'one_time' && values.run_at ? values.run_at.toISOString() : null,
      timezone: values.timezone,
      is_enabled: values.is_enabled,
      is_paused: values.is_paused,
    };

    try {
      if (isNew) {
        const { error } = await client.POST('/scheduler', { body });
        if (error) throw new Error((error as { detail?: string }).detail);
        message.success('Scheduler created');
      } else {
        const { error } = await client.PUT('/scheduler/{uuid}', {
          params: { path: { uuid: uuid! } },
          body,
        });
        if (error) throw new Error((error as { detail?: string }).detail);
        message.success('Scheduler updated');
      }

      navigate(`/w/${slug}/schedulers`);
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'Failed to save scheduler');
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async () => {
    if (isNew) return;

    const { error } = await client.DELETE('/scheduler/{uuid}', {
      params: { path: { uuid: uuid! } },
    });

    if (error) {
      message.error('Failed to delete scheduler');
      return;
    }

    message.success('Scheduler deleted');
    navigate(`/w/${slug}/schedulers`);
  };

  const handleScheduleTypeChange = (type: 'cron' | 'one_time') => {
    setScheduleType(type);
    form.setFieldsValue({ schedule_type: type });
  };

  const handlePresetSelect = (preset: string) => {
    form.setFieldsValue({ cron_expression: preset });
  };

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
        <Spin size="large" />
      </div>
    );
  }

  const pipelineOptions = pipelines.map((p) => ({
    value: p.uuid,
    label: p.name || p.uuid,
  }));

  const formatDate = (dateStr?: string) => {
    if (!dateStr) return 'Not scheduled yet';
    return new Date(dateStr).toLocaleString();
  };

  return (
    <Row gutter={24}>
      <Col xs={24} lg={16}>
        <Space style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
          <Title level={4} style={{ margin: 0 }}>
            {isNew ? 'Add Scheduler' : 'Edit Scheduler'}
          </Title>
        </Space>

        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            schedule_type: 'cron',
            timezone: 'UTC',
            is_enabled: true,
            is_paused: false,
            cron_expression: '0 * * * *',
          }}
        >
          <Form.Item
            name="pipeline_uuid"
            label="Pipeline"
            rules={[{ required: true, message: 'Pipeline is required' }]}
          >
            <Select
              options={pipelineOptions}
              placeholder="Select a pipeline to schedule"
              showSearch
              filterOption={(input, option) =>
                (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
              }
              disabled={!isNew}
            />
          </Form.Item>

          <Form.Item name="schedule_type" label="Schedule Type">
            <Radio.Group
              value={scheduleType}
              onChange={(e) => handleScheduleTypeChange(e.target.value)}
            >
              <Radio.Button value="cron">Recurring (Cron)</Radio.Button>
              <Radio.Button value="one_time">One-time</Radio.Button>
            </Radio.Group>
          </Form.Item>

          {scheduleType === 'cron' && (
            <>
              <Form.Item
                name="cron_expression"
                label="Cron Expression"
                rules={[{ required: true, message: 'Cron expression is required' }]}
                extra={
                  <Text type="secondary">
                    Format: minute hour day month day-of-week (e.g., &quot;0 9 * * 1&quot; for every
                    Monday at 9 AM)
                  </Text>
                }
              >
                <Input placeholder="0 * * * *" style={{ fontFamily: 'monospace' }} />
              </Form.Item>

              <Card size="small" title="Quick Presets" style={{ marginBottom: 16 }}>
                <Space wrap>
                  {CRON_PRESETS.map((preset) => (
                    <Button
                      key={preset.value}
                      size="small"
                      onClick={() => handlePresetSelect(preset.value)}
                    >
                      {preset.label}
                    </Button>
                  ))}
                </Space>
              </Card>
            </>
          )}

          {scheduleType === 'one_time' && (
            <Form.Item
              name="run_at"
              label="Run At"
              rules={[{ required: true, message: 'Run time is required' }]}
            >
              <DatePicker
                showTime
                format="YYYY-MM-DD HH:mm:ss"
                style={{ width: '100%' }}
                placeholder="Select date and time"
              />
            </Form.Item>
          )}

          <Form.Item name="timezone" label="Timezone">
            <Select
              options={TIMEZONES}
              showSearch
              filterOption={(input, option) =>
                (option?.label ?? '').toLowerCase().includes(input.toLowerCase())
              }
            />
          </Form.Item>

          <Divider />

          <Row gutter={24}>
            <Col span={12}>
              <Form.Item name="is_enabled" label="Enabled" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="is_paused"
                label="Paused"
                valuePropName="checked"
                tooltip="Temporarily pause the scheduler without disabling it"
              >
                <Switch />
              </Form.Item>
            </Col>
          </Row>

          {!isNew && scheduler && (
            <>
              <Divider />
              <Descriptions title="Scheduler Status" size="small" column={1}>
                <Descriptions.Item label="Next Run">
                  {formatDate(scheduler.next_run)}
                </Descriptions.Item>
                <Descriptions.Item label="Last Run">
                  {formatDate(scheduler.last_run)}
                </Descriptions.Item>
                <Descriptions.Item label="Created">
                  {formatDate(scheduler.created_at)}
                </Descriptions.Item>
                {scheduler.updated_at && (
                  <Descriptions.Item label="Updated">
                    {formatDate(scheduler.updated_at)}
                  </Descriptions.Item>
                )}
              </Descriptions>
            </>
          )}

          <Divider />

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={saving}>
                {isNew ? 'Create' : 'Save'}
              </Button>
              <Button onClick={() => navigate(`/w/${slug}/schedulers`)}>Cancel</Button>
              {!isNew && (
                <Popconfirm
                  title="Delete scheduler"
                  description="Are you sure you want to delete this scheduler? This action cannot be undone."
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
        <Card title="About Schedulers" size="small">
          <Paragraph>
            Schedulers control when your data pipelines run. You can set up recurring schedules
            using cron expressions or schedule a one-time execution.
          </Paragraph>
          <Title level={5}>Recurring (Cron)</Title>
          <Paragraph type="secondary">
            Use cron expressions to run pipelines on a regular schedule. The format is: minute,
            hour, day of month, month, day of week.
          </Paragraph>
          <Title level={5}>One-time</Title>
          <Paragraph type="secondary">
            Schedule a pipeline to run once at a specific date and time. Useful for one-off data
            syncs or testing.
          </Paragraph>
          <Title level={5}>Enabled vs Paused</Title>
          <Paragraph type="secondary">
            <strong>Enabled</strong>: Controls whether the scheduler is active. Disabled schedulers
            won&apos;t run at all.
          </Paragraph>
          <Paragraph type="secondary">
            <strong>Paused</strong>: Temporarily stops the scheduler while keeping it enabled.
            Useful for maintenance windows.
          </Paragraph>
        </Card>
      </Col>
    </Row>
  );
}

export default SchedulerEdit;
