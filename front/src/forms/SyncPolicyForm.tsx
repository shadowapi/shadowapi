import { ReactElement, useEffect } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { Button, Flex, Form, Header, Switch, TextArea, TextField } from '@adobe/react-spectrum'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import client from '@/api/client'
import type { components } from '@/api/v1'

type SyncPolicyFormData = {
  user_id: string
  service: string
  blocklist?: string[]
  exclude_list?: string[]
  sync_all: boolean
  settings?: string
}

export function SyncPolicyForm({ policyUUID }: { policyUUID: string }): ReactElement {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const form = useForm<SyncPolicyFormData>({})

  const isAdd = policyUUID === 'add'

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
          form.setError('service', { message: resp.error.detail })
          throw new Error(resp.error.detail)
        }
        return resp
      } else {
        const resp = await client.PUT('/syncpolicy/' + policyUUID, { body: payload })
        if (resp.error) {
          form.setError('service', { message: resp.error.detail })
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
        form.setError('service', { message: resp.error.detail })
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

  if (query.isPending && !isAdd) return <></>

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" height="100vh">
      <Form onSubmit={form.handleSubmit(onSubmit)}>
        <Flex direction="column" width="size-4600" gap="size-100">
          <Header marginBottom="size-160">{isAdd ? 'Add Sync Policy' : 'Edit Sync Policy'}</Header>
          <Controller
            name="user_id"
            control={form.control}
            rules={{ required: 'User ID is required' }}
            render={({ field, fieldState }) => (
              <TextField
                label="User ID"
                isRequired
                type="text"
                width="100%"
                {...field}
                validationState={fieldState.invalid ? 'invalid' : undefined}
                errorMessage={fieldState.error?.message}
              />
            )}
          />
          <Controller
            name="service"
            control={form.control}
            rules={{ required: 'Service is required' }}
            render={({ field, fieldState }) => (
              <TextField
                label="Service"
                isRequired
                type="text"
                width="100%"
                {...field}
                validationState={fieldState.invalid ? 'invalid' : undefined}
                errorMessage={fieldState.error?.message}
              />
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
