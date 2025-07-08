import { queryOptions } from "@tanstack/react-query";

export const sessionOptions = () => {
  return queryOptions({
    queryKey: ["session"],
    queryFn: async () => {
      const resp = await fetch("/api/v1/session", {
        credentials: "include",
      });
      if (resp.status === 401) {
        return { active: false };
      }
      if (!resp.ok) {
        throw new Error("session check failed");
      }
      return (await resp.json()) as { active: boolean; uuid?: string };
    },
  });
};
