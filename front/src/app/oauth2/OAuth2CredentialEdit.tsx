import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router';
import { Typography, Form, Input, Select, Button, Space, message, Popconfirm, Row, Col, Card } from 'antd';
import client from '../../api/client';
import type { components } from '../../api/v1';
import { SmartLink } from '../../lib/SmartLink';
import { useWorkspace } from '../../lib/workspace/WorkspaceContext';

const { Title, Paragraph, Text } = Typography;

function CredentialDocumentation({ provider }: { provider?: string }) {
  if (provider === 'google') {
    return (
      <Card title="Google OAuth2 Setup" size="small">
        <Paragraph>
          To connect Google services as a data source, you need OAuth2 credentials from Google Cloud Console.
        </Paragraph>
        <Paragraph>
          <Text strong>Quick steps:</Text>
        </Paragraph>
        <ul>
          <li>Create a Google Cloud project</li>
          <li>Enable required APIs (e.g., Gmail API)</li>
          <li>Configure OAuth consent screen</li>
          <li>Create OAuth2 credentials (Web application)</li>
          <li>Copy Client ID and Client Secret here</li>
        </ul>
        <Paragraph>
          <SmartLink to="/documentation/datasource/gmail">
            View detailed Gmail setup guide →
          </SmartLink>
        </Paragraph>
      </Card>
    );
  }

  return (
    <Card title="About OAuth2 Credentials" size="small">
      <Paragraph>
        OAuth2 credentials are required to authenticate MeshPump with external services
        like Gmail, allowing secure access to your data without sharing passwords.
      </Paragraph>
      <Paragraph>
        Each provider requires its own OAuth2 application configuration. Select a provider
        on the left to see specific setup instructions.
      </Paragraph>
      <Paragraph>
        <Text strong>You will need:</Text>
      </Paragraph>
      <ul>
        <li><Text strong>Client ID</Text> — identifies your OAuth2 application</li>
        <li><Text strong>Client Secret</Text> — authenticates your application</li>
      </ul>
    </Card>
  );
}

type OAuth2Client = components['schemas']['oauth2_client'];

function OAuth2CredentialEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const { slug } = useWorkspace();
  const isNew = !uuid;
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);
  const provider = Form.useWatch('provider', form);

  useEffect(() => {
    if (!isNew && uuid) {
      const loadCredential = async () => {
        setLoading(true);
        const { data, error } = await client.GET('/oauth2/client/{uuid}', {
          params: { path: { uuid } },
        });
        if (error) {
          message.error('Failed to load OAuth2 credential');
          setLoading(false);
          return;
        }
        form.setFieldsValue(data);
        setLoading(false);
      };
      loadCredential();
    }
  }, [uuid, isNew, form]);

  const onFinish = async (values: OAuth2Client) => {
    setSaving(true);
    let result;
    if (isNew) {
      result = await client.POST('/oauth2/client', {
        body: {
          name: values.name,
          provider: values.provider,
          client_id: values.client_id,
          secret: values.secret,
        },
      });
    } else {
      result = await client.PUT('/oauth2/client/{uuid}', {
        params: { path: { uuid: uuid! } },
        body: {
          name: values.name,
          provider: values.provider,
          client_id: values.client_id,
          secret: values.secret,
        },
      });
    }

    if (result.error) {
      message.error(result.error.detail || 'Failed to save OAuth2 credential');
      setSaving(false);
      return;
    }

    message.success(isNew ? 'OAuth2 credential created' : 'OAuth2 credential updated');
    navigate(`/w/${slug}/oauth2/credentials`);
  };

  const onDelete = async () => {
    if (!uuid) return;
    setDeleting(true);
    const { error } = await client.DELETE('/oauth2/client/{uuid}', {
      params: { path: { uuid } },
    });
    if (error) {
      message.error(error.detail || 'Failed to delete OAuth2 credential');
      setDeleting(false);
      return;
    }
    message.success('OAuth2 credential deleted');
    navigate(`/w/${slug}/oauth2/credentials`);
  };

  return (
    <>
      <Title level={4}>{isNew ? 'Add' : 'Edit'} OAuth2 Credential</Title>
      <Row gutter={32}>
        <Col xs={24} lg={12}>
          <Form
            form={form}
            layout="vertical"
            onFinish={onFinish}
            disabled={loading}
          >
            <Form.Item
              label="Name"
              name="name"
              rules={[{ required: true, message: 'Please enter a name' }]}
            >
              <Input placeholder="My Google OAuth2" />
            </Form.Item>

            <Form.Item
              label="Provider"
              name="provider"
              rules={[{ required: true, message: 'Please select a provider' }]}
            >
              <Select placeholder="Select provider">
                <Select.Option value="google">Google</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              label="Client ID"
              name="client_id"
              rules={[{ required: true, message: 'Please enter the client ID' }]}
            >
              <Input placeholder="OAuth2 Client ID from provider" />
            </Form.Item>

            <Form.Item
              label="Client Secret"
              name="secret"
              rules={[{ required: true, message: 'Please enter the client secret' }]}
            >
              <Input.Password placeholder="OAuth2 Client Secret" />
            </Form.Item>

            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={saving}>
                  {isNew ? 'Create' : 'Update'}
                </Button>
                <Button onClick={() => navigate(`/w/${slug}/oauth2/credentials`)}>
                  Cancel
                </Button>
                {!isNew && (
                  <Popconfirm
                    title="Delete OAuth2 Credential"
                    description="Are you sure you want to delete this credential?"
                    onConfirm={onDelete}
                    okText="Yes"
                    cancelText="No"
                    okButtonProps={{ danger: true }}
                  >
                    <Button danger loading={deleting}>
                      Delete
                    </Button>
                  </Popconfirm>
                )}
              </Space>
            </Form.Item>
          </Form>
        </Col>
        <Col xs={24} lg={12}>
          <CredentialDocumentation provider={provider} />
        </Col>
      </Row>
    </>
  );
}

export default OAuth2CredentialEdit;
