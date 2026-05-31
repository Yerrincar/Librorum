export type RegisterRequest = {
  username: string
  email: string
  password: string
  display_name: string
}

export type RegisterResponse = {
  id?: number
  username: string
  email: string
  display_name: string
}

type ApiErrorResponse = {
  error?: string
  message?: string
}

export async function registerUser(payload: RegisterRequest): Promise<RegisterResponse> {
  const response = await fetch('/users/register', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response))
  }

  if (response.status === 204) {
    return {
      username: payload.username,
      email: payload.email,
      display_name: payload.display_name,
    }
  }

  return response.json()
}

async function readErrorMessage(response: Response): Promise<string> {
  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    const body = (await response.json()) as ApiErrorResponse
    return body.error ?? body.message ?? 'Registration failed'
  }

  const text = await response.text()
  return text || 'Registration failed'
}
