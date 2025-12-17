import createClient from 'openapi-fetch'
import type { paths } from './v1'

// API base URL - use environment variable or fallback to default
const API_BASE_URL =
  import.meta.env.VITE_API_BASE_URL || 'http://api.localtest.me/api/v1'

const client = createClient<paths>({
  baseUrl: API_BASE_URL,
  credentials: 'include',
})

export default client
