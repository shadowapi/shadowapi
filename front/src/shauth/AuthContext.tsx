import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react'

export interface AuthUser {
  email: string
  sessionToken: string
  sessionId?: string
}

interface AuthContextType {
  user: AuthUser | null
  isAuthenticated: boolean
  isLoading: boolean
  login: (email: string, sessionToken: string, sessionId?: string) => void
  logout: () => Promise<void>
  checkAuth: () => Promise<boolean>
}

const AuthContext = createContext<AuthContextType | undefined>(undefined)

const AUTH_STORAGE_KEY = 'shadowapi_auth'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  // Load auth from localStorage on mount
  useEffect(() => {
    const storedAuth = localStorage.getItem(AUTH_STORAGE_KEY)
    if (storedAuth) {
      try {
        const authData = JSON.parse(storedAuth)
        setUser(authData)
      } catch (error) {
        console.error('Failed to parse stored auth data:', error)
        localStorage.removeItem(AUTH_STORAGE_KEY)
      }
    }
    setIsLoading(false)
  }, [])

  const login = (email: string, sessionToken: string, sessionId?: string) => {
    const userData: AuthUser = { email, sessionToken, sessionId }
    setUser(userData)
    localStorage.setItem(AUTH_STORAGE_KEY, JSON.stringify(userData))
  }

  const logout = async () => {
    if (user?.sessionId) {
      try {
        await fetch(`http://auth.localtest.me/v2/sessions/${user.sessionId}`, {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${user.sessionToken}`,
          },
        })
      } catch (error) {
        console.error('Failed to logout from Zitadel:', error)
      }
    }

    setUser(null)
    localStorage.removeItem(AUTH_STORAGE_KEY)
  }

  const checkAuth = async (): Promise<boolean> => {
    if (!user?.sessionToken) {
      return false
    }

    try {
      // TODO: Implement session validation with Zitadel
      // For now, just check if we have a session token
      // In the future, we can make a request to validate the token

      // Simple validation - check if token exists and is not empty
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
