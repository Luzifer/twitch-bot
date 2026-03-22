export class HttpError extends Error {
  data: unknown
  status: number

  constructor(status: number, data: unknown) {
    super(`HTTP ${status}`)
    this.status = status
    this.data = data
  }
}

type ApiMethod = 'DELETE' | 'GET' | 'POST' | 'PUT'
let authToken: string | null = null

async function parseResponse<TResponse>(resp: Response): Promise<TResponse> {
  const contentType = resp.headers.get('content-type') || ''

  if (contentType.includes('application/json')) {
    return resp.json() as Promise<TResponse>
  }

  return resp.text() as Promise<TResponse>
}

export function createApiClient(getAuthToken: () => string | null) {
  async function request<TResponse = unknown>(method: ApiMethod, path: string, body?: unknown, withAuth = true): Promise<TResponse | undefined> {
    const headers = new Headers()
    const authToken = getAuthToken()

    if (withAuth && authToken) {
      headers.set('authorization', authToken)
    }

    let payload: BodyInit | undefined
    if (body !== undefined) {
      headers.set('content-type', 'application/json')
      payload = JSON.stringify(body)
    }

    const resp = await fetch(path, {
      body: payload,
      headers,
      method,
    })

    if (!resp.ok) {
      throw new HttpError(resp.status, await resp.text())
    }

    if (resp.status === 201 || resp.status === 204) {
      return undefined
    }

    return await parseResponse<TResponse>(resp)
  }

  return {
    delete<TResponse = undefined>(path: string, withAuth: boolean | Record<string, unknown> = true) {
      return request<TResponse>('DELETE', path, undefined, withAuth !== false)
    },
    get<TResponse = unknown>(path: string, withAuth: boolean | Record<string, unknown> = true) {
      return request<TResponse>('GET', path, undefined, withAuth !== false)
    },
    post<TResponse = undefined>(path: string, body?: unknown, withAuth: boolean | Record<string, unknown> = true) {
      return request<TResponse>('POST', path, body, withAuth !== false)
    },
    put<TResponse = undefined>(path: string, body?: unknown, withAuth: boolean | Record<string, unknown> = true) {
      return request<TResponse>('PUT', path, body, withAuth !== false)
    },
  }
}

export function setApiAuthToken(token: string | null) {
  authToken = token
}

export const api = createApiClient(() => authToken)
