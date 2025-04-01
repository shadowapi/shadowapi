import { ReactElement, useEffect } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { Button, Flex, Form, Header, Item, Picker, Switch, TextArea, TextField } from '@adobe/react-spectrum'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import client from '@/api/client'
import type { components } from '@/api/v1'

type SyncPolicyFormData = {
  pipeline_uuid: string
  blocklist?: string[]
  exclude_list?: string[]
  sync_all: boolean
  settings?: string
}

export function SyncPolicyForm({ policyUUID }: { policyUUID: string }): ReactElement {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const form = useForm<SyncPolicyFormData>({
    defaultValues: {
      settings: '{}',
    },
  })

  const isAdd = policyUUID === 'add'

  const pipelinesQuery = useQuery({
    queryKey: ['pipelines'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/pipeline', { signal })
      return data as components['schemas']['pipeline'][]
    },
  })

  const query = useQuery({
    queryKey: isAdd ? ['/syncpolicy', 'add'] : ['/syncpolicy', { uuid: policyUUID }],
    queryFn: async ({ signal }) => {
      if (isAdd) return {}
      const { data } = await client.GET('/syncpolicy/' + policyUUID, { signal })
      return data
    },
    enabled: !isAdd,
  })

  useEffect(() => {
    if (query.data && !isAdd) {
      const data = { ...query.data }
      if (data.settings) {
        data.settings = JSON.stringify(data.settings, null, 2)
      }
      form.reset(data)
    }
  }, [query.data, isAdd, form])

  const mutation = useMutation({
    mutationFn: async (data: SyncPolicyFormData) => {
      const payload = {
        ...data,
        settings: data.settings ? JSON.parse(data.settings) : null,
      }
      if (isAdd) {
        const resp = await client.POST('/syncpolicy', { body: payload })
        if (resp.error) {
          form.setError('pipeline_uuid', { message: resp.error.detail })
          throw new Error(resp.error.detail)
        }
        return resp
      } else {
        const resp = await client.PUT('/syncpolicy/' + policyUUID, { body: payload })
        if (resp.error) {
          form.setError('pipeline_uuid', { message: resp.error.detail })
          throw new Error(resp.error.detail)
        }
        return resp
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/syncpolicy' })
      navigate('/syncpolicies')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (uuid: string) => {
      const resp = await client.DELETE('/syncpolicy/' + uuid)
      if (resp.error) {
        form.setError('pipeline_uuid', { message: resp.error.detail })
        throw new Error(resp.error.detail)
      }
      return resp
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/syncpolicy' })
      navigate('/syncpolicies')
    },
  })

  const onSubmit = (data: SyncPolicyFormData) => {
    mutation.mutate(data)
  }

  const onDelete = () => {
    deleteMutation.mutate(policyUUID)
  }

  if (query.isPending && pipelinesQuery.isPending && !isAdd) return <></>

  console.log('pipelinesQuery', { pipelinesQuery })

  return (
    <Flex direction="row" justifyContent="center" height="100vh">
      <Form onSubmit={form.handleSubmit(onSubmit)}>
        <Flex direction="column" width="size-4600" gap="size-100">
          <Header marginBottom="size-160">{isAdd ? 'Add Sync Policy' : 'Edit Sync Policy'}</Header>
          <Controller
            name="name"
            control={form.control}
            rules={{ required: 'Name is required' }}
            render={({ field, fieldState }) => (
              <TextField
                label="Name"
                isRequired
                type="text"
                width="100%"
                validationState={fieldState.invalid ? 'invalid' : undefined}
                errorMessage={fieldState.error?.message}
                {...field}
              />
            )}
          />
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
                {pipelinesQuery &&
                  pipelinesQuery.data?.pipelines?.map((pipeline: components['schemas']['pipeline']) => 
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
                  )}
              </Picker>
            )}
          />
          <Controller
            name="sync_all"
            control={form.control}
            render={({ field }) => (
              <Switch isSelected={field.value} onChange={field.onChange}>
                Sync All
              </Switch>
            )}
          />
          <Controller
            name="blocklist"
            control={form.control}
            render={({ field }) => (
              <TextField
                label="Blocklist (comma separated)"
                type="text"
                width="100%"
                {...field}
                onChange={(value) => field.onChange(value.split(',').map((s) => s.trim()))}
              />
            )}
          />
          <Controller
            name="exclude_list"
            control={form.control}
            render={({ field }) => (
              <TextField
                label="Exclude List (comma separated)"
                type="text"
                width="100%"
                {...field}
                onChange={(value) => field.onChange(value.split(',').map((s) => s.trim()))}
              />
            )}
          />
          <Controller
            name="settings"
            control={form.control}
            render={({ field, fieldState }) => (
              <TextArea
                label="Settings (JSON)"
                width="100%"
                {...field}
                validationState={fieldState.invalid ? 'invalid' : undefined}
                errorMessage={fieldState.error?.message}
              />
            )}
          />
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
