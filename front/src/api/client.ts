import createClient from 'openapi-fetch'
import type { paths } from './v1'
import { getRuntimeConfig } from '../lib/runtime-config'

// API base URL - runtime config (from SSR env) takes precedence over build-time value
const API_BASE_URL = getRuntimeConfig(
  'VITE_API_BASE_URL',
  import.meta.env.VITE_API_BASE_URL || 'http://api.localtest.me/api/v1'
)

const client = createClient<paths>({
  baseUrl: API_BASE_URL,
  credentials: 'include',
})

export default client
