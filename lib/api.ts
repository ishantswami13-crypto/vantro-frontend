export function getApiBase() {
  const api = process.env.NEXT_PUBLIC_API_URL;
  if (!api) {
    throw new Error("NEXT_PUBLIC_API_URL is not set");
  }

  return api.replace(/\/+$/, "");
}

export function getToken() {
  if (typeof window === "undefined") return "";
  return localStorage.getItem("token") || "";
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const apiBase = getApiBase();
  const token = getToken();

  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  const res = await fetch(`${apiBase}${normalizedPath}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options.headers || {}),
    },
    cache: "no-store",
  });

  if (!res.ok) {
    const text = await res.text().catch(() => "");
    throw new Error(text || `Request failed: ${res.status}`);
  }

  return res.json() as Promise<T>;
}
