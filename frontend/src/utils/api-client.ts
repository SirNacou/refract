import { env } from '../env'

export class ApiError extends Error {
  status: number
  code?: string
  details?: unknown

  constructor(
    message: string,
    status: number,
    code?: string,
    details?: unknown,
  ) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}

type RequestOptions = {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: unknown
  headers?: HeadersInit
  signal?: AbortSignal
}

// Access token provider for future auth integration (T046)
let accessTokenProvider: null | (() => Promise<string | null>) = null

export function setAccessTokenProvider(fn: () => Promise<string | null>) {
  accessTokenProvider = fn
}

async function request<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const { method = 'GET', body, headers = {}, signal } = opts

  // Build full URL from base + path
  const url = new URL(path, env.VITE_API_BASE_URL).toString()

  // Get auth token if provider is configured
  const token = accessTokenProvider ? await accessTokenProvider() : null

  // Build headers
  const requestHeaders: HeadersInit = {
    'Accept': 'application/json',
    ...(!!body && { 'Content-Type': 'application/json' }),
    ...(token && { 'Authorization': `Bearer ${token}` }),
    ...headers,
  }

  const res = await fetch(url, {
    method,
    headers: requestHeaders,
    body: body ? JSON.stringify(body) : undefined,
    signal,
  })

  // Handle 204 No Content
  if (res.status === 204) {
    return undefined as T
  }

  const resText = await res.text()
  let resJson: any = null
  try {
    resJson = resText ? JSON.parse(resText) : null
  } catch (e) {
    // Ignore JSON parse errors
  }

  if (!res.ok) {
    const errorMessage =
      resJson?.message ||
      `API request failed with status ${res.status} ${res.statusText}`
    const errorCode = resJson?.code
    const errorDetails = resJson?.details

    throw new ApiError(errorMessage, res.status, errorCode, errorDetails)
  }

  return resJson as T
}

async function get<T>(
  path: string,
  opts: Omit<RequestOptions, 'method' | 'body'> = {},
): Promise<T> {
  return request<T>(path, { ...opts, method: 'GET' })
}

async function post<T>(
  path: string,
  body: unknown,
  opts: Omit<RequestOptions, 'method' | 'body'> = {},
): Promise<T> {
  return request<T>(path, { ...opts, method: 'POST', body })
}

async function put<T>(
  path: string,
  body: unknown,
  opts: Omit<RequestOptions, 'method' | 'body'> = {},
): Promise<T> {
  return request<T>(path, { ...opts, method: 'PUT', body })
}

async function patch<T>(
  path: string,
  body: unknown,
  opts: Omit<RequestOptions, 'method' | 'body'> = {},
): Promise<T> {
  return request<T>(path, { ...opts, method: 'PATCH', body })
}

async function del<T>(
  path: string,
  opts: Omit<RequestOptions, 'method' | 'body'> = {},
): Promise<T> {
  return request<T>(path, { ...opts, method: 'DELETE' })
}

export const apiClient = {
  request,
  get,
  post,
  put,
  patch,
  delete: del,
}
