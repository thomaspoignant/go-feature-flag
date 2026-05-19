import { API_URL } from "./constants";
import type { APIResponse } from "@/types/api";

export class HttpError extends Error {
  status: number;
  errors?: Record<string, string[]>;
  constructor(
    message: string,
    status: number,
    errors?: Record<string, string[]>
  ) {
    super(message);
    this.status = status;
    this.errors = errors;
  }
}

export async function apiFetch<T = unknown>(
  path: string,
  init: RequestInit = {}
): Promise<APIResponse<T>> {
  const url = path.startsWith("http")
    ? path
    : `${API_URL}${path.startsWith("/") ? "" : "/"}${path}`;

  const res = await fetch(url, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      ...(init.headers ?? {}),
    },
    ...init,
  });

  let body: APIResponse<T> | null = null;
  const text = await res.text();
  if (text) {
    try {
      body = JSON.parse(text) as APIResponse<T>;
    } catch {
      body = null;
    }
  }

  if (!res.ok) {
    throw new HttpError(
      body?.message ?? `Request failed with status ${res.status}`,
      res.status,
      body?.errors
    );
  }

  return body ?? { success: true };
}
