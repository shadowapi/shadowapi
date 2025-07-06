import { useCallback } from "react";

export const useLogout = () =>
  useCallback(() => {
    fetch("/logout/callback", {
      method: "GET",
      credentials: "include",
    }).finally(() => {
      const base = import.meta.env.VITE_ZITADEL_INSTANCE_URL;
      if (base) {
        const logoutUrl = `${base}/oidc/v2/logout?post_logout_redirect_uri=${encodeURIComponent(import.meta.env.VITE_ZITADEL_REDIRECT_URI)}`;
        window.location.href = logoutUrl;
      } else {
        window.location.href = "/";
      }
    });
  }, []);
