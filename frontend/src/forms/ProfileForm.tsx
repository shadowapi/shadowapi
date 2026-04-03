import { ReactElement, useEffect } from 'react'
import { Button, Form, Input, Typography, message } from 'antd'
import { useSWRConfig } from 'swr'

import apiClient from '@/api/client'
import { useApiGet } from '@/api/hooks'

interface ProfileData {
  first_name: string
  last_name: string
}

export function ProfileForm(): ReactElement {
  const { mutate: globalMutate } = useSWRConfig()
  const [form] = Form.useForm<ProfileData>()

  const { data, isLoading } = useApiGet<ProfileData>('/profile')

  useEffect(() => {
    if (data) {
      form.setFieldsValue({ first_name: data.first_name, last_name: data.last_name })
    }
  }, [data, form])

  const onSubmit = async (values: ProfileData) => {
    try {
      await apiClient.put('/profile', values)
      message.success('Profile updated')
      globalMutate('/profile')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  if (isLoading) return <></>

  return (
    <div style={{ display: 'flex', justifyContent: 'center', minHeight: '100vh' }}>
      <Form
        form={form}
        onFinish={onSubmit}
        layout="horizontal" labelCol={{ span: 6 }} wrapperCol={{ span: 14 }}
        style={{ width: 400 }}
        initialValues={{ first_name: '', last_name: '' }}
      >
        <Typography.Title level={4}>Edit Profile</Typography.Title>

        <Form.Item name="first_name" label="First Name" rules={[{ required: true, message: 'First name is required' }]}>
          <Input />
        </Form.Item>

        <Form.Item name="last_name" label="Last Name" rules={[{ required: true, message: 'Last name is required' }]}>
          <Input />
        </Form.Item>

        <Form.Item>
          <Button type="primary" htmlType="submit">
            Update
          </Button>
        </Form.Item>
      </Form>
    </div>
  )
}
