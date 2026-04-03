import { ReactElement, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Space, Switch, Typography, message } from 'antd'
import { useSWRConfig } from 'swr'

import apiClient from '@/api/client'
import { useApiGet } from '@/api/hooks'

type UserFormData = {
  email: string
  password: string
  first_name: string
  last_name: string
  is_enabled: boolean
  is_admin: boolean
}

export function UserForm({ userUUID }: { userUUID: string }): ReactElement {
  const navigate = useNavigate()
  const { mutate: globalMutate } = useSWRConfig()
  const [form] = Form.useForm<UserFormData>()

  const isAdd = userUUID === 'add'

  const { data, isLoading } = useApiGet<UserFormData>(isAdd ? null : `/user/${userUUID}`)

  useEffect(() => {
    if (data && !isAdd) {
      form.setFieldsValue(data)
    }
  }, [data, isAdd, form])

  const onSubmit = async (values: UserFormData) => {
    try {
      if (isAdd) {
        await apiClient.post('/user', values)
      } else {
        await apiClient.put(`/user/${userUUID}`, values)
      }
      message.success(isAdd ? 'User created' : 'User updated')
      globalMutate('/user')
      navigate('/users')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  const onDelete = async () => {
    try {
      await apiClient.delete(`/user/${userUUID}`)
      message.success('User deleted')
      globalMutate('/user')
      navigate('/users')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  if (isLoading && !isAdd) return <></>

  return (
    <div style={{ display: 'flex', justifyContent: 'center', minHeight: '100vh' }}>
      <Form
        form={form}
        onFinish={onSubmit}
        layout="horizontal" labelCol={{ span: 6 }} wrapperCol={{ span: 14 }}
        style={{ width: 400 }}
        initialValues={{ email: '', password: '', first_name: '', last_name: '', is_enabled: false, is_admin: false }}
      >
        <Typography.Title level={4}>{isAdd ? 'Add User' : 'Edit User'}</Typography.Title>

        <Form.Item name="email" label="Email" rules={[{ required: true, message: 'Email is required' }]}>
          <Input type="email" />
        </Form.Item>

        <Form.Item
          name="password"
          label="Password"
          rules={[{ required: isAdd, message: 'Password is required' }]}
        >
          <Input.Password />
        </Form.Item>

        <Form.Item name="first_name" label="First Name" rules={[{ required: true, message: 'First name is required' }]}>
          <Input />
        </Form.Item>

        <Form.Item name="last_name" label="Last Name" rules={[{ required: true, message: 'Last name is required' }]}>
          <Input />
        </Form.Item>

        <Form.Item name="is_enabled" label="Enabled" valuePropName="checked">
          <Switch />
        </Form.Item>

        <Form.Item name="is_admin" label="Admin" valuePropName="checked">
          <Switch />
        </Form.Item>

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">
              {isAdd ? 'Create' : 'Update'}
            </Button>
            {!isAdd && (
              <Button danger onClick={onDelete}>
                Delete
              </Button>
            )}
          </Space>
        </Form.Item>
      </Form>
    </div>
  )
}
