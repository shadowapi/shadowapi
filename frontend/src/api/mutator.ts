import axios, { AxiosRequestConfig } from 'axios'

export const apiInstance = <T>(config: AxiosRequestConfig, options?: AxiosRequestConfig): Promise<T> => {
  const source = axios.CancelToken.source()

  const promise = axios({
    ...config,
    ...options,
    baseURL: '/api/v1',
    cancelToken: source.token,
    withCredentials: true,
    headers: {
      'Content-Type': 'application/json',
      ...config.headers,
      ...options?.headers,
    },
  }).then(({ data }) => data)

  // @ts-expect-error - SWR expects cancel method on promise
  promise.cancel = () => {
    source.cancel('Request cancelled')
  }

  return promise
}

export default apiInstance
