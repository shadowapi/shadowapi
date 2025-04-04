import { ReactElement, useCallback, useEffect, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { Button, Flex, Form, Grid, Header, Item, Picker, Switch, Text, TextField } from '@adobe/react-spectrum'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import client from '@/api/client'
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
  next_run?: string
  last_run?: string
}

type CronExpressionInputProps = {
  value: string
  onChange: (value: string) => void
  errorMessage?: string
}

function CronExpressionInput({ value, onChange, errorMessage }: CronExpressionInputProps) {
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
    onChange(`${minute} ${hour} ${dayOfMonth} ${month} ${dayOfWeek}`)
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
    { key: '*', label: '*' },
    { key: '1', label: 'Jan' },
    { key: '2', label: 'Feb' },
    { key: '3', label: 'Mar' },
    { key: '4', label: 'Apr' },
    { key: '5', label: 'May' },
    { key: '6', label: 'Jun' },
    { key: '7', label: 'Jul' },
    { key: '8', label: 'Aug' },
    { key: '9', label: 'Sep' },
    { key: '10', label: 'Oct' },
    { key: '11', label: 'Nov' },
    { key: '12', label: 'Dec' },
  ]
  const dayOfWeekOptions = [
    { key: '*', label: '*' },
    { key: '1', label: 'Mon' },
    { key: '2', label: 'Tue' },
    { key: '3', label: 'Wed' },
    { key: '4', label: 'Thu' },
    { key: '5', label: 'Fri' },
    { key: '6', label: 'Sat' },
    { key: '7', label: 'Sun' },
  ]

  return (
    <>
      <Flex direction="column" gap="size-100">
        <TextField label="Minute" value={minute} onChange={setMinute} width="48px" />
        <TextField label="Hour" value={hour} onChange={setHour} width="48px" />
        <TextField label="Day of Month" value={dayOfMonth} onChange={setDayOfMonth} width="128px" />

        <Picker label="Month" selectedKey={month} onSelectionChange={(key) => setMonth(key.toString())} width="124px">
          {monthOptions.map((opt) => (
            <Item key={opt.key}>{opt.label}</Item>
          ))}
        </Picker>
        <Picker
          label="Day of Week"
          selectedKey={dayOfWeek}
          onSelectionChange={(key) => setDayOfWeek(key.toString())}
          width="124px"
        >
          {dayOfWeekOptions.map((opt) => (
            <Item key={opt.key}>{opt.label}</Item>
          ))}
        </Picker>
      </Flex>
      {errorMessage && <Text UNSAFE_style={{ fontSize: '0.8rem', color: 'red' }}>{errorMessage}</Text>}
      <Text marginTop="size-100" UNSAFE_style={{ fontSize: '0.8rem', color: 'gray' }}>
        Cron format: [Minute] [Hour] [Day of Month] [Month] [Day of Week] – Week starts from Monday
      </Text>
    </>
  )
}

