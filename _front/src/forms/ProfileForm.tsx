import { ReactElement, useEffect } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { Button, Flex, Form, Header, TextField } from '@adobe/react-spectrum'
import { useMutation, useQuery } from '@tanstack/react-query'

import client from '@/api/client'

interface ProfileData {
  first_name: string
  last_name: string
}

export function ProfileForm(): ReactElement {
  const form = useForm<ProfileData>({
    defaultValues: { first_name: '', last_name: '' },
  })

  const query = useQuery({
    queryKey: ['/profile'],
    queryFn: async ({ signal }) => {
      const { data } = await client.GET('/profile', { signal })
      return data
    },
  })

  useEffect(() => {
    if (query.data) {
      form.reset({ first_name: query.data.first_name, last_name: query.data.last_name })
    }
  }, [query.data, form])

  const mutation = useMutation({
    mutationFn: async (data: ProfileData) => {
      const resp = await client.PUT('/profile', { body: data })
      if (resp.error) throw new Error(resp.error.detail)
      return resp
    },
    onSuccess: () => {
      query.refetch()
    },
  })

  const onSubmit = (data: ProfileData) => mutation.mutate(data)

  if (query.isLoading) return <></>

  return (
    <Flex direction="row" justifyContent="center" height="100vh">
      <Form onSubmit={form.handleSubmit(onSubmit)}>
        <Flex direction="column" width="size-4600" gap="size-100">
          <Header marginBottom="size-160">Edit Profile</Header>
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
          <Button type="submit" variant="cta">
            Update
          </Button>
        </Flex>
      </Form>
    </Flex>
  )
}
