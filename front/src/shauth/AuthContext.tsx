import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'

export interface AuthUser {
  email: string
  sessionToken: string
  sessionId: string
}

interface AuthContextType {
  user: AuthUser | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (email: string, sessionToken: string, sessionId: string) => void
  logout: () => Promise<void>
  checkAuth: () => Promise<boolean>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

const AUTH_STORAGE_KEY = 'shadowapi_auth'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  // Load auth from sessionStorage on mount
  useEffect(() => {
    const storedAuth = sessionStorage.getItem(AUTH_STORAGE_KEY)
    if (storedAuth) {
      try {
        const authData: AuthUser = JSON.parse(storedAuth)
        setUser(authData)
      } catch (error) {
        console.error('Failed to parse stored auth data:', error)
        sessionStorage.removeItem(AUTH_STORAGE_KEY)
      }
    }
    setIsLoading(false)
  }, [])

  const login = (email: string, sessionToken: string, sessionId: string) => {
    const userData: AuthUser = { email, sessionToken, sessionId }
    setUser(userData)
    // Store in sessionStorage instead of localStorage for better security
    sessionStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(userData))
  }

  const logout = async () => {
    if (user?.sessionId && user?.sessionToken) {
      try {
        const zitadelUrl = import.meta.env.VITE_ZITADEL_URL || 'http://auth.localtest.me'
        await fetch(`${zitadelUrl}/v2/sessions/${user.sessionId}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${user.sessionToken}`,
          },
        })
      } catch (error) {
        console.error('Failed to delete Zitadel session:', error)
      }
    }

    setUser(null)
    sessionStorage.removeItem(AUTH_STORAGE_KEY)
  }

  const checkAuth = async (): Promise<boolean> => {
    if (!user?.sessionToken) {
      return false
    }

    try {
      // Session token exists - consider the user authenticated
      // TODO: Optionally validate the session with Zitadel
      return user.sessionToken.length > 0
    } catch (error) {
      console.error('Auth validation failed:', error)
      logout()
      return false
    }
  }

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    logout,
    checkAuth
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
