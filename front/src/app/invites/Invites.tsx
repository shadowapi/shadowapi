import { useEffect, useState, useCallback } from 'react';
import { Table, Button, Space, Typography, message, Tag, Modal, Form, Input, Select, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined, MailOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import type { components } from '../../api/v1';

const { Title } = Typography;

type UserInvite = components['schemas']['user_invite'];

function Invites() {
  const { workspace } = useWorkspace();
  const workspaceUUID = workspace?.uuid;
  const [loading, setLoading] = useState(true);
  const [invites, setInvites] = useState<UserInvite[]>([]);
  const [modalOpen, setModalOpen] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [form] = Form.useForm();

  const loadInvites = useCallback(async () => {
    if (!workspaceUUID) return;
    setLoading(true);
    const { data, error } = await client.GET('/workspace/{uuid}/invites', {
      params: { path: { uuid: workspaceUUID } },
    });
    if (error) {
      message.error('Failed to load invites');
      setLoading(false);
      return;
    }
    setInvites(data || []);
    setLoading(false);
  }, [workspaceUUID]);

  useEffect(() => {
    loadInvites();
  }, [loadInvites]);

  const handleCreate = async (values: { email: string; role: 'admin' | 'member' }) => {
    if (!workspaceUUID) return;
    setSubmitting(true);
    const { error } = await client.POST('/workspace/{uuid}/invites', {
      params: { path: { uuid: workspaceUUID } },
      body: values,
    });
    setSubmitting(false);
    if (error) {
      if (error.detail?.includes('already exists')) {
        message.error('An invite already exists for this email');
      } else if (error.detail?.includes('already a member')) {
        message.error('This user is already a member of this workspace');
      } else {
        message.error(error.detail || 'Failed to send invite');
      }
      return;
    }
    message.success('Invite sent successfully');
    setModalOpen(false);
    form.resetFields();
    loadInvites();
  };

  const handleDelete = async (inviteUUID: string) => {
    if (!workspaceUUID) return;
    const { error } = await client.DELETE('/workspace/{uuid}/invites/{invite_uuid}', {
      params: { path: { uuid: workspaceUUID, invite_uuid: inviteUUID } },
    });
    if (error) {
      message.error('Failed to cancel invite');
      return;
    }
    message.success('Invite cancelled');
    loadInvites();
  };

  const columns: ColumnsType<UserInvite> = [
    {
      title: 'Email',
      dataIndex: 'email',
      key: 'email',
      render: (email: string) => (
        <Space>
          <MailOutlined />
          {email}
        </Space>
      ),
    },
    {
      title: 'Role',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => (
        <Tag color={role === 'admin' ? 'blue' : 'default'}>
          {role}
        </Tag>
      ),
    },
    {
      title: 'Invited By',
      key: 'invited_by',
      render: (_, record) => record.invited_by_name || record.invited_by_email || '-',
    },
    {
      title: 'Expires',
      dataIndex: 'expires_at',
      key: 'expires_at',
      render: (value: string) => {
        if (!value) return '-';
        const date = new Date(value);
        const now = new Date();
        const hoursLeft = Math.round((date.getTime() - now.getTime()) / (1000 * 60 * 60));
        if (hoursLeft < 0) return <Tag color="red">Expired</Tag>;
        if (hoursLeft < 6) return <Tag color="orange">{hoursLeft}h left</Tag>;
        return date.toLocaleDateString();
      },
    },
    {
      title: '',
      key: 'actions',
      width: 60,
      render: (_, record) => (
        <Popconfirm
          title="Cancel invite"
          description="Are you sure you want to cancel this invite?"
          onConfirm={() => handleDelete(record.uuid!)}
          okButtonProps={{ danger: true }}
          okText="Cancel Invite"
        >
          <Button type="text" danger icon={<DeleteOutlined />} />
        </Popconfirm>
      ),
    },
  ];

  return (
    <>
      <Space style={{ marginBottom: 16, display: 'flex', justifyContent: 'space-between' }}>
        <Title level={4} style={{ margin: 0 }}>Pending Invites</Title>
        <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
          Invite User
        </Button>
      </Space>

      <Table
        columns={columns}
        dataSource={invites}
        rowKey="uuid"
        loading={loading}
        pagination={false}
        locale={{ emptyText: 'No pending invites' }}
      />

      <Modal
        title="Invite User to Workspace"
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        footer={null}
        destroyOnClose
      >
        <Form form={form} layout="vertical" onFinish={handleCreate}>
          <Form.Item
            name="email"
            label="Email Address"
            rules={[
              { required: true, message: 'Please enter an email' },
              { type: 'email', message: 'Please enter a valid email' },
            ]}
          >
            <Input placeholder="user@example.com" />
          </Form.Item>

          <Form.Item
            name="role"
            label="Role"
            initialValue="member"
            rules={[{ required: true }]}
          >
            <Select>
              <Select.Option value="member">Member</Select.Option>
              <Select.Option value="admin">Admin</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalOpen(false)}>Cancel</Button>
              <Button type="primary" htmlType="submit" loading={submitting}>
                Send Invite
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}

export default Invites;
