import { useCallback } from 'react'
import { FrontendApi, Configuration } from '@ory/client'

export const FrontendAPI = new FrontendApi(
  new Configuration({
    basePath: '/auth/user',
    baseOptions: {
      withCredentials: true,
    },
  }),
);

export const useLogout = () => {
  const logout = useCallback(async () => {
    const { data } = await FrontendAPI.createBrowserLogoutFlow();
    window.location.href = data.logout_url;
  }, []);

  return logout;
};