export function SchedulerForm({ schedulerUUID }: { schedulerUUID: string }): ReactElement {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const form = useForm<SchedulerFormData>({
    defaultValues: {
      pipeline_uuid: '',
      schedule_type: 'cron',
      cron_expression: '0 * * * *',
      run_at: '',
      timezone: 'Europe/London',
      is_enabled: true,
      is_paused: false,
    },
  })

  const scheduleType = form.watch('schedule_type')

  // Set a dummy run_at value if not already set.
  // (Since the datetime-local input only returns "YYYY-MM-DDTHH:mm",
  // we'll override this in onSubmit so backend receives a valid RFC3339 timestamp.)
  useEffect(() => {
    if (!form.getValues('run_at')) {
      const now = new Date()
      // Use datetime-local format as default ("YYYY-MM-DDTHH:mm")
      const formattedNow = now.toISOString().slice(0, 16)
      form.setValue('run_at', formattedNow)
    }
  }, [scheduleType, form])

  const isAdd = schedulerUUID === 'add'
  const pipelinesQuery = useQuery({
    queryKey: ['pipelines'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/pipeline', { signal })
      return data as { pipelines: components['schemas']['pipeline'][] }
    },
  })
  const schedulerQuery = useQuery({
    queryKey: isAdd ? ['/scheduler', 'add'] : ['/scheduler', { uuid: schedulerUUID }],
    queryFn: async ({ signal }) => {
      if (isAdd) return {}
      const { data } = await client.GET('/scheduler/' + schedulerUUID, { signal })
      return data as SchedulerFormData
    },
    enabled: !isAdd,
  })
  useEffect(() => {
    if (schedulerQuery.data && !isAdd) {
      form.reset(schedulerQuery.data)
    }
  }, [schedulerQuery.data, isAdd, form])
  const mutation = useMutation({
    mutationFn: async (data: SchedulerFormData) => {
      if (isAdd) {
        const resp = await client.POST('/scheduler', { body: data })
        if (resp.error) throw new Error(resp.error.detail)
        return resp
      } else {
        const resp = await client.PUT('/scheduler/' + schedulerUUID, { body: data })
        if (resp.error) throw new Error(resp.error.detail)
        return resp
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/scheduler' })
      navigate('/schedulers')
    },
  })
  const deleteMutation = useMutation({
    mutationFn: async (uuid: string) => {
      const resp = await client.DELETE('/scheduler/' + uuid)
      if (resp.error) throw new Error(resp.error.detail)
      return resp
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/scheduler' })
      navigate('/schedulers')
    },
  })
  const onSubmit = (data: SchedulerFormData) => {
    const transformed = { ...data }
    if (data.schedule_type === 'cron') {
      // For cron type, set a dummy run_at value as current time (full ISO string)
      transformed.run_at = new Date().toISOString()
    } else {
      // For one_time, convert the datetime-local value ("YYYY-MM-DDTHH:mm") into full ISO string
      if (!data.run_at) {
        transformed.run_at = new Date().toISOString()
      } else if (data.run_at.length === 16) {
        transformed.run_at = data.run_at + ':00Z'
      } else {
        transformed.run_at = data.run_at
      }
    }
    mutation.mutate(transformed)
  }
  const onDelete = () => {
    deleteMutation.mutate(schedulerUUID)
  }
  if (!isAdd && schedulerQuery.isPending) return <></>
  if (pipelinesQuery.isPending) return <></>
  return (
    <Flex direction="row" justifyContent="center" height="100vh">
      <Form onSubmit={form.handleSubmit(onSubmit)}>
        <Flex direction="column" width="size-4600" gap="size-100">
          <Header marginBottom="size-160">{isAdd ? 'Add Scheduler' : 'Edit Scheduler'}</Header>
          <Controller
            name="pipeline_uuid"
            control={form.control}
            rules={{ required: 'Pipeline is required' }}
            render={({ field, fieldState }) => (
              <Picker
                label="Pipeline"
                isRequired
                selectedKey={field.value}
                onSelectionChange={(key) => field.onChange(key.toString())}
                errorMessage={fieldState.error?.message}
                width="100%"
              >
                {pipelinesQuery.data?.pipelines?.map((pipeline) => (
                  <Item key={pipeline.uuid}>
                    <span
                      style={{
                        whiteSpace: 'nowrap',
                        height: '24px',
                        lineHeight: '24px',
                        marginLeft: 10,
                        marginRight: 10,
                      }}
                    >
                      {pipeline.name} {pipeline.type}
                    </span>
                  </Item>
                ))}
              </Picker>
            )}
          />
          <Controller
            name="schedule_type"
            control={form.control}
            rules={{ required: 'Schedule type is required' }}
            render={({ field, fieldState }) => (
              <Picker
                label="Schedule Type"
                isRequired
                selectedKey={field.value}
                onSelectionChange={(key) => field.onChange(key.toString())}
                errorMessage={fieldState.error?.message}
                width="100%"
              >
                <Item key="cron">Cron</Item>
                <Item key="one_time">One Time</Item>
              </Picker>
            )}
          />
          {scheduleType === 'cron' && (
            <Controller
              name="cron_expression"
              control={form.control}
              rules={{
                required: 'Cron expression is required',
                validate: (value: string) => {
                  const parts = value.trim().split(/\s+/)
                  return parts.length === 5 || 'Cron expression must have exactly 5 fields'
                },
              }}
              render={({ field, fieldState }) => (
                <CronExpressionInput
                  value={field.value || ''}
                  onChange={field.onChange}
                  errorMessage={fieldState.error?.message}
                />
              )}
            />
          )}
          {scheduleType === 'one_time' && (
            <Controller
              name="run_at"
              control={form.control}
              rules={{ required: 'Run At is required' }}
              render={({ field, fieldState }) => (
                <TextField
                  label="Run At"
                  isRequired
                  type="datetime-local"
                  width="100%"
                  validationState={fieldState.invalid ? 'invalid' : undefined}
                  errorMessage={fieldState.error?.message}
                  {...field}
                />
              )}
            />
          )}
          <Controller
            name="timezone"
            control={form.control}
            rules={{ required: 'Timezone is required' }}
            render={({ field, fieldState }) => (
              <Picker
                label="Timezone"
                isRequired
                selectedKey={field.value}
                onSelectionChange={(key) => field.onChange(key.toString())}
                errorMessage={fieldState.error?.message}
                width="100%"
              >
                {timezones.map((tz) => (
                  <Item key={tz.value}>{tz.label}</Item>
                ))}
              </Picker>
            )}
          />
          <Controller
            name="is_enabled"
            control={form.control}
            render={({ field }) => (
              <Switch isSelected={field.value} onChange={field.onChange}>
                Enabled
              </Switch>
            )}
          />
          <Controller
            name="is_paused"
            control={form.control}
            render={({ field }) => (
              <Switch isSelected={field.value} onChange={field.onChange}>
                Paused
              </Switch>
            )}
          />
          {!isAdd && (
            <>
              <Controller
                name="last_run"
                control={form.control}
                render={({ field }) => <TextField label="Last Run" isDisabled value={field.value || ''} width="100%" />}
              />
              <Controller
                name="next_run"
                control={form.control}
                render={({ field }) => <TextField label="Next Run" isDisabled value={field.value || ''} width="100%" />}
              />
            </>
          )}
          <Flex direction="row" gap="size-100" marginTop="size-300" justifyContent="center">
            <Button type="submit" variant="cta">
              {isAdd ? 'Create' : 'Update'}
            </Button>
            {!isAdd && (
              <Button variant="negative" onPress={onDelete}>
                Delete
              </Button>
            )}
          </Flex>
        </Flex>
      </Form>
    </Flex>
  )
}
