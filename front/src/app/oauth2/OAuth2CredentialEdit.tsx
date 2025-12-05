import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router';
import { Typography, Form, Input, Select, Button, Space, message, Popconfirm } from 'antd';
import client from '../../api/client';
import type { components } from '../../api/v1';

const { Title } = Typography;

type OAuth2Client = components['schemas']['oauth2_client'];

function OAuth2CredentialEdit() {
  const navigate = useNavigate();
  const { uuid } = useParams<{ uuid: string }>();
  const isNew = !uuid;
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deleting, setDeleting] = useState(false);

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
    navigate('/oauth2/credentials');
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
    navigate('/oauth2/credentials');
  };

  return (
    <>
      <Title level={4}>{isNew ? 'Add' : 'Edit'} OAuth2 Credential</Title>
      <Form
        form={form}
        layout="vertical"
        style={{ maxWidth: 500 }}
        onFinish={onFinish}
        disabled={loading}
      >
        <Form.Item
          label="Name"
          name="name"
          rules={[{ required: true, message: 'Please enter a name' }]}
        >
          <Input placeholder="My Gmail OAuth2" />
        </Form.Item>

        <Form.Item
          label="Provider"
          name="provider"
          rules={[{ required: true, message: 'Please select a provider' }]}
        >
          <Select placeholder="Select provider">
            <Select.Option value="GMAIL">Gmail</Select.Option>
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
            <Button onClick={() => navigate('/oauth2/credentials')}>
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
    </>
  );
}

export default OAuth2CredentialEdit;
