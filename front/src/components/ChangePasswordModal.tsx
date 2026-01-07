import { useState } from 'react';
import { Modal, Form, Input, message, Alert } from 'antd';
import { LockOutlined } from '@ant-design/icons';
import client from '../api/client';

interface ChangePasswordModalProps {
  open: boolean;
  onClose: () => void;
}

interface PasswordFormValues {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}

function ChangePasswordModal({ open, onClose }: ChangePasswordModalProps) {
  const [form] = Form.useForm<PasswordFormValues>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (values: PasswordFormValues) => {
    setLoading(true);
    setError(null);

    const { error: apiError } = await client.PUT('/profile/password', {
      body: {
        current_password: values.currentPassword,
        new_password: values.newPassword,
      },
    });

    setLoading(false);

    if (apiError) {
      const errorDetail = (apiError as { detail?: string }).detail;
      if (errorDetail?.toLowerCase().includes('current password')) {
        setError('Current password is incorrect');
      } else {
        setError(errorDetail || 'Failed to change password');
      }
      return;
    }

    message.success('Password changed successfully');
    form.resetFields();
    onClose();
  };

  const handleCancel = () => {
    form.resetFields();
    setError(null);
    onClose();
  };

  return (
    <Modal
      title="Change Password"
      open={open}
      onCancel={handleCancel}
      onOk={() => form.submit()}
      confirmLoading={loading}
      okText="Change Password"
      destroyOnClose
    >
      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
      >
        <Form.Item
          name="currentPassword"
          label="Current Password"
          rules={[
            { required: true, message: 'Please enter your current password' },
          ]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="Enter current password"
          />
        </Form.Item>

        <Form.Item
          name="newPassword"
          label="New Password"
          rules={[
            { required: true, message: 'Please enter a new password' },
            { min: 8, message: 'Password must be at least 8 characters' },
          ]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="Enter new password"
          />
        </Form.Item>

        <Form.Item
          name="confirmPassword"
          label="Confirm New Password"
          dependencies={['newPassword']}
          rules={[
            { required: true, message: 'Please confirm your new password' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('newPassword') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error('Passwords do not match'));
              },
            }),
          ]}
        >
          <Input.Password
            prefix={<LockOutlined />}
            placeholder="Confirm new password"
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}

export default ChangePasswordModal;
