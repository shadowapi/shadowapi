import { useEffect, useState } from 'react'
import { Controller, useForm } from 'react-hook-form'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Button, Flex, Form, Header, Link, Text, TextField, View } from '@adobe/react-spectrum'
import { RegistrationFlow } from '@ory/client'
import { AxiosError } from 'axios'
import { FrontendAPI } from './api'
import { handleFlowError } from './handler_flow_error'

interface FormFields {
  email: string
  password: string
  passwordConfirm: string
  firstName: string
  lastName: string
  csrfToken: string
}

export function SignupPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const [flow, setFlow] = useState<RegistrationFlow>()
  const form = useForm({
    defaultValues: {
      email: '',
      password: '',
      passwordConfirm: '',
      firstName: '',
      lastName: '',
      csrfToken: '',
    },
  })

  const flow_id = searchParams.get('flow_id')

  useEffect(() => {
    const startRegistrationFlow = async () => {
      const { data } = await FrontendAPI.createBrowserRegistrationFlow()
      navigate(`/signup?flow_id=${data.id}`)
    }

    const getRegistrationFlow = async (flowID: string) => {
      FrontendAPI.getRegistrationFlow({ id: flowID })
        .then(({ data }) => setFlow(data))
        .catch(() => startRegistrationFlow())
    }

    if (flow_id) {
      getRegistrationFlow(flow_id)
      return
    }

    startRegistrationFlow()
  }, [flow_id, navigate])

  // if we have a flow, we map the errors to the form fields
  useEffect(() => {
    if (!flow) {
      return
    }
    const fields: { [key: string]: keyof FormFields } = {
      'traits.email': 'email',
      password: 'password',
      csrf_token: 'csrfToken',
      'traits.name.first': 'firstName',
      'traits.name.last': 'lastName',
    }
    const data = flow as RegistrationFlow
    data.ui.nodes.forEach((node: any) => {
      if (node.attributes.name in fields) {
        const fieldName = fields[node.attributes.name]
        if (node.attributes.value) {
          form.setValue(fieldName, node.attributes.value)
        }
        if (node.messages.length > 0) {
          form.setError(fieldName, { type: 'manual', message: node.messages[0] })
        }
      }
    })
  }, [flow, form])

  const onSubmit = async (fields: FormFields) => {
    if (fields.password !== fields.passwordConfirm) {
      form.setError('passwordConfirm', { type: 'manual', message: 'Passwords do not match' })
      return
    }
    FrontendAPI.updateRegistrationFlow({
      flow: flow!.id,
      updateRegistrationFlowBody: {
        csrf_token: fields.csrfToken,
        method: 'password',
        password: fields.password,
        traits: {
          email: fields.email,
          name: {
            first: fields.firstName,
            last: fields.lastName,
          },
        },
      },
    })
      .then(async () => {
        navigate('/')
      })
      .catch(handleFlowError(navigate, 'signup', form.reset))
      .catch((err: AxiosError) => {
        if (err.response?.status === 400) {
          const flowData = err.response.data as RegistrationFlow
          setFlow(flowData)
          if (flowData.ui.messages) {
            form.setError('email', { type: 'manual', message: flowData.ui.messages[0].text })
          }
          return
        }
        return Promise.reject(err)
      })
  }
  return (
    <Flex direction="row" alignItems="center" justifyContent="center" flexBasis="100%" height="100vh">
      <View padding="size-200" backgroundColor="gray-200" borderRadius="medium" width="size-3600">
        <Form onSubmit={form.handleSubmit(onSubmit)}>
          <Flex direction="column" gap="size-100">
            <Header>Sign Up</Header>
            <input type="hidden" name="csrfToken" value={form.watch('csrfToken')} />
            <Controller
              name="email"
              control={form.control}
              rules={{ required: 'Email is required' }}
              render={({ field: { name, value, onChange, onBlur, ref }, fieldState: { invalid, error } }) => (
                <TextField
                  label="Email"
                  type="email"
                  width="100%"
                  isRequired
                  name={name}
                  value={value}
                  onChange={onChange}
                  onBlur={onBlur}
                  ref={ref}
                  validationState={invalid ? 'invalid' : undefined}
                  errorMessage={error?.message}
                />
              )}
            />
            <Controller
              name="password"
              control={form.control}
              rules={{ required: 'Password is required' }}
              render={({ field: { name, value, onChange, onBlur, ref }, fieldState: { invalid, error } }) => (
                <TextField
                  label="Password"
                  type="password"
                  width="100%"
                  isRequired
                  name={name}
                  value={value}
                  onChange={onChange}
                  onBlur={onBlur}
                  ref={ref}
                  validationState={invalid ? 'invalid' : undefined}
                  errorMessage={error?.message}
                />
              )}
            />
            <Controller
              name="passwordConfirm"
              control={form.control}
              rules={{ required: 'Password confirmation is required' }}
              render={({ field: { name, value, onChange, onBlur, ref }, fieldState: { invalid, error } }) => (
                <TextField
                  label="Repeat Password"
                  type="password"
                  width="100%"
                  isRequired
                  name={name}
                  value={value}
                  onChange={onChange}
                  onBlur={onBlur}
                  ref={ref}
                  validationState={invalid ? 'invalid' : undefined}
                  errorMessage={error?.message}
                />
              )}
            />
            <Controller
              name="firstName"
              control={form.control}
              render={({ field: { name, value, onChange, onBlur, ref }, fieldState: { invalid, error } }) => (
                <TextField
                  label="First Name"
                  type="text"
                  width="100%"
                  name={name}
                  value={value}
                  onChange={onChange}
                  onBlur={onBlur}
                  ref={ref}
                  validationState={invalid ? 'invalid' : undefined}
                  errorMessage={error?.message}
                />
              )}
            />
            <Controller
              name="lastName"
              control={form.control}
              render={({ field: { name, value, onChange, onBlur, ref }, fieldState: { invalid, error } }) => (
                <TextField
                  label="Last Name"
                  type="text"
                  width="100%"
                  name={name}
                  value={value}
                  onChange={onChange}
                  onBlur={onBlur}
                  ref={ref}
                  validationState={invalid ? 'invalid' : undefined}
                  errorMessage={error?.message}
                />
              )}
            />
            <Button variant="cta" alignSelf="end" marginTop="size-150" width="size-1250" type="submit">
              Submit
            </Button>
            <Text alignSelf="end" marginTop="size-100">
              Already have an account? <Link href="/login">Sign in</Link>
            </Text>
          </Flex>
        </Form>
      </View>
    </Flex>
  )
}
