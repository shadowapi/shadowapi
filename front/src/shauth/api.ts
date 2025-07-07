import { useCallback } from "react";

export const useLogout = () =>
  useCallback(() => {
    window.location.href = "/logout";
  }, []);
