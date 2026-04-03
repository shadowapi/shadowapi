import useSWR from 'swr'

interface Session {
  active: boolean
  uuid?: string
  email?: string
}

export function useSession() {
  return useSWR<Session>('/session', async (url: string) => {
    const resp = await fetch(`/api/v1${url}`, { credentials: 'include' })
    if (resp.status === 401) return { active: false }
    if (!resp.ok) throw new Error('session check failed')
    return resp.json()
  })
}
