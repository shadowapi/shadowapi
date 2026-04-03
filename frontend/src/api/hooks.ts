import useSWR, { type SWRConfiguration } from 'swr'
import apiClient from './client'

const fetcher = (url: string) => apiClient.get(url).then((res) => res.data)

export function useApiGet<T = any>(key: string | null, config?: SWRConfiguration) {
  return useSWR<T>(key, fetcher, config)
}
