import { getToken } from "./auth";

const BASE_URL = process.env.NEXT_PUBLIC_API_URL!;

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

export async function apiFetch(path: string, options?: RequestInit): Promise<unknown>;
export async function apiFetch<T>(path: string, options?: RequestInit): Promise<T>;
export async function apiFetch<T = unknown>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  if (!BASE_URL) {
    throw new Error("NEXT_PUBLIC_API_URL is not set");
  }

  const token = getToken();

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...headersToObject(options.headers),
    },
  });

  if (res.status === 401) {
    if (typeof window !== "undefined") {
      localStorage.removeItem("token");
      window.location.href = "/login";
    }
    throw new Error("Unauthorized");
  }

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    let message = text || "API error";
    if (text) {
      try {
        const json = JSON.parse(text) as { error?: string };
        if (json && typeof json.error === "string" && json.error.length > 0) {
          message = json.error;
        }
      } catch {
        // ignore parse errors
      }
    }

    throw new Error(message);
  }

  return res.json() as Promise<T>;
}
