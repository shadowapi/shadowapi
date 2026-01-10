import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router';
import {
  Typography,
  Input,
  Button,
  Space,
  message,
  Popconfirm,
  Row,
  Col,
  Card,
  Alert,
  Result,
  Spin,
  Dropdown,
} from 'antd';
import { PlusOutlined, DownOutlined } from '@ant-design/icons';
import type { MenuProps } from 'antd';
import yaml from 'js-yaml';
import client from '../../api/client';
import type { components } from '../../api/v1';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';
import { useAuth, isAdmin } from '../../lib/auth';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;

type PolicySet = components['schemas']['policy_set'];

// Predefined policy set templates in YAML format
const predefinedTemplates: Record<string, string> = {
  super_admin: `name: super_admin
display_name: Super Admin
description: Full system access including user management and all workspaces
scope: global
permissions:
  - resource: "*"
    action: "*"`,
  workspace_owner: `name: workspace_owner
display_name: Workspace Owner
description: Full control over the workspace including member management
scope: workspace
permissions:
  - resource: workspace
    action: admin
  - resource: datasource
    action: "*"
  - resource: pipeline
    action: "*"
  - resource: storage
    action: "*"
  - resource: contact
    action: "*"
  - resource: message
    action: "*"
  - resource: scheduler
    action: "*"
  - resource: member
    action: "*"`,
  workspace_admin: `name: workspace_admin
display_name: Workspace Admin
description: Can manage workspace resources but not members
scope: workspace
permissions:
  - resource: workspace
    action: read
  - resource: datasource
    action: "*"
  - resource: pipeline
    action: "*"
  - resource: storage
    action: "*"
  - resource: contact
    action: "*"
  - resource: message
    action: "*"
  - resource: scheduler
    action: "*"
  - resource: member
    action: read`,
  workspace_member: `name: workspace_member
display_name: Workspace Member
description: Read-only access to workspace resources
scope: workspace
permissions:
  - resource: workspace
    action: read
  - resource: datasource
    action: read
  - resource: pipeline
    action: read
  - resource: storage
    action: read
  - resource: contact
    action: read
  - resource: message
    action: read
  - resource: scheduler
    action: read
  - resource: member
    action: read`,
};

interface PolicyYaml {
  name: string;
  display_name: string;
  description?: string;
  scope: 'global' | 'workspace';
  permissions: Array<{ resource: string; action: string }>;
}

function PolicySetDocumentation() {
  return (
    <Card title="About Policy Sets" size="small">
      <Paragraph>
        Policy sets define collections of permissions that can be assigned to users. Write your
        policy in YAML format with the structure shown in the editor.
      </Paragraph>
      <Paragraph>
        <Text strong>Scope types:</Text>
      </Paragraph>
      <ul>
        <li>
          <Text strong>global</Text> — System-wide operations (managing users, workspaces)
        </li>
        <li>
          <Text strong>workspace</Text> — Operations within a specific workspace
        </li>
      </ul>
      <Paragraph>
        <Text strong>Wildcard:</Text> Use <Text code>*</Text> for action to grant all actions on a
        resource.
      </Paragraph>
    </Card>
  );
}

function PolicySetEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const { user: currentUser } = useAuth();
  const isNew = !uuid;
  const [loading, setLoading] = useState(!isNew);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const [isSystemPolicySet, setIsSystemPolicySet] = useState(false);
  const [policyYaml, setPolicyYaml] = useState<string>('');
  const [yamlError, setYamlError] = useState<string | null>(null);

  // Load policy set for editing
  useEffect(() => {
    if (!isNew && uuid && isAdmin(currentUser)) {
      const loadPolicySet = async () => {
        setLoading(true);
        const { data, error } = await client.GET('/access/policy-set/{uuid}', {
          params: { path: { uuid } },
        });
        if (error) {
          message.error('Failed to load policy set');
          navigate(`/w/${slug}/access/policy-sets`);
          return;
        }
        setIsSystemPolicySet(data.is_system || false);

        // Convert to YAML for display
        const yamlContent = yaml.dump(
          {
            name: data.name,
            display_name: data.display_name,
            description: data.description || undefined,
            scope: data.scope,
            permissions:
              data.permissions?.map((p) => ({
                resource: p.resource,
                action: p.action,
              })) || [],
          },
          { lineWidth: -1, quotingType: '"', forceQuotes: false }
        );

        setPolicyYaml(yamlContent);
        setLoading(false);
      };
      loadPolicySet();
    }
  }, [uuid, isNew, slug, navigate, currentUser]);

  // Set default template for new policy sets
  useEffect(() => {
    if (isNew && !policyYaml) {
      setPolicyYaml(`name: my_policy
display_name: My Policy
description: Description of what this policy allows
scope: workspace
permissions:
  - resource: datasource
    action: read
  - resource: pipeline
    action: read`);
    }
  }, [isNew, policyYaml]);

  const validateYaml = (yamlStr: string): PolicyYaml | null => {
    try {
      const parsed = yaml.load(yamlStr) as PolicyYaml;

      // Validate required fields
      if (!parsed.name || typeof parsed.name !== 'string') {
        setYamlError('Missing or invalid "name" field');
        return null;
      }
      if (!parsed.display_name || typeof parsed.display_name !== 'string') {
        setYamlError('Missing or invalid "display_name" field');
        return null;
      }
      if (!parsed.scope || !['global', 'workspace'].includes(parsed.scope)) {
        setYamlError('Missing or invalid "scope" field (must be "global" or "workspace")');
        return null;
      }
      if (!Array.isArray(parsed.permissions)) {
        setYamlError('Missing or invalid "permissions" array');
        return null;
      }

      // Validate name format
      if (!/^[a-z][a-z0-9_]*$/.test(parsed.name)) {
        setYamlError(
          'Name must start with lowercase letter and contain only lowercase letters, numbers, and underscores'
        );
        return null;
      }

      // Validate each permission
      for (const perm of parsed.permissions) {
        if (!perm.resource || !perm.action) {
          setYamlError('Each permission must have "resource" and "action" fields');
          return null;
        }
      }

      setYamlError(null);
      return parsed;
    } catch (e) {
      setYamlError(`Invalid YAML: ${e instanceof Error ? e.message : 'Unknown error'}`);
      return null;
    }
  };

  const handleYamlChange = (value: string) => {
    setPolicyYaml(value);
    // Clear error while typing, will validate on submit
    if (yamlError) {
      setYamlError(null);
    }
  };

  const onFinish = async () => {
    const parsed = validateYaml(policyYaml);
    if (!parsed) {
      return;
    }

    setSaving(true);

    const policySetData: PolicySet = {
      name: parsed.name,
      display_name: parsed.display_name,
      description: parsed.description,
      scope: parsed.scope,
      permissions: parsed.permissions.map((p) => ({
        name: `${p.resource}:${p.action}`,
        resource: p.resource,
        action: p.action as 'read' | 'write' | 'create' | 'delete' | 'admin' | '*',
      })),
    };

    let result;
    if (isNew) {
      result = await client.POST('/access/policy-set', {
        body: policySetData,
      });
    } else {
      result = await client.PUT('/access/policy-set/{uuid}', {
        params: { path: { uuid: uuid! } },
        body: policySetData,
      });
    }

    if (result.error) {
      message.error(
        (result.error as { detail?: string }).detail ||
          `Failed to ${isNew ? 'create' : 'update'} policy set`
      );
      setSaving(false);
      return;
    }

    message.success(`Policy set ${isNew ? 'created' : 'updated'} successfully`);
    navigate(`/w/${slug}/access/policy-sets`);
  };

  const onDelete = async () => {
    if (!uuid) return;
    setDeleting(true);
    const { error } = await client.DELETE('/access/policy-set/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error('Failed to delete policy set');
      setDeleting(false);
      return;
    }
    message.success('Policy set deleted');
    navigate(`/w/${slug}/access/policy-sets`);
  };

  const handlePredefinedClick: MenuProps['onClick'] = ({ key }) => {
    const template = predefinedTemplates[key];
    if (template) {
      setPolicyYaml(template);
      setYamlError(null);
    }
  };

  const predefinedMenuItems: MenuProps['items'] = [
    { key: 'super_admin', label: 'Super Admin (global)' },
    { key: 'workspace_owner', label: 'Workspace Owner' },
    { key: 'workspace_admin', label: 'Workspace Admin' },
    { key: 'workspace_member', label: 'Workspace Member' },
  ];

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
      <Title level={4}>{isNew ? 'Create' : isSystemPolicySet ? 'View' : 'Edit'} Policy Set</Title>

      {isSystemPolicySet && (
        <Alert
          message="System Policy Set"
          description="This is a system-defined policy set and cannot be modified. You can view its configuration but changes are not allowed."
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <Row gutter={32}>
        <Col xs={24} lg={14}>
          <Title level={5} style={{ marginBottom: 16 }}>
            Policy Definition (YAML)
          </Title>

          {!isSystemPolicySet && (
            <Space style={{ marginBottom: 12 }}>
              <Dropdown menu={{ items: predefinedMenuItems, onClick: handlePredefinedClick }}>
                <Button icon={<PlusOutlined />}>
                  Add Predefined Template <DownOutlined />
                </Button>
              </Dropdown>
            </Space>
          )}

          <TextArea
            rows={18}
            style={{
              fontFamily: 'monospace',
              fontSize: 13,
              marginBottom: 16,
            }}
            value={policyYaml}
            onChange={(e) => handleYamlChange(e.target.value)}
            disabled={isSystemPolicySet}
            status={yamlError ? 'error' : undefined}
          />

          {yamlError && (
            <Alert message={yamlError} type="error" showIcon style={{ marginBottom: 16 }} />
          )}

          <Alert
            type="info"
            message="Available Resources & Actions"
            description={
              <>
                <p>
                  <Text strong>Global resources:</Text> user, workspace, policy_set, rbac, worker
                </p>
                <p>
                  <Text strong>Workspace resources:</Text> workspace, datasource, pipeline, storage,
                  contact, message, scheduler, member
                </p>
                <p>
                  <Text strong>Actions:</Text> read, write, create, delete, admin, * (all)
                </p>
              </>
            }
            style={{ marginBottom: 24 }}
          />

          {!isSystemPolicySet && (
            <Space>
              <Button type="primary" onClick={onFinish} loading={saving}>
                {isNew ? 'Create' : 'Update'}
              </Button>
              <Button onClick={() => navigate(`/w/${slug}/access/policy-sets`)}>Cancel</Button>
              {!isNew && (
                <Popconfirm
                  title="Delete Policy Set"
                  description="Are you sure you want to delete this policy set? Users with this policy set will lose its permissions."
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
          )}

          {isSystemPolicySet && (
            <Button onClick={() => navigate(`/w/${slug}/access/policy-sets`)}>
              Back to Policy Sets
            </Button>
          )}
        </Col>
        <Col xs={24} lg={10}>
          <PolicySetDocumentation />
        </Col>
      </Row>
    </>
  );
}

export default PolicySetEdit;
