import { ReactElement, useCallback, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Form, Input, Select, Space, Switch, Typography, message } from 'antd'
import { useSWRConfig } from 'swr'

import apiClient from '@/api/client'
import { useApiGet } from '@/api/hooks'
import type { components } from '@/api/v1'

const timezones = [
  { value: 'Pacific/Pago_Pago', label: 'Pacific/Pago_Pago (UTC -11:00)' },
  { value: 'Pacific/Honolulu', label: 'Pacific/Honolulu (UTC -10:00)' },
  { value: 'America/Anchorage', label: 'America/Anchorage (UTC -9:00)' },
  { value: 'America/Los_Angeles', label: 'America/Los_Angeles (UTC -8:00)' },
  { value: 'America/Denver', label: 'America/Denver (UTC -7:00)' },
  { value: 'America/Chicago', label: 'America/Chicago (UTC -6:00)' },
  { value: 'America/New_York', label: 'America/New_York (UTC -5:00)' },
  { value: 'America/Halifax', label: 'America/Halifax (UTC -4:00)' },
  { value: 'America/Sao_Paulo', label: 'America/Sao_Paulo (UTC -3:00)' },
  { value: 'America/Noronha', label: 'America/Noronha (UTC -2:00)' },
  { value: 'Atlantic/Azores', label: 'Atlantic/Azores (UTC -1:00)' },
  { value: 'Europe/London', label: 'Europe/London (UTC +0:00)' },
  { value: 'Europe/Paris', label: 'Europe/Paris (UTC +1:00)' },
  { value: 'Europe/Athens', label: 'Europe/Athens (UTC +2:00)' },
  { value: 'Europe/Moscow', label: 'Europe/Moscow (UTC +3:00)' },
  { value: 'Asia/Dubai', label: 'Asia/Dubai (UTC +4:00)' },
  { value: 'Asia/Karachi', label: 'Asia/Karachi (UTC +5:00)' },
  { value: 'Asia/Dhaka', label: 'Asia/Dhaka (UTC +6:00)' },
  { value: 'Asia/Bangkok', label: 'Asia/Bangkok (UTC +7:00)' },
  { value: 'Asia/Hong_Kong', label: 'Asia/Hong_Kong (UTC +8:00)' },
  { value: 'Asia/Tokyo', label: 'Asia/Tokyo (UTC +9:00)' },
  { value: 'Australia/Sydney', label: 'Australia/Sydney (UTC +10:00)' },
  { value: 'Pacific/Noumea', label: 'Pacific/Noumea (UTC +11:00)' },
  { value: 'Pacific/Auckland', label: 'Pacific/Auckland (UTC +12:00)' },
]

type SchedulerFormData = {
  pipeline_uuid: string
  schedule_type: string
  cron_expression?: string | null
  run_at?: string | null
  timezone?: string
  is_enabled: boolean
  is_paused: boolean
  next_run?: string
  last_run?: string
}

type CronExpressionInputProps = {
  value?: string
  onChange?: (value: string) => void
}

