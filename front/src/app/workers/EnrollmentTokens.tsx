import { useState, useEffect, useCallback, useRef } from 'react';
import {
  Typography,
  Space,
  Button,
  Table,
  Tag,
  message,
  Modal,
  Form,
  Input,
  DatePicker,
  Checkbox,
  Popconfirm,
  Alert,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  CopyOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import client from '../../api/client';
import type { components } from '../../api/v1';

const { Title } = Typography;

type WorkerEnrollmentToken = components['schemas']['worker_enrollment_token'];

const POLLING_INTERVAL = 10000; // 10 seconds

function getTokenStatus(token: WorkerEnrollmentToken): { status: string; color: string } {
  if (token.used_at) return { status: 'used', color: 'default' };
  if (token.expires_at && new Date(token.expires_at) < new Date()) return { status: 'expired', color: 'error' };
  return { status: 'available', color: 'success' };
}

interface CreateTokenFormValues {
  name: string;
  is_global: boolean;
  expires_at?: dayjs.Dayjs;
}

function EnrollmentTokens() {
  const [tokens, setTokens] = useState<WorkerEnrollmentToken[]>([]);
  const [loading, setLoading] = useState(true);
  const [createTokenModalOpen, setCreateTokenModalOpen] = useState(false);
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const [form] = Form.useForm<CreateTokenFormValues>();
  const pollingPausedRef = useRef(false);

  const loadTokens = useCallback(async () => {
    setLoading(true);
    const { data, error } = await client.GET('/workers/enrollment-tokens');
    if (error) {
      message.error('Failed to load tokens');
      setLoading(false);
      return;
    }
    setTokens(data || []);
    setLoading(false);
  }, []);

  const handleRevokeToken = async (uuid: string) => {
    const { error } = await client.DELETE('/workers/enrollment-tokens/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to revoke token');
      return;
    }
    message.success('Token revoked');
    loadTokens();
  };

  const handleCreateToken = async () => {
    try {
      const values = await form.validateFields();
      setCreating(true);

      const { data, error } = await client.POST('/workers/enrollment-tokens', {
        body: {
          name: values.name,
          is_global: values.is_global || false,
          expires_at: values.expires_at?.toISOString(),
        },
      });

      if (error) {
        message.error('Failed to create token');
        setCreating(false);
        return;
      }

      setCreatedToken(data?.token || null);
      message.success('Token created successfully');
      loadTokens();
    } finally {
      setCreating(false);
    }
  };

  const handleCloseModal = () => {
    setCreateTokenModalOpen(false);
    setCreatedToken(null);
    form.resetFields();
  };

  const handleCopyToken = () => {
    if (createdToken) {
      navigator.clipboard.writeText(createdToken);
      message.success('Token copied to clipboard');
    }
  };

  useEffect(() => {
    const poll = async () => {
      if (pollingPausedRef.current) return;
      loadTokens();
    };

    poll();
    const intervalId = setInterval(poll, POLLING_INTERVAL);

    return () => clearInterval(intervalId);
  }, [loadTokens]);

  useEffect(() => {
    const handleVisibilityChange = () => {
      pollingPausedRef.current = document.hidden;
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, []);

  useEffect(() => {
    pollingPausedRef.current = createTokenModalOpen;
  }, [createTokenModalOpen]);

  const columns: ColumnsType<WorkerEnrollmentToken> = [
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: 'Scope',
      key: 'scope',
      render: (_, record) =>
        record.is_global ? <Tag color="purple">Global</Tag> : <Tag color="green">Workspace</Tag>,
    },
    {
      title: 'Expires At',
      dataIndex: 'expires_at',
      key: 'expires_at',
      render: (v: string) => (v ? new Date(v).toLocaleString() : 'Never'),
    },
    {
      title: 'Status',
      key: 'status',
      render: (_, record) => {
        const { status, color } = getTokenStatus(record);
        return <Tag color={color}>{status}</Tag>;
      },
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => (v ? new Date(v).toLocaleDateString() : '-'),
    },
    {
      title: '',
      key: 'actions',
      width: 60,
      render: (_, record) => {
        const { status } = getTokenStatus(record);
        return status === 'available' ? (
          <Popconfirm
            title="Revoke this token?"
            description="The token will no longer be usable for enrollment."
            onConfirm={() => handleRevokeToken(record.uuid!)}
            okText="Revoke"
            okButtonProps={{ danger: true }}
          >
            <Button type="text" danger icon={<DeleteOutlined />} title="Revoke" />
          </Popconfirm>
        ) : null;
      },
    },
  ];

  return (
    <>
      <Space style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 16 }}>
        <Title level={4} style={{ margin: 0 }}>Enrollment Tokens</Title>
        <Space>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateTokenModalOpen(true)}>
            Create Token
          </Button>
          <Button icon={<ReloadOutlined />} onClick={loadTokens}>
            Refresh
          </Button>
        </Space>
      </Space>
      <Table
        columns={columns}
        dataSource={tokens}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        locale={{ emptyText: 'No enrollment tokens created' }}
      />

      <Modal
        title={createdToken ? 'Token Created' : 'Create Enrollment Token'}
        open={createTokenModalOpen}
        onCancel={handleCloseModal}
        footer={
          createdToken ? (
            <Button type="primary" onClick={handleCloseModal}>
              Done
            </Button>
          ) : undefined
        }
        onOk={createdToken ? undefined : handleCreateToken}
        confirmLoading={creating}
        okText="Create"
      >
        {createdToken ? (
          <Space direction="vertical" style={{ width: '100%' }}>
            <Alert
              type="warning"
              message="Save this token now!"
              description="This token will only be shown once. Copy it now and store it securely."
              showIcon
            />
            <Input.TextArea
              value={createdToken}
              readOnly
              autoSize
              style={{ fontFamily: 'monospace', marginTop: 16 }}
            />
            <Button icon={<CopyOutlined />} onClick={handleCopyToken} style={{ marginTop: 8 }}>
              Copy to Clipboard
            </Button>
          </Space>
        ) : (
          <Form form={form} layout="vertical">
            <Form.Item
              name="name"
              label="Worker Name"
              rules={[{ required: true, message: 'Please enter a worker name' }]}
            >
              <Input placeholder="worker-01" />
            </Form.Item>
            <Form.Item name="is_global" valuePropName="checked">
              <Checkbox>Global access (all workspaces)</Checkbox>
            </Form.Item>
            <Form.Item name="expires_at" label="Expiration (optional)">
              <DatePicker showTime style={{ width: '100%' }} />
            </Form.Item>
          </Form>
        )}
      </Modal>
    </>
  );
}

export default EnrollmentTokens;
