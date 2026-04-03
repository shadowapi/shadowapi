import '@xyflow/react/dist/style.css'
import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Button,
  Form,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
  message,
} from 'antd'
import {
  CalendarOutlined,
  DeleteOutlined,
  EditOutlined,
  PlayCircleOutlined,
  SaveOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons'
import { useSWRConfig } from 'swr'
import {
  addEdge,
  applyEdgeChanges,
  applyNodeChanges,
  Background,
  BackgroundVariant,
  Connection,
  Controls,
  Edge,
  EdgeChange,
  Node,
  NodeChange,
  ReactFlow,
  useReactFlow,
} from '@xyflow/react'

import apiClient from '@/api/client'
import { useApiGet } from '@/api/hooks'
import type { components } from '@/api/v1'
import { CustomNode } from '@/components/CustomeNode'
import { SchedulerForm } from '@/forms/SchedulerForm'

interface PipelineProps {
  pipelineUUID: string
  userUUID: string
}

const initialNodes = [
  {
    id: '1',
    type: 'customNode',
    data: { label: 'Scheduler' },
    position: { x: 250, y: 25 },
  },
  {
    id: '2',
    type: 'customNode',
    data: { label: 'Data Source' },
    position: { x: 250, y: 125 },
  },
  {
    id: '3',
    type: 'customNode',
    data: { label: 'Contact Extractor' },
    position: { x: 250, y: 225 },
  },
  {
    id: '4',
    type: 'customNode',
    data: { label: 'Storage S3' },
    position: { x: 250, y: 350 },
  },
]

const initialEdges = [
  { id: 'e1-2', source: '1', target: '2', animated: true },
  { id: 'e2-3', source: '2', target: '3' },
  { id: 'e2-4', source: '3', target: '4' },
]

export const PipelineForm = ({ pipelineUUID, userUUID }: PipelineProps) => {
  const navigate = useNavigate()
  const { mutate: globalMutate } = useSWRConfig()
  const [form] = Form.useForm<components['schemas']['pipeline']>()
  const rf = useReactFlow()

  const [nodes, setNodes] = useState<Node[]>(initialNodes)
  const [edges, setEdges] = useState<Edge[]>(initialEdges)

  // Scheduler modal state
  const [showSchedulerForm, setShowSchedulerForm] = useState(false)
  const [editingSchedulerUUID, setEditingSchedulerUUID] = useState<'add' | string>('add')

  const openScheduler = (uuid: 'add' | string) => {
    setEditingSchedulerUUID(uuid)
    setShowSchedulerForm(true)
  }

  const closeScheduler = () => {
    setShowSchedulerForm(false)
  }

  const isAdd = pipelineUUID === 'add'

  // ======== Data fetching with SWR ========

  const { data: datasources } = useApiGet<components['schemas']['datasource'][]>('/datasource')
  const { data: storages } = useApiGet<components['schemas']['storage'][]>('/storage')

  const { data: schedulersData, mutate: mutateSchedulers } = useApiGet<components['schemas']['scheduler'][]>(
    isAdd ? null : `/scheduler?pipeline_uuid=${pipelineUUID}`
  )

  const { data: pipelineData, isLoading: pipelineLoading } = useApiGet<components['schemas']['pipeline']>(
    isAdd ? null : `/pipeline/${pipelineUUID}?user_uuid=${userUUID}`
  )

  // Populate form and React Flow with pipeline data
  useEffect(() => {
    if (pipelineData) {
      form.setFieldsValue(pipelineData as any)
      if (pipelineData.flow?.nodes) rf.addNodes(pipelineData.flow.nodes as Node[])
      if (pipelineData.flow?.edges) rf.addEdges(pipelineData.flow.edges as Edge[])
    }
  }, [pipelineData, rf, form])

  // React Flow event handlers
  const onNodesChange = useCallback((changes: NodeChange<Node>[]) => {
    setNodes((nds) => applyNodeChanges(changes, nds))
  }, [])

  const onEdgesChange = useCallback((changes: EdgeChange<Edge>[]) => {
    setEdges((eds) => applyEdgeChanges(changes, eds))
  }, [])

  const onConnect = useCallback((connection: Connection) => {
    setEdges((oldEdges) => addEdge(connection, oldEdges))
  }, [])

  // ======== Submission handlers ========

  const onSubmit = async (values: components['schemas']['pipeline']) => {
    try {
      if (!values.flow) values.flow = {}
      values.flow.nodes = nodes as any
      values.flow.edges = edges as any

      if (isAdd) {
        await apiClient.post('/pipeline', {
          name: values.name || '',
          datasource_uuid: values.datasource_uuid || '',
          storage_uuid: values.storage_uuid || '',
          flow: values.flow || {},
        } as any)
      } else {
        await apiClient.put(`/pipeline/${pipelineUUID}`, {
          name: values.name,
          datasource_uuid: values.datasource_uuid,
          storage_uuid: values.storage_uuid,
          flow: values.flow,
        } as any)
      }
      message.success(isAdd ? 'Pipeline created' : 'Pipeline updated')
      globalMutate('/pipeline')
      navigate('/pipelines')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  const onDelete = async () => {
    try {
      await apiClient.delete(`/pipeline/${pipelineUUID}`)
      message.success('Pipeline deleted')
      globalMutate('/pipeline')
      navigate('/pipelines')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  const deleteScheduler = async (uuid: string) => {
    try {
      await apiClient.delete(`/scheduler/${uuid}`)
      message.success('Scheduler deleted')
      mutateSchedulers()
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  // Drop zone ref for ReactFlow
  const dropZoneRef = useRef<HTMLDivElement>(null)

  const nodeTypes = useMemo(() => ({ customNode: CustomNode }), [])

  const schedulerColumns = [
    {
      title: 'Type',
      dataIndex: 'schedule_type',
      key: 'schedule_type',
    },
    {
      title: 'Cron / RunAt',
      key: 'expr',
      render: (_: any, item: components['schemas']['scheduler']) =>
        item.schedule_type === 'cron'
          ? item.cron_expression
          : new Date(item.run_at ?? '').toLocaleString(),
    },
    {
      title: 'Next Run',
      key: 'next_run',
      render: (_: any, item: components['schemas']['scheduler']) =>
        item.next_run ? new Date(item.next_run).toLocaleString() : '--',
    },
    {
      title: 'Enabled',
      key: 'enabled',
      render: (_: any, item: components['schemas']['scheduler']) => (
        <Tag color={item.is_enabled ? 'green' : 'red'}>{item.is_enabled ? 'On' : 'Off'}</Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      render: (_: any, item: components['schemas']['scheduler']) => (
        <Space size="small">
          <Tooltip title="Edit">
            <Button type="text" size="small" icon={<EditOutlined />} onClick={() => openScheduler(item.uuid!)} />
          </Tooltip>
          <Tooltip title="Remove">
            <Button type="text" size="small" icon={<DeleteOutlined />} onClick={() => deleteScheduler(item.uuid!)} />
          </Tooltip>
        </Space>
      ),
    },
  ]

  if (!isAdd && pipelineLoading) return <></>

  return (
    <div style={{ display: 'flex', width: '100%', height: '100%' }}>
      {/* Left: Form section */}
      <div style={{ width: '50%', padding: 24, borderRight: '1px solid #303030', overflow: 'auto' }}>
        <Form form={form} onFinish={onSubmit} layout="horizontal" labelCol={{ span: 6 }} wrapperCol={{ span: 14 }}>
          <div style={{ display: 'flex', gap: 8, alignItems: 'flex-start' }}>
            <Form.Item
              name="name"
              rules={[{ required: true, message: 'Name is required' }]}
              style={{ flex: 1 }}
            >
              <Input placeholder="Pipeline name" />
            </Form.Item>

            <Tooltip title={isAdd ? 'Create pipeline' : 'Save pipeline'}>
              <Button type="primary" htmlType="submit" icon={<SaveOutlined />} />
            </Tooltip>

            <Tooltip title="Delete pipeline">
              <Button
                danger
                icon={<DeleteOutlined />}
                onClick={onDelete}
                disabled={pipelineLoading || isAdd}
              />
            </Tooltip>
          </div>

          <Form.Item
            name="datasource_uuid"
            label="Data Source"
            rules={[{ required: true, message: 'Datasource is required' }]}
          >
            <Select>
              {datasources?.map((ds) => (
                <Select.Option key={ds.uuid} value={ds.uuid}>
                  {ds.name} {ds.type}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Form.Item
            name="contact_extractor"
            label="Contact Extractor"
            rules={[{ required: true, message: 'Contact extractor is required' }]}
          >
            <Select>
              <Select.Option value="extractor1">Standard Extractor</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="storage_uuid"
            label="Storage"
            rules={[{ required: true, message: 'Storage is required' }]}
          >
            <Select>
              {storages?.map((s) => (
                <Select.Option key={s.uuid} value={s.uuid}>
                  {s.name} {s.type}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>

          <Space style={{ marginBottom: 16 }}>
            <Button icon={<PlayCircleOutlined />} disabled={pipelineLoading || isAdd}>
              Run
            </Button>
            <Button icon={<ThunderboltOutlined />} disabled={pipelineLoading || isAdd}>
              Test
            </Button>
            <Button icon={<CalendarOutlined />} disabled={pipelineLoading || isAdd} onClick={() => openScheduler('add')}>
              Schedule
            </Button>
          </Space>

          {/* Schedulers Table */}
          {!isAdd && (
            <div style={{ marginTop: 16 }}>
              <Typography.Title level={5}>Schedulers</Typography.Title>
              <Table
                dataSource={schedulersData ?? []}
                columns={schedulerColumns}
                rowKey="uuid"
                size="small"
                pagination={false}
              />
            </div>
          )}
        </Form>

        {/* Scheduler Modal */}
        <Modal
          open={showSchedulerForm}
          onCancel={closeScheduler}
          footer={null}
          width={600}
          destroyOnClose
        >
          <SchedulerForm schedulerUUID={editingSchedulerUUID} />
        </Modal>
      </div>

      {/* Right: Flow section */}
      <div style={{ flex: 1, overflow: 'auto' }}>
        <ReactFlow
          nodes={nodes}
          edges={edges}
          onNodesChange={onNodesChange}
          onEdgesChange={onEdgesChange}
          onConnect={onConnect}
          nodeTypes={nodeTypes}
          ref={dropZoneRef}
          fitView
        >
          <Controls />
          <Background id="1" gap={10} variant={BackgroundVariant.Lines} />
        </ReactFlow>
      </div>
    </div>
  )
}