function CronExpressionInput({ value = '* * * * *', onChange }: CronExpressionInputProps) {
  const parseCron = useCallback((cron: string) => {
    const parts = cron.trim().split(/\s+/)
    return {
      minute: parts[0] || '*',
      hour: parts[1] || '*',
      dayOfMonth: parts[2] || '*',
      month: parts[3] || '*',
      dayOfWeek: parts[4] || '*',
    }
  }, [])

  const [minute, setMinute] = useState(parseCron(value).minute)
  const [hour, setHour] = useState(parseCron(value).hour)
  const [dayOfMonth, setDayOfMonth] = useState(parseCron(value).dayOfMonth)
  const [month, setMonth] = useState(parseCron(value).month)
  const [dayOfWeek, setDayOfWeek] = useState(parseCron(value).dayOfWeek)

  useEffect(() => {
    onChange?.(`${minute} ${hour} ${dayOfMonth} ${month} ${dayOfWeek}`)
  }, [minute, hour, dayOfMonth, month, dayOfWeek, onChange])

  useEffect(() => {
    const parts = parseCron(value)
    setMinute(parts.minute)
    setHour(parts.hour)
    setDayOfMonth(parts.dayOfMonth)
    setMonth(parts.month)
    setDayOfWeek(parts.dayOfWeek)
  }, [value, parseCron])

  const monthOptions = [
    { value: '*', label: '*' },
    { value: '1', label: 'Jan' },
    { value: '2', label: 'Feb' },
    { value: '3', label: 'Mar' },
    { value: '4', label: 'Apr' },
    { value: '5', label: 'May' },
    { value: '6', label: 'Jun' },
    { value: '7', label: 'Jul' },
    { value: '8', label: 'Aug' },
    { value: '9', label: 'Sep' },
    { value: '10', label: 'Oct' },
    { value: '11', label: 'Nov' },
    { value: '12', label: 'Dec' },
  ]

  const dayOfWeekOptions = [
    { value: '*', label: '*' },
    { value: '1', label: 'Mon' },
    { value: '2', label: 'Tue' },
    { value: '3', label: 'Wed' },
    { value: '4', label: 'Thu' },
    { value: '5', label: 'Fri' },
    { value: '6', label: 'Sat' },
    { value: '7', label: 'Sun' },
  ]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
      <Input
        addonBefore="Minute"
        value={minute}
        onChange={(e) => setMinute(e.target.value)}
        style={{ width: 200 }}
      />
      <Input
        addonBefore="Hour"
        value={hour}
        onChange={(e) => setHour(e.target.value)}
        style={{ width: 200 }}
      />
      <Input
        addonBefore="Day of Month"
        value={dayOfMonth}
        onChange={(e) => setDayOfMonth(e.target.value)}
        style={{ width: 200 }}
      />
      <div>
        <Typography.Text style={{ marginRight: 8 }}>Month</Typography.Text>
        <Select value={month} onChange={setMonth} style={{ width: 120 }} options={monthOptions} />
      </div>
      <div>
        <Typography.Text style={{ marginRight: 8 }}>Day of Week</Typography.Text>
        <Select value={dayOfWeek} onChange={setDayOfWeek} style={{ width: 120 }} options={dayOfWeekOptions} />
      </div>
      <Typography.Text type="secondary" style={{ fontSize: '0.8rem' }}>
        Cron format: [Minute] [Hour] [Day of Month] [Month] [Day of Week] - Week starts from Monday
      </Typography.Text>
    </div>
  )
}

