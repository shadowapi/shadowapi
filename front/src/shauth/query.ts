import { queryOptions } from '@tanstack/react-query'

interface SessionResponse {
  active: boolean
  uuid?: string
  reason?: string
  session?: {
    id?: string
    userId?: string
    loginName?: string
    displayName?: string
    factors?: {
      user?: {
        id?: string
        loginName?: string
        displayName?: string
        organizationId?: string
        verifiedAt?: string
      }
      password?: {
        verifiedAt?: string
      }
      webAuthN?: {
        verifiedAt?: string
        userVerified?: boolean
      }
      intent?: {
        verifiedAt?: string
      }
      totp?: {
        verifiedAt?: string
      }
      otpSMS?: {
        verifiedAt?: string
      }
      otpEmail?: {
        verifiedAt?: string
      }
    }
    expirationDate?: string
    creationDate?: string
    changeDate?: string
    lifetime?: string
  }
}

export const sessionOptions = () => {
  return queryOptions({
    queryKey: ['session'],
    queryFn: async () => {
      // Temporarily disabled for testing - always return inactive session
      console.log('Session check disabled for testing')
      return { active: false, reason: 'disabled_for_testing' }
    },
    retry: false,
    staleTime: 30000,
    gcTime: 60000,
  })
}
