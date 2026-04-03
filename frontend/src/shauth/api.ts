import { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'

export function useLogout() {
  const navigate = useNavigate()
  return useCallback(async () => {
    await fetch('/logout', { credentials: 'include' })
    navigate('/login')
  }, [navigate])
}
