import { logout } from "./auth";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL;

function headersToObject(headers: HeadersInit | undefined): Record<string, string> {
  if (!headers) return {};
  if (typeof Headers !== "undefined" && headers instanceof Headers) {
    return Object.fromEntries(headers.entries());
  }
  if (Array.isArray(headers)) {
    return Object.fromEntries(headers);
  }
  return headers as Record<string, string>;
}

export async function apiFetch<T = any>(path: string, options: RequestInit = {}): Promise<T> {
  if (!BASE_URL) {
    throw new Error("NEXT_PUBLIC_API_URL is not set");
  }

  const token = typeof window !== "undefined" ? localStorage.getItem("token") : null;

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...headersToObject(options.headers),
    },
  });

  const text = await res.text();
  let data: any = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch {
    data = text;
  }

  if (!res.ok) {
    if (res.status === 401) logout();
    const msg =
      (data && (data.error || data.message)) ||
      (typeof data === "string" ? data : "") ||
      `Request failed (${res.status})`;
    throw new Error(msg);
  }

  return data as T;
}
