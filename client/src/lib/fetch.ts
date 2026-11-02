export type Method = "POST" | "GET" | "DELETE" | "PUT"

interface ApiResponse<T> {
  data?: T
  error?: string
}

const api = "http://localhost:8080/"

export default async function apiFetch<T, R>(
  url: string,
  body: T | undefined,
  method: Method,
): Promise<ApiResponse<R>> {
  try {
    const options: RequestInit = {
      method,
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
      },
    }

    if (body !== undefined && method !== "GET") {
      options.body = JSON.stringify(body)
    }

    const res = await fetch(api+url, options)
    if (!res.ok) {
      throw new Error(`HTTP Error: ${res.status}`)
    }

    if (res.status === 204) {
      return {}
    }

    const data: R = await res.json()

    return { data }
  } catch (err) {
    return {
      error: err instanceof Error ? err.message : "Unknown error",
    }
  }
}
