import { useEffect, useState, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router';
import {
  Form,
  Input,
  Button,
  Space,
  Typography,
  message,
  Checkbox,
  Popconfirm,
  Result,
  Spin,
  Divider,
  Table,
  Modal,
  Select,
  Tag,
} from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth } from '../../lib/auth';
import type { components } from '../../api/v1';

const { Title } = Typography;

type User = components['schemas']['user'];
type RBACRole = components['schemas']['rbac_role'];
type RBACRoleAssignment = components['schemas']['rbac_role_assignment'];

interface UserFormValues {
  email: string;
  first_name: string;
  last_name: string;
  password?: string;
  is_enabled: boolean;
  is_admin: boolean;
}

function UserEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const isNew = !uuid;
  const [form] = Form.useForm<UserFormValues>();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

  // Role assignment state
  const [roleAssignments, setRoleAssignments] = useState<RBACRoleAssignment[]>([]);
  const [rolesLoading, setRolesLoading] = useState(false);
  const [assignModalVisible, setAssignModalVisible] = useState(false);
  const [availableRoles, setAvailableRoles] = useState<RBACRole[]>([]);
  const [assigning, setAssigning] = useState(false);
  const [selectedRole, setSelectedRole] = useState<string | undefined>();
  const [selectedDomain, setSelectedDomain] = useState<string>('global');

  // Load user data
  useEffect(() => {
    if (!isNew && uuid && currentUser?.is_admin) {
      const loadUser = async () => {
        setLoading(true);
        const { data, error } = await client.GET('/user/{uuid}', {
          params: { path: { uuid } },
        });
        if (error) {
          message.error('Failed to load user');
          navigate(`/w/${slug}/users`);
          return;
        }
        form.setFieldsValue({
          email: data.email,
          first_name: data.first_name,
          last_name: data.last_name,
          is_enabled: data.is_enabled ?? true,
          is_admin: data.is_admin ?? false,
        });
        setLoading(false);
      };
      loadUser();
    }
  }, [uuid, isNew, form, slug, navigate, currentUser?.is_admin]);

  // Load user's role assignments
  const loadRoleAssignments = useCallback(async () => {
    if (!uuid) return;
    setRolesLoading(true);
    const { data, error } = await client.GET('/rbac/user/{user_uuid}/roles', {
      params: { path: { user_uuid: uuid } },
    });
    if (error) {
      message.error('Failed to load role assignments');
      setRolesLoading(false);
      return;
    }
    setRoleAssignments(data?.roles || []);
    setRolesLoading(false);
  }, [uuid]);

  useEffect(() => {
    if (!isNew && uuid && currentUser?.is_admin) {
      loadRoleAssignments();
    }
  }, [uuid, isNew, loadRoleAssignments, currentUser?.is_admin]);

  // Load available roles for assignment modal
  const loadAvailableRoles = async () => {
    const { data, error } = await client.GET('/rbac/role');
    if (error) {
      message.error('Failed to load roles');
      return;
    }
    setAvailableRoles(data || []);
  };

  const openAssignModal = () => {
    loadAvailableRoles();
    setSelectedRole(undefined);
    setSelectedDomain(slug); // Default to current workspace
    setAssignModalVisible(true);
  };

  const handleAssignRole = async () => {
    if (!uuid || !selectedRole) {
      message.error('Please select a role');
      return;
    }
    setAssigning(true);
    const { error } = await client.POST('/rbac/user/{user_uuid}/roles', {
      params: { path: { user_uuid: uuid } },
      body: { role_name: selectedRole, domain: selectedDomain },
    });
    if (error) {
      message.error((error as { detail?: string }).detail || 'Failed to assign role');
      setAssigning(false);
      return;
    }
    message.success('Role assigned successfully');
    setAssignModalVisible(false);
    setAssigning(false);
    loadRoleAssignments();
  };

  const handleRemoveRole = async (roleName: string, domain: string) => {
    if (!uuid) return;
    const { error } = await client.DELETE('/rbac/user/{user_uuid}/roles/{role_name}', {
      params: {
        path: { user_uuid: uuid, role_name: roleName },
        query: { domain },
      },
    });
    if (error) {
      message.error('Failed to remove role');
      return;
    }
    message.success('Role removed');
    loadRoleAssignments();
  };

  // Role assignment table columns
  const roleAssignmentColumns: ColumnsType<RBACRoleAssignment> = [
    {
      title: 'Role',
      key: 'role',
      render: (_, record) => (
        <Space>
          {record.role.display_name}
          {record.role.is_system && <Tag color="blue">System</Tag>}
        </Space>
      ),
    },
    {
      title: 'Domain',
      dataIndex: 'domain',
      key: 'domain',
      render: (value: string) => (
        <Tag color={value === 'global' ? 'purple' : 'green'}>
          {value === 'global' ? 'Global' : value}
        </Tag>
      ),
    },
    {
      title: 'Assigned',
      dataIndex: 'assigned_at',
      key: 'assigned_at',
      render: (value: string) => (value ? new Date(value).toLocaleDateString() : '-'),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_, record) => (
        <Popconfirm
          title="Remove role"
          description="Are you sure you want to remove this role from the user?"
          onConfirm={() => handleRemoveRole(record.role.name, record.domain)}
          okButtonProps={{ danger: true }}
          okText="Remove"
        >
          <Button type="text" danger icon={<DeleteOutlined />} title="Remove" />
        </Popconfirm>
      ),
    },
  ];

  const onFinish = async (values: UserFormValues) => {
    setSaving(true);

    const userData: User = {
      email: values.email,
      first_name: values.first_name,
      last_name: values.last_name,
      password: values.password || '',
      is_enabled: values.is_enabled,
      is_admin: values.is_admin,
    };

    let result;
    if (isNew) {
      result = await client.POST('/user', {
        body: userData,
      });
    } else {
      result = await client.PUT('/user/{uuid}', {
        params: { path: { uuid: uuid! } },
        body: userData,
      });
    }

    if (result.error) {
      message.error(
        (result.error as { detail?: string }).detail ||
          `Failed to ${isNew ? 'create' : 'update'} user`
      );
      setSaving(false);
      return;
    }

    message.success(`User ${isNew ? 'created' : 'updated'} successfully`);
    navigate(`/w/${slug}/users`);
  };

  const onDelete = async () => {
    if (!uuid) return;
    setDeleting(true);
    const { error } = await client.DELETE('/user/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to delete user');
      setDeleting(false);
      return;
    }
    message.success('User deleted');
    navigate(`/w/${slug}/users`);
  };

  // Admin access check - after hooks
  if (!currentUser?.is_admin) {
    return (
      <Result
        status="403"
        title="Access Denied"
        subTitle="You need administrator privileges to access this page."
        extra={
          <Button type="primary" onClick={() => navigate(`/w/${slug}/`)}>
            Back to Dashboard
          </Button>
        }
      />
    );
  }

  if (loading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    );
  }

  return (
    <>
      <Title level={4}>{isNew ? 'Add' : 'Edit'} User</Title>
      <Form
        form={form}
        layout="vertical"
        onFinish={onFinish}
        initialValues={{
          is_enabled: true,
          is_admin: false,
        }}
        style={{ maxWidth: 500 }}
      >
        <Form.Item
          label="Email"
          name="email"
          rules={[
            { required: true, message: 'Please enter an email address' },
            { type: 'email', message: 'Please enter a valid email address' },
          ]}
        >
          <Input placeholder="user@example.com" />
        </Form.Item>

        <Form.Item
          label="First Name"
          name="first_name"
          rules={[{ required: true, message: 'Please enter a first name' }]}
        >
          <Input placeholder="John" />
        </Form.Item>

        <Form.Item
          label="Last Name"
          name="last_name"
          rules={[{ required: true, message: 'Please enter a last name' }]}
        >
          <Input placeholder="Doe" />
        </Form.Item>

        {isNew && (
          <Form.Item
            label="Password"
            name="password"
            rules={[
              { required: true, message: 'Please enter a password' },
              { min: 8, message: 'Password must be at least 8 characters' },
            ]}
          >
            <Input.Password placeholder="Enter password" />
          </Form.Item>
        )}

        <Form.Item name="is_enabled" valuePropName="checked">
          <Checkbox>User is enabled</Checkbox>
        </Form.Item>

        <Form.Item name="is_admin" valuePropName="checked">
          <Checkbox>User is administrator</Checkbox>
        </Form.Item>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={saving}>
              {isNew ? 'Create' : 'Update'}
            </Button>
            <Button onClick={() => navigate(`/w/${slug}/users`)}>Cancel</Button>
            {!isNew && uuid !== currentUser?.uuid && (
              <Popconfirm
                title="Delete User"
                description="Are you sure you want to delete this user? This action cannot be undone."
                onConfirm={onDelete}
                okButtonProps={{ danger: true }}
                okText="Delete"
              >
                <Button danger loading={deleting}>
                  Delete
                </Button>
              </Popconfirm>
            )}
          </Space>
        </Form.Item>
      </Form>

      {/* Role Assignments Section - only shown when editing */}
      {!isNew && uuid && (
        <>
          <Divider />
          <Space
            style={{
              marginBottom: 16,
              display: 'flex',
              justifyContent: 'space-between',
            }}
          >
            <Title level={5} style={{ margin: 0 }}>
              Role Assignments
            </Title>
            <Button type="primary" icon={<PlusOutlined />} onClick={openAssignModal}>
              Assign Role
            </Button>
          </Space>
          <Table
            columns={roleAssignmentColumns}
            dataSource={roleAssignments}
            rowKey={(record) => `${record.role.name}-${record.domain}`}
            loading={rolesLoading}
            pagination={false}
            size="small"
            locale={{ emptyText: 'No roles assigned' }}
          />

          {/* Assign Role Modal */}
          <Modal
            title="Assign Role"
            open={assignModalVisible}
            onCancel={() => setAssignModalVisible(false)}
            onOk={handleAssignRole}
            confirmLoading={assigning}
            okText="Assign"
          >
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <div>
                <Typography.Text strong>Role</Typography.Text>
                <Select
                  placeholder="Select a role"
                  style={{ width: '100%', marginTop: 8 }}
                  value={selectedRole}
                  onChange={setSelectedRole}
                  options={availableRoles.map((role) => ({
                    label: (
                      <Space>
                        {role.display_name}
                        <Tag color={role.scope === 'global' ? 'purple' : 'green'}>
                          {role.scope}
                        </Tag>
                      </Space>
                    ),
                    value: role.name,
                  }))}
                />
              </div>
              <div>
                <Typography.Text strong>Domain</Typography.Text>
                <Select
                  style={{ width: '100%', marginTop: 8 }}
                  value={selectedDomain}
                  onChange={setSelectedDomain}
                  options={[
                    { label: 'Global', value: 'global' },
                    { label: `Current Workspace (${slug})`, value: slug },
                  ]}
                />
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  Global roles apply system-wide. Workspace roles apply only to the selected workspace.
                </Typography.Text>
              </div>
            </Space>
          </Modal>
        </>
      )}
    </>
  );
}

export default UserEdit;
