import { queryOptions } from '@tanstack/react-query'
import { AxiosError } from "axios";
import { FrontendAPI } from './api'

export const sessionOptions = () => {
  return queryOptions({
    queryKey: ['session'],
    queryFn: async () => {
      try {
        const { data } = await FrontendAPI.toSession();
        return data;
      } catch (error) {
        if (error instanceof AxiosError) {
          if (error.response?.status === 401) {
            return {
              active: false,
            };
          }
        }
        return Promise.reject(error);
      }
    },
  })
}