export function SchedulerForm({ schedulerUUID }: { schedulerUUID: string }): ReactElement {
  const navigate = useNavigate()
  const { mutate: globalMutate } = useSWRConfig()
  const [form] = Form.useForm<SchedulerFormData>()

  const scheduleType = Form.useWatch('schedule_type', form)

  const isAdd = schedulerUUID === 'add'

  const { data: pipelinesData, isLoading: pipelinesLoading } = useApiGet<{
    pipelines: components['schemas']['pipeline'][]
  }>('/pipeline')

  const { data: schedulerData, isLoading: schedulerLoading } = useApiGet<SchedulerFormData>(
    isAdd ? null : `/scheduler/${schedulerUUID}`
  )

  useEffect(() => {
    if (schedulerData && !isAdd) {
      form.setFieldsValue(schedulerData)
    }
  }, [schedulerData, isAdd, form])

  // Set a dummy run_at value if not already set.
  useEffect(() => {
    if (!form.getFieldValue('run_at')) {
      const now = new Date()
      const formattedNow = now.toISOString().slice(0, 16)
      form.setFieldValue('run_at', formattedNow)
    }
  }, [scheduleType, form])

  const onSubmit = async (values: SchedulerFormData) => {
    try {
      const transformed = { ...values }
      if (values.schedule_type === 'cron') {
        transformed.run_at = new Date().toISOString()
      } else {
        if (!values.run_at) {
          transformed.run_at = new Date().toISOString()
        } else if (values.run_at.length === 16) {
          transformed.run_at = values.run_at + ':00Z'
        } else {
          transformed.run_at = values.run_at
        }
      }
      if (isAdd) {
        await apiClient.post('/scheduler', transformed)
      } else {
        await apiClient.put(`/scheduler/${schedulerUUID}`, transformed)
      }
      message.success(isAdd ? 'Scheduler created' : 'Scheduler updated')
      globalMutate('/scheduler')
      globalMutate((key: string) => typeof key === 'string' && key.startsWith('/scheduler'), undefined, {
        revalidate: true,
      })
      navigate('/schedulers')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  const onDelete = async () => {
    try {
      await apiClient.delete(`/scheduler/${schedulerUUID}`)
      message.success('Scheduler deleted')
      globalMutate('/scheduler')
      navigate('/schedulers')
    } catch (err: any) {
      const detail = err?.response?.data?.detail || err.message
      message.error(detail)
    }
  }

  if (!isAdd && schedulerLoading) return <></>
  if (pipelinesLoading) return <></>

  return (
    <div style={{ display: 'flex', justifyContent: 'center', minHeight: '100vh' }}>
      <Form
        form={form}
        onFinish={onSubmit}
        layout="horizontal" labelCol={{ span: 6 }} wrapperCol={{ span: 14 }}
        style={{ width: 400 }}
        initialValues={{
          pipeline_uuid: '',
          schedule_type: 'cron',
          cron_expression: '0 * * * *',
          run_at: '',
          timezone: 'Europe/London',
          is_enabled: true,
          is_paused: false,
        }}
      >
        <Typography.Title level={4}>{isAdd ? 'Add Scheduler' : 'Edit Scheduler'}</Typography.Title>

        <Form.Item
          name="pipeline_uuid"
          label="Pipeline"
          rules={[{ required: true, message: 'Pipeline is required' }]}
        >
          <Select>
            {pipelinesData?.pipelines?.map((pipeline) => (
              <Select.Option key={pipeline.uuid} value={pipeline.uuid}>
                {pipeline.name} {pipeline.type}
              </Select.Option>
            ))}
          </Select>
        </Form.Item>

        <Form.Item
          name="schedule_type"
          label="Schedule Type"
          rules={[{ required: true, message: 'Schedule type is required' }]}
        >
          <Select>
            <Select.Option value="cron">Cron</Select.Option>
            <Select.Option value="one_time">One Time</Select.Option>
          </Select>
        </Form.Item>

        {scheduleType === 'cron' && (
          <>
            <Form.Item label="Select a Cron Template">
              <Select
                placeholder="Pick a scenario"
                onChange={(val) => {
                  const templates: Record<string, string> = {
                    every_10_min: '*/10 * * * *',
                    every_hour: '0 * * * *',
                    every_6_hours: '0 */6 * * *',
                    every_night_2am: '0 2 * * *',
                    every_sunday_2am: '0 2 * * 7',
                  }
                  if (templates[val]) {
                    form.setFieldValue('cron_expression', templates[val])
                  }
                }}
              >
                <Select.Option value="every_10_min">Every 10 mins</Select.Option>
                <Select.Option value="every_hour">Every Hour</Select.Option>
                <Select.Option value="every_6_hours">Every 6 Hours</Select.Option>
                <Select.Option value="every_night_2am">Every night at 2am</Select.Option>
                <Select.Option value="every_sunday_2am">Every Sunday at 2am</Select.Option>
              </Select>
            </Form.Item>

            <Form.Item
              name="cron_expression"
              label="Cron Expression"
              rules={[
                { required: true, message: 'Cron expression is required' },
                {
                  validator: (_, value) => {
                    if (!value) return Promise.resolve()
                    const parts = value.trim().split(/\s+/)
                    return parts.length === 5
                      ? Promise.resolve()
                      : Promise.reject('Cron expression must have exactly 5 fields')
                  },
                },
              ]}
            >
              <CronExpressionInput />
            </Form.Item>
          </>
        )}

        {scheduleType === 'one_time' && (
          <Form.Item name="run_at" label="Run At" rules={[{ required: true, message: 'Run At is required' }]}>
            <Input type="datetime-local" />
          </Form.Item>
        )}

        <Form.Item name="timezone" label="Timezone" rules={[{ required: true, message: 'Timezone is required' }]}>
          <Select showSearch optionFilterProp="label" options={timezones} />
        </Form.Item>

        <Form.Item name="is_enabled" label="Enabled" valuePropName="checked">
          <Switch />
        </Form.Item>

        <Form.Item name="is_paused" label="Paused" valuePropName="checked">
          <Switch />
        </Form.Item>

        {!isAdd && (
          <>
            <Form.Item name="last_run" label="Last Run">
              <Input disabled />
            </Form.Item>
            <Form.Item name="next_run" label="Next Run">
              <Input disabled />
            </Form.Item>
          </>
        )}

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
