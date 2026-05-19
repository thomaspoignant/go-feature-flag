import type { DataProvider } from "@refinedev/core";
import { API_URL } from "./constants";
import { apiFetch } from "./http";

function resourcePath(resource: string, id?: string | number): string {
  const base = `/${resource}`;
  return id !== undefined ? `${base}/${id}` : base;
}

export const dataProvider: DataProvider = {
  getApiUrl: () => API_URL,

  getList: async ({ resource }) => {
    const body = await apiFetch<unknown[]>(resourcePath(resource));
    const data = (body.data ?? []) as unknown[];
    return { data: data as never, total: data.length };
  },

  getOne: async ({ resource, id }) => {
    const body = await apiFetch<unknown>(resourcePath(resource, id));
    return { data: body.data as never };
  },

  create: async ({ resource, variables }) => {
    const body = await apiFetch<unknown>(resourcePath(resource), {
      method: "POST",
      body: JSON.stringify(variables),
    });
    return { data: body.data as never };
  },

  update: async ({ resource, id, variables }) => {
    const body = await apiFetch<unknown>(resourcePath(resource, id), {
      method: "PATCH",
      body: JSON.stringify(variables),
    });
    return { data: body.data as never };
  },

  deleteOne: async ({ resource, id }) => {
    const body = await apiFetch<unknown>(resourcePath(resource, id), {
      method: "DELETE",
    });
    return { data: (body.data ?? null) as never };
  },

  custom: async ({ url, method = "get", payload, headers }) => {
    const body = await apiFetch<unknown>(url, {
      method: method.toUpperCase(),
      body: payload ? JSON.stringify(payload) : undefined,
      headers: headers as HeadersInit | undefined,
    });
    return { data: body.data as never };
  },
};
