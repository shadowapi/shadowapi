import createClient, { type Middleware } from 'openapi-fetch'
import type { paths } from './v1'

const AUTH_STORAGE_KEY = 'shadowapi_auth'

// Middleware to add Bearer token to all requests
const authMiddleware: Middleware = {
  async onRequest({ request }) {
    // Get auth data from sessionStorage
    const storedAuth = sessionStorage.getItem(AUTH_STORAGE_KEY)
    if (storedAuth) {
      try {
        const authData = JSON.parse(storedAuth)
        // Check if token is not expired
        if (authData.expiresAt && authData.expiresAt > Date.now()) {
          // Use accessToken (session token) as Bearer token
          request.headers.set('Authorization', `Bearer ${authData.accessToken}`)
        }
      } catch (error) {
        console.error('Failed to parse auth data for request:', error)
      }
    }
    return request
  },
}

const client = createClient<paths>({ baseUrl: '/api/v1' })

// Register the auth middleware
client.use(authMiddleware)

export default client
