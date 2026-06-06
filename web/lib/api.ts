import type { UserRole } from "@/app/types";

export const API_URL =
  process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export function apiFetch(
  path: string,
  role: UserRole,
  init: RequestInit = {},
): Promise<Response> {
  const headers = new Headers(init.headers);
  headers.set("X-Role", role);
  return fetch(`${API_URL}${path}`, { ...init, headers });
}
