import {
  Form,
  Flex,
  View,
  Header,
  Button,
  TextField,
  Link,
  Text,
} from '@adobe/react-spectrum';

import { useEffect, useState } from "react"
import { useNavigate, useSearchParams } from "react-router-dom"
import { useForm, Controller } from "react-hook-form"
import { LoginFlow, UiNode } from '@ory/client'
import { AxiosError } from "axios";

import { FrontendAPI } from './api'
import { handleFlowError } from './handler_flow_error'

import { useQuery, useQueryClient } from '@tanstack/react-query'
import { sessionOptions } from './query'

interface FormFields {
  email: string;
  password: string;
  csrfToken: string;
}

export function LoginPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [flow, setFlow] = useState<LoginFlow>()
  const queryClient = useQueryClient()
  const session = useQuery(sessionOptions())

  const form = useForm({
    defaultValues: { email: "", password: "", csrfToken: "" },
  })

  useEffect(() => {
    const flowID = searchParams.get('flow_id')
    if (flowID) {
      FrontendAPI.getLoginFlow({ id: searchParams.get('flow_id')! })
        .then(({ data }) => { setFlow(data) })
      return
    }
    FrontendAPI.createBrowserLoginFlow().
      then(({ data }) => { navigate(`/login?flow_id=${data.id}`) })
  }, [searchParams, navigate])

  useEffect(() => {
    if (!flow) { return }
    const fields: { [key: string]: keyof FormFields; } = {
      'identifier': 'email',
      'password': 'password',
      'csrf_token': 'csrfToken',
    }
    const data = flow as LoginFlow
    data.ui.nodes.forEach((node: UiNode) => {
      if ('name' in node.attributes) {
        if (node.attributes.name in fields) {
          const fieldName = fields[node.attributes.name]
          if (node.attributes.value) { form.setValue(fieldName, node.attributes.value) }
          if (node.messages.length > 0) {
            form.setError(fieldName, { type: 'manual', message: node.messages[0].text })
          }
        }
      }
    })
  }, [flow, form])

  const onSubmit = async (fields: FormFields) => {
    FrontendAPI.updateLoginFlow({
      flow: flow!.id,
      updateLoginFlowBody: {
        method: 'password',
        csrf_token: fields.csrfToken,
        identifier: fields.email,
        password: fields.password,
      },
    })
      .then(({ data }) => {
        queryClient.setQueryData(sessionOptions().queryKey, data.session)
        navigate('/')
      })
      .catch(handleFlowError(navigate, 'signup', form.reset))
      .catch((err: AxiosError) => {
        if (err.response?.status === 400) {
          const flowData = err.response.data as LoginFlow
          setFlow(flowData)
          if (flowData.ui.messages) {
            form.setError('email', { type: 'manual', message: flowData.ui.messages[0].text })
          }
          return;
        }
        return Promise.reject(err);
      })
  }

  useEffect(() => {
    if (session.data?.active) {
      navigate('/')
    }
  }, [session.data, navigate])

  if (session.isPending) {
    return <span>Loading...</span>
  }
  if (session.isError) {
    return <span>Error '{session.error.name}': {session.error.message}</span>
  }

  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
      <View
        padding="size-200"
        backgroundColor="gray-200"
        borderRadius="medium"
        width="size-3600"
      >
        <Form onSubmit={form.handleSubmit(onSubmit)}>
          <Flex direction="column" gap="size-100">
            <Header>Login</Header>
            <input type="hidden" name="csrfToken" value={form.watch('csrfToken')} />
            <Controller
              name="email"
              control={form.control}
              rules={{ required: 'Email is required' }}
              render={({
                field: { name, value, onChange, onBlur, ref },
                fieldState: { invalid, error },
              }) => (
                <TextField
                  label="Email" type="email" width="100%" isRequired
                  name={name} value={value} onChange={onChange} onBlur={onBlur} ref={ref}
                  validationState={invalid ? 'invalid' : undefined} errorMessage={error?.message}
                />
              )}
            />
            <Controller
              name="password"
              control={form.control}
              rules={{ required: 'Password is required' }}
              render={({
                field: { name, value, onChange, onBlur, ref },
                fieldState: { invalid, error },
              }) => (
                <TextField
                  label="Password" type="password" width="100%" isRequired
                  name={name} value={value} onChange={onChange} onBlur={onBlur} ref={ref}
                  validationState={invalid ? 'invalid' : undefined} errorMessage={error?.message}
                />
              )}
            />
            <Button variant="cta" alignSelf="end" marginTop="size-150" width="size-1250" type="submit">Login</Button>
            <Text alignSelf="end" marginTop="size-100">
              Don&apos;t have an account?{" "}
              <Link href="/signup" alignSelf="end" marginTop="size-100">Sign up</Link>
            </Text>

          </Flex>
        </Form>
      </View>
    </Flex>
  )
}
