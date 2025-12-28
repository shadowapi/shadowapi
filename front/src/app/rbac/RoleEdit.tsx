import { useEffect, useState, useMemo, useCallback } from 'react';
import { useNavigate, useParams } from 'react-router';
import {
  Typography,
  Form,
  Input,
  Select,
  Button,
  Space,
  message,
  Popconfirm,
  Row,
  Col,
  Card,
  Checkbox,
  Table,
  Alert,
  Result,
  Spin,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import client from '../../api/client';
import type { components } from '../../api/v1';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth } from '../../lib/auth';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;

type RBACRole = components['schemas']['rbac_role'];
type RBACPermission = components['schemas']['rbac_permission'];

interface RoleFormValues {
  name: string;
  display_name: string;
  description?: string;
  scope: 'global' | 'workspace';
}

interface PermissionsByResource {
  resource: string;
  permissions: RBACPermission[];
}

function RoleDocumentation() {
  return (
    <Card title="About Roles" size="small">
      <Paragraph>
        Roles define sets of permissions that can be assigned to users. Each role has a scope
        that determines where it can be applied.
      </Paragraph>
      <Paragraph>
        <Text strong>Scope types:</Text>
      </Paragraph>
      <ul>
        <li>
          <Text strong>Global</Text> — Applies system-wide, for operations like managing users
          and workspaces
        </li>
        <li>
          <Text strong>Workspace</Text> — Applies within a specific workspace, for operations
          like managing data sources and pipelines
        </li>
      </ul>
      <Paragraph>
        <Text strong>Permissions:</Text>
      </Paragraph>
      <Paragraph>
        Select the actions users with this role can perform on each resource type. Common
        actions include read, write, create, delete, and admin.
      </Paragraph>
    </Card>
  );
}

function RoleEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const isNew = !uuid;
  const [form] = Form.useForm<RoleFormValues>();
  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [isSystemRole, setIsSystemRole] = useState(false);
  const [permissions, setPermissions] = useState<RBACPermission[]>([]);
  const [selectedPermissions, setSelectedPermissions] = useState<Set<string>>(new Set());
  const [permissionsLoading, setPermissionsLoading] = useState(true);

  const formScope = Form.useWatch('scope', form);

  // Load role for editing
  useEffect(() => {
    if (!isNew && uuid && currentUser?.is_admin) {
      const loadRole = async () => {
        setLoading(true);
        const { data, error } = await client.GET('/rbac/role/{uuid}', {
          params: { path: { uuid } },
        });
        if (error) {
          message.error('Failed to load role');
          navigate(`/w/${slug}/rbac/roles`);
          return;
        }
        form.setFieldsValue({
          name: data.name,
          display_name: data.display_name,
          description: data.description,
          scope: data.scope,
        });
        setIsSystemRole(data.is_system || false);
        // Set selected permissions
        const permNames = new Set(data.permissions?.map((p) => p.name) || []);
        setSelectedPermissions(permNames);
        setLoading(false);
      };
      loadRole();
    }
  }, [uuid, isNew, form, slug, navigate, currentUser?.is_admin]);

  // Load permissions based on scope
  const loadPermissions = useCallback(async () => {
    setPermissionsLoading(true);
    const { data, error } = await client.GET('/rbac/permission', {
      params: { query: formScope ? { scope: formScope } : {} },
    });
    if (error) {
      message.error('Failed to load permissions');
      setPermissionsLoading(false);
      return;
    }
    setPermissions(data || []);
    setPermissionsLoading(false);
  }, [formScope]);

  useEffect(() => {
    if (currentUser?.is_admin) {
      loadPermissions();
    }
  }, [loadPermissions, currentUser?.is_admin]);

  // Group permissions by resource
  const permissionsByResource = useMemo(() => {
    const groups: Record<string, RBACPermission[]> = {};
    permissions.forEach((perm) => {
      if (!groups[perm.resource]) {
        groups[perm.resource] = [];
      }
      groups[perm.resource].push(perm);
    });
    return Object.entries(groups)
      .map(([resource, perms]) => ({
        resource,
        permissions: perms.sort((a, b) => {
          const order = ['read', 'write', 'create', 'delete', 'admin', '*'];
          return order.indexOf(a.action) - order.indexOf(b.action);
        }),
      }))
      .sort((a, b) => a.resource.localeCompare(b.resource));
  }, [permissions]);

  const togglePermission = (permName: string) => {
    setSelectedPermissions((prev) => {
      const next = new Set(prev);
      if (next.has(permName)) {
        next.delete(permName);
      } else {
        next.add(permName);
      }
      return next;
    });
  };

  const onFinish = async (values: RoleFormValues) => {
    setSaving(true);

    // Map selected permission names to full permission objects
    const selectedPerms = Array.from(selectedPermissions)
      .map((name) => permissions.find((p) => p.name === name))
      .filter((p): p is RBACPermission => p !== undefined);

    const roleData: RBACRole = {
      name: values.name,
      display_name: values.display_name,
      description: values.description,
      scope: values.scope,
      permissions: selectedPerms,
    };

    let result;
    if (isNew) {
      result = await client.POST('/rbac/role', {
        body: roleData,
      });
    } else {
      result = await client.PUT('/rbac/role/{uuid}', {
        params: { path: { uuid: uuid! } },
        body: roleData,
      });
    }

    if (result.error) {
      message.error(
        (result.error as { detail?: string }).detail ||
          `Failed to ${isNew ? 'create' : 'update'} role`
      );
      setSaving(false);
      return;
    }

    message.success(`Role ${isNew ? 'created' : 'updated'} successfully`);
    navigate(`/w/${slug}/rbac/roles`);
  };

  const onDelete = async () => {
    if (!uuid) return;
    setDeleting(true);
    const { error } = await client.DELETE('/rbac/role/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to delete role');
      setDeleting(false);
      return;
    }
    message.success('Role deleted');
    navigate(`/w/${slug}/rbac/roles`);
  };

  const permissionColumns: ColumnsType<PermissionsByResource> = [
    {
      title: 'Resource',
      dataIndex: 'resource',
      key: 'resource',
      width: 150,
      render: (value: string) => <Text strong style={{ textTransform: 'capitalize' }}>{value}</Text>,
    },
    {
      title: 'Read',
      key: 'read',
      width: 80,
      render: (_, record) => {
        const perm = record.permissions.find((p) => p.action === 'read');
        if (!perm) return '-';
        return (
          <Checkbox
            checked={selectedPermissions.has(perm.name)}
            onChange={() => togglePermission(perm.name)}
            disabled={isSystemRole}
          />
        );
      },
    },
    {
      title: 'Write',
      key: 'write',
      width: 80,
      render: (_, record) => {
        const perm = record.permissions.find((p) => p.action === 'write');
        if (!perm) return '-';
        return (
          <Checkbox
            checked={selectedPermissions.has(perm.name)}
            onChange={() => togglePermission(perm.name)}
            disabled={isSystemRole}
          />
        );
      },
    },
    {
      title: 'Create',
      key: 'create',
      width: 80,
      render: (_, record) => {
        const perm = record.permissions.find((p) => p.action === 'create');
        if (!perm) return '-';
        return (
          <Checkbox
            checked={selectedPermissions.has(perm.name)}
            onChange={() => togglePermission(perm.name)}
            disabled={isSystemRole}
          />
        );
      },
    },
    {
      title: 'Delete',
      key: 'delete',
      width: 80,
      render: (_, record) => {
        const perm = record.permissions.find((p) => p.action === 'delete');
        if (!perm) return '-';
        return (
          <Checkbox
            checked={selectedPermissions.has(perm.name)}
            onChange={() => togglePermission(perm.name)}
            disabled={isSystemRole}
          />
        );
      },
    },
    {
      title: 'Admin',
      key: 'admin',
      width: 80,
      render: (_, record) => {
        const perm = record.permissions.find((p) => p.action === 'admin');
        if (!perm) return '-';
        return (
          <Checkbox
            checked={selectedPermissions.has(perm.name)}
            onChange={() => togglePermission(perm.name)}
            disabled={isSystemRole}
          />
        );
      },
    },
  ];

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
      <Title level={4}>
        {isNew ? 'Create' : isSystemRole ? 'View' : 'Edit'} Role
      </Title>

      {isSystemRole && (
        <Alert
          message="System Role"
          description="This is a system-defined role and cannot be modified. You can view its configuration but changes are not allowed."
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <Row gutter={32}>
        <Col xs={24} lg={14}>
          <Form
            form={form}
            layout="vertical"
            onFinish={onFinish}
            initialValues={{ scope: 'workspace' }}
            disabled={isSystemRole}
          >
            <Form.Item
              label="Name"
              name="name"
              rules={[
                { required: true, message: 'Please enter a role name' },
                {
                  pattern: /^[a-z][a-z0-9_]*$/,
                  message: 'Name must start with lowercase letter and contain only lowercase letters, numbers, and underscores',
                },
              ]}
              extra="Unique identifier for the role (e.g., custom_editor)"
            >
              <Input placeholder="custom_role" disabled={!isNew || isSystemRole} />
            </Form.Item>

            <Form.Item
              label="Display Name"
              name="display_name"
              rules={[{ required: true, message: 'Please enter a display name' }]}
            >
              <Input placeholder="Custom Editor" />
            </Form.Item>

            <Form.Item label="Description" name="description">
              <TextArea rows={3} placeholder="Describe what this role is for..." />
            </Form.Item>

            <Form.Item
              label="Scope"
              name="scope"
              rules={[{ required: true, message: 'Please select a scope' }]}
            >
              <Select disabled={!isNew || isSystemRole}>
                <Select.Option value="global">Global</Select.Option>
                <Select.Option value="workspace">Workspace</Select.Option>
              </Select>
            </Form.Item>

            <Title level={5} style={{ marginTop: 24, marginBottom: 16 }}>
              Permissions
            </Title>

            <Table
              columns={permissionColumns}
              dataSource={permissionsByResource}
              rowKey="resource"
              loading={permissionsLoading}
              pagination={false}
              size="small"
              style={{ marginBottom: 24 }}
            />

            {!isSystemRole && (
              <Form.Item>
                <Space>
                  <Button type="primary" htmlType="submit" loading={saving}>
                    {isNew ? 'Create' : 'Update'}
                  </Button>
                  <Button onClick={() => navigate(`/w/${slug}/rbac/roles`)}>Cancel</Button>
                  {!isNew && (
                    <Popconfirm
                      title="Delete Role"
                      description="Are you sure you want to delete this role? Users with this role will lose its permissions."
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
            )}

            {isSystemRole && (
              <Form.Item>
                <Button onClick={() => navigate(`/w/${slug}/rbac/roles`)}>Back to Roles</Button>
              </Form.Item>
            )}
          </Form>
        </Col>
        <Col xs={24} lg={10}>
          <RoleDocumentation />
        </Col>
      </Row>
    </>
  );
}

export default RoleEdit;
