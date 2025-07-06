import { queryOptions } from '@tanstack/react-query'

export const sessionOptions = () => {
  return queryOptions({
    queryKey: ['session'],
    queryFn: async () => {
      const hasCookie = document.cookie
        .split(';')
        .some((c) => c.trim().startsWith('zitadel_access_token='))
      return { active: hasCookie }
    },
  })
}
