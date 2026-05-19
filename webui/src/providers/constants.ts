export const API_URL =
  (import.meta.env.VITE_API_URL as string | undefined) ?? "/api/v1";

export const API_BASE = API_URL.replace(/\/api\/v1\/?$/, "");
