export type HttpClientOptions = {
  baseURL: string
  getToken?: () => string | null
}

export function createHttpClient(options: HttpClientOptions) {
  const { baseURL, getToken } = options

  async function request<T>(
    path: string,
    init: RequestInit & { json?: unknown } = {},
  ): Promise<T> {
    const headers = new Headers(init.headers)
    if (init.json !== undefined) {
      headers.set('Content-Type', 'application/json')
    }
    const token = getToken?.()
    if (token) {
      headers.set('Authorization', `Bearer ${token}`)
    }

    const res = await fetch(`${baseURL}${path}`, {
      ...init,
      headers,
      credentials: 'include',
      body: init.json !== undefined ? JSON.stringify(init.json) : init.body,
    })

    const text = await res.text()
    let parsed: unknown
    try {
      parsed = text ? JSON.parse(text) : null
    } catch {
      throw new Error(`Invalid JSON: ${text.slice(0, 200)}`)
    }

    const envelope = parsed as { code?: number | string; message?: string }
    if (!res.ok) {
      throw new Error(envelope.message ?? res.statusText)
    }
    const bizCode =
      envelope.code === undefined || envelope.code === null
        ? 200
        : Number(envelope.code)
    if (!Number.isNaN(bizCode) && bizCode !== 200) {
      throw new Error(envelope.message ?? `业务错误 ${bizCode}`)
    }
    return parsed as T
  }

  return {
    get: <T>(path: string) => request<T>(path, { method: 'GET' }),
    post: <T>(path: string, json: unknown) => request<T>(path, { method: 'POST', json }),
  }
}
