import { useCallback } from "react";

export const useLogout = () =>
  useCallback(() => {
    const hasZitadel = document.cookie
      .split("; ")
      .some((c) => c.startsWith("zitadel_access_token="));
    if (hasZitadel) {
      const logoutUrl = `${import.meta.env.VITE_ZITADEL_INSTANCE_URL}/oidc/v2/logout?post_logout_redirect_uri=${encodeURIComponent(window.location.origin + "/logout/callback")}`;
      window.location.href = logoutUrl;
    } else {
      window.location.href = "/logout/callback";
    }
  }, []);
