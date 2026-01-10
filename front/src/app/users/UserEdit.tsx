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
import { useAuth, isAdmin } from '../../lib/auth';
import type { components } from '../../api/v1';

const { Title } = Typography;

type User = components['schemas']['user'];
type PolicySet = components['schemas']['policy_set'];
type UserPolicySetAssignment = components['schemas']['user_policy_set_assignment'];

interface UserFormValues {
  email: string;
  first_name: string;
  last_name: string;
  password?: string;
  is_enabled: boolean;
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

  // Policy set assignment state
  const [policySetAssignments, setPolicySetAssignments] = useState<UserPolicySetAssignment[]>([]);
  const [policySetsLoading, setPolicySetsLoading] = useState(false);
  const [assignModalVisible, setAssignModalVisible] = useState(false);
  const [availablePolicySets, setAvailablePolicySets] = useState<PolicySet[]>([]);
  const [assigning, setAssigning] = useState(false);
  const [selectedPolicySet, setSelectedPolicySet] = useState<string | undefined>();
  const [selectedDomain, setSelectedDomain] = useState<string>('global');

  // Load user data
  useEffect(() => {
    if (!isNew && uuid && isAdmin(currentUser)) {
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
        });
        setLoading(false);
      };
      loadUser();
    }
  }, [uuid, isNew, form, slug, navigate, currentUser]);

  // Load user's policy set assignments
  const loadPolicySetAssignments = useCallback(async () => {
    if (!uuid) return;
    setPolicySetsLoading(true);
    const { data, error } = await client.GET('/access/user/{user_uuid}/policy-sets', {
      params: { path: { user_uuid: uuid } },
    });
    if (error) {
      message.error('Failed to load policy set assignments');
      setPolicySetsLoading(false);
      return;
    }
    setPolicySetAssignments(data?.policy_sets || []);
    setPolicySetsLoading(false);
  }, [uuid]);

  useEffect(() => {
    if (!isNew && uuid && isAdmin(currentUser)) {
      loadPolicySetAssignments();
    }
  }, [uuid, isNew, loadPolicySetAssignments, currentUser]);

  // Load available policy sets for assignment modal
  const loadAvailablePolicySets = async () => {
    const { data, error } = await client.GET('/access/policy-set');
    if (error) {
      message.error('Failed to load policy sets');
      return;
    }
    setAvailablePolicySets(data || []);
  };

  const openAssignModal = () => {
    loadAvailablePolicySets();
    setSelectedPolicySet(undefined);
    setSelectedDomain(slug); // Default to current workspace
    setAssignModalVisible(true);
  };

  const handleAssignPolicySet = async () => {
    if (!uuid || !selectedPolicySet) {
      message.error('Please select a policy set');
      return;
    }
    setAssigning(true);
    const { error } = await client.POST('/access/user/{user_uuid}/policy-sets', {
      params: { path: { user_uuid: uuid } },
      body: { policy_set_name: selectedPolicySet, domain: selectedDomain },
    });
    if (error) {
      message.error((error as { detail?: string }).detail || 'Failed to assign policy set');
      setAssigning(false);
      return;
    }
    message.success('Policy set assigned successfully');
    setAssignModalVisible(false);
    setAssigning(false);
    loadPolicySetAssignments();
  };

  const handleRemovePolicySet = async (policySetName: string, domain: string) => {
    if (!uuid) return;
    const { error } = await client.DELETE('/access/user/{user_uuid}/policy-sets/{policy_set_name}', {
      params: {
        path: { user_uuid: uuid, policy_set_name: policySetName },
        query: { domain },
      },
    });
    if (error) {
      message.error('Failed to remove policy set');
      return;
    }
    message.success('Policy set removed');
    loadPolicySetAssignments();
  };

  // Policy set assignment table columns
  const policySetAssignmentColumns: ColumnsType<UserPolicySetAssignment> = [
    {
      title: 'Policy Set',
      dataIndex: 'policy_set',
      key: 'policy_set',
      render: (value: string) => <span>{value}</span>,
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
          title="Remove policy set"
          description="Are you sure you want to remove this policy set from the user?"
          onConfirm={() => handleRemovePolicySet(record.policy_set, record.domain)}
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
  if (!isAdmin(currentUser)) {
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

      {/* Policy Set Assignments Section - only shown when editing */}
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
              Policy Set Assignments
            </Title>
            <Button type="primary" icon={<PlusOutlined />} onClick={openAssignModal}>
              Assign Policy Set
            </Button>
          </Space>
          <Table
            columns={policySetAssignmentColumns}
            dataSource={policySetAssignments}
            rowKey={(record) => `${record.policy_set}-${record.domain}`}
            loading={policySetsLoading}
            pagination={false}
            size="small"
            locale={{ emptyText: 'No policy sets assigned' }}
          />

          {/* Assign Policy Set Modal */}
          <Modal
            title="Assign Policy Set"
            open={assignModalVisible}
            onCancel={() => setAssignModalVisible(false)}
            onOk={handleAssignPolicySet}
            confirmLoading={assigning}
            okText="Assign"
          >
            <Space direction="vertical" style={{ width: '100%' }} size="middle">
              <div>
                <Typography.Text strong>Policy Set</Typography.Text>
                <Select
                  placeholder="Select a policy set"
                  style={{ width: '100%', marginTop: 8 }}
                  value={selectedPolicySet}
                  onChange={setSelectedPolicySet}
                  options={availablePolicySets.map((ps) => ({
                    label: (
                      <Space>
                        {ps.display_name}
                        <Tag color={ps.scope === 'global' ? 'purple' : 'green'}>{ps.scope}</Tag>
                      </Space>
                    ),
                    value: ps.name,
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
                  Global policy sets apply system-wide. Workspace policy sets apply only to the
                  selected workspace.
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
