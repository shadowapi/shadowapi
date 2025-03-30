import { ReactElement, useEffect } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate } from 'react-router-dom'
import { Button, Flex, Form, Header, Switch, TextField } from '@adobe/react-spectrum'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import client from '@/api/client'
import type { components } from '@/api/v1'

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
  const queryClient = useQueryClient()
  const form = useForm<UserFormData>({
    defaultValues: {
      email: '',
      password: '',
      first_name: '',
      last_name: '',
      is_enabled: false,
      is_admin: false,
    },
  })

  const isAdd = userUUID === 'add'

  const query = useQuery({
    queryKey: isAdd ? ['/user', 'add'] : ['/user', { uuid: userUUID }],
    queryFn: async ({ signal }) => {
      if (isAdd) return {}
      const { data } = await client.GET('/user/' + userUUID, { signal })
      return data
    },
    enabled: !isAdd,
  })

  useEffect(() => {
    if (query.data && !isAdd) {
      form.reset(query.data)
    }
  }, [query.data, isAdd, form])

  const mutation = useMutation({
    mutationFn: async (data: UserFormData) => {
      if (isAdd) {
        const resp = await client.POST('/user', { body: data })
        if (resp.error) {
          form.setError('email', { message: resp.error.detail })
          throw new Error(resp.error.detail)
        }
        return resp
      } else {
        const resp = await client.PUT('/user/' + userUUID, { body: data })
        if (resp.error) {
          form.setError('email', { message: resp.error.detail })
          throw new Error(resp.error.detail)
        }
        return resp
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/user' })
      navigate('/users')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (uuid: string) => {
      const resp = await client.DELETE('/user/' + uuid)
      if (resp.error) {
        form.setError('email', { message: resp.error.detail })
        throw new Error(resp.error.detail)
      }
      return resp
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: '/user' })
      navigate('/users')
    },
  })

  const onSubmit = (data: UserFormData) => {
    mutation.mutate(data)
  }

  const onDelete = () => {
    deleteMutation.mutate(userUUID)
  }

  if (query.isLoading && !isAdd) return <></>

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" height="100vh">
      <Form onSubmit={form.handleSubmit(onSubmit)}>
        <Flex direction="column" width="size-4600" gap="size-100">
          <Header marginBottom="size-160">{isAdd ? 'Add User' : 'Edit User'}</Header>
          <Controller
            name="email"
            control={form.control}
            rules={{ required: 'Email is required' }}
            render={({ field, fieldState }) => (
              <TextField
                label="Email"
                isRequired
                type="email"
                width="100%"
                {...field}
                validationState={fieldState.invalid ? 'invalid' : undefined}
                errorMessage={fieldState.error?.message}
              />
            )}
          />
          <Controller
            name="password"
            control={form.control}
            rules={{ required: isAdd ? 'Password is required' : false }}
            render={({ field, fieldState }) => (
              <TextField
                label="Password"
                isRequired={isAdd}
                type="password"
                width="100%"
                {...field}
                validationState={fieldState.invalid ? 'invalid' : undefined}
                errorMessage={fieldState.error?.message}
              />
            )}
          />
          <Controller
            name="first_name"
            control={form.control}
            rules={{ required: 'First name is required' }}
            render={({ field, fieldState }) => (
              <TextField
                label="First Name"
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
            name="last_name"
            control={form.control}
            rules={{ required: 'Last name is required' }}
            render={({ field, fieldState }) => (
              <TextField
                label="Last Name"
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
            name="is_enabled"
            control={form.control}
            render={({ field }) => (
              <Switch isSelected={field.value} onChange={field.onChange}>
                Enabled
              </Switch>
            )}
          />
          <Controller
            name="is_admin"
            control={form.control}
            render={({ field }) => (
              <Switch isSelected={field.value} onChange={field.onChange}>
                Admin
              </Switch>
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
