import createClient from 'openapi-fetch'
import type { paths } from './v1'

const client = createClient<paths>({
  baseUrl: '/api/v1',
  credentials: 'include',
})

export default client
