import { ReactElement, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Select, Space, Switch, Typography, message } from 'antd'
import { useSWRConfig } from 'swr'

import apiClient from '@/api/client'
import { useApiGet } from '@/api/hooks'
import type { components, paths } from '@/api/v1'

export type StorageKind = 's3' | 'hostfiles' | 'postgres'

type StorageBase = {
  name: string
  type: string
  is_enabled: boolean
}

type StorageFormData =
  | (StorageBase & { type: 's3' } & components['schemas']['storage_s3'])
  | (StorageBase & { type: 'hostfiles' } & components['schemas']['storage_hostfiles'])
  | (StorageBase & { type: 'postgres' } & components['schemas']['storage_postgres'])

const createEndpoints: Record<StorageKind, string> = {
  s3: '/storage/s3',
  hostfiles: '/storage/hostfiles',
  postgres: '/storage/postgres',
}

const updateEndpoints: Record<StorageKind, string> = {
  s3: '/storage/s3',
  hostfiles: '/storage/hostfiles',
  postgres: '/storage/postgres',
}

const deleteEndpoints: Record<StorageKind, string> = {
  s3: '/storage/s3',
  hostfiles: '/storage/hostfiles',
  postgres: '/storage/postgres',
}

export function StorageForm({
  storageUUID,
  storageKind,
}: {
  storageUUID: string
  storageKind?: StorageKind
}): ReactElement {
  const navigate = useNavigate()
  const { mutate: globalMutate } = useSWRConfig()
  const [form] = Form.useForm<StorageFormData>()

  if (storageUUID !== 'add' && !storageKind) {
    throw new Error('storageKind is required for editing storage')
  }

  const storageType = Form.useWatch('type', form)

  const { data: detailData, isLoading } = useApiGet<StorageFormData>(
    storageUUID !== 'add' && storageKind ? `${updateEndpoints[storageKind]}/${storageUUID}` : null
  )

  useEffect(() => {
    if (detailData) {
      form.setFieldsValue({
        ...detailData,
        type: storageKind ?? (detailData as Partial<StorageFormData>).type,
      } as StorageFormData)
    } else if (storageKind) {
      form.setFieldValue('type', storageKind)
    }
  }, [detailData, storageKind, form])

  const onSubmit = async (data: StorageFormData) => {
    try {
      const { type, ...payload } = data
      if (storageUUID === 'add') {
        const currentKind = type as StorageKind
        await apiClient.post(createEndpoints[currentKind], payload)
      } else {
        await apiClient.put(`${updateEndpoints[storageKind!]}/${storageUUID}`, payload)
      }
      message.success(storageUUID === 'add' ? 'Storage created' : 'Storage updated')
      globalMutate('/storage')
      navigate('/storages')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  const onDelete = async () => {
    try {
      await apiClient.delete(`${deleteEndpoints[storageKind!]}/${storageUUID}`)
      message.success('Storage deleted')
      globalMutate('/storage')
      navigate('/storages')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  if (isLoading && storageUUID !== 'add') {
    return <></>
  }

  return (
    <div style={{ display: 'flex', justifyContent: 'center', width: '100%', minHeight: '100vh' }}>
      <Form form={form} onFinish={onSubmit} layout="horizontal" labelCol={{ span: 6 }} wrapperCol={{ span: 14 }} style={{ width: 400 }}>
        <Typography.Title level={4}>Storage</Typography.Title>

        <Form.Item name="name" label="Name" rules={[{ required: true, message: 'Name is required' }]}>
          <Input />
        </Form.Item>

        <Form.Item name="type" label="Type" rules={[{ required: true, message: 'Type is required' }]}>
          <Select disabled={!!storageKind || storageUUID !== 'add'}>
            <Select.Option value="s3">S3</Select.Option>
            <Select.Option value="hostfiles">File System</Select.Option>
            <Select.Option value="postgres">PostgreSQL</Select.Option>
          </Select>
        </Form.Item>

        <Form.Item name="is_enabled" label="Enabled" valuePropName="checked">
          <Switch />
        </Form.Item>

        {storageType === 's3' && (
          <>
            <Form.Item name="provider" label="S3 Provider">
              <Input />
            </Form.Item>
            <Form.Item name="region" label="Region">
              <Input />
            </Form.Item>
            <Form.Item name="bucket" label="Bucket Name">
              <Input />
            </Form.Item>
            <Form.Item name="access_key_id" label="Access Key ID">
              <Input />
            </Form.Item>
            <Form.Item name="secret_access_key" label="Secret Access Key">
              <Input.Password />
            </Form.Item>
          </>
        )}

        {storageType === 'hostfiles' && (
          <Form.Item name="path" label="File System Path">
            <Input />
          </Form.Item>
        )}

        {storageType === 'postgres' && (
          <>
            <Form.Item name="host" label="PostgreSQL Host">
              <Input />
            </Form.Item>
            <Form.Item name="port" label="Port">
              <Input />
            </Form.Item>
            <Form.Item name="user" label="User">
              <Input />
            </Form.Item>
            <Form.Item name="password" label="Password">
              <Input.Password />
            </Form.Item>
            <Form.Item name="name" label="DB Name">
              <Input />
            </Form.Item>
            <Form.Item name="options" label="Options">
              <Input />
            </Form.Item>
          </>
        )}

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">
              {storageUUID === 'add' ? 'Create' : 'Update'}
            </Button>
            {storageUUID !== 'add' && (
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
