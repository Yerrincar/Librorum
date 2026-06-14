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

export type LoginRequest = {
  username: string
  password: string
}

export type UserResponse = {
  id?: number
  username: string
  email?: string
  display_name?: string
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

export async function loginUser(payload: LoginRequest): Promise<UserResponse> {
  const response = await fetch('/users/login', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    credentials: 'include',
    body: JSON.stringify(payload),
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Login failed'))
  }

  return response.json()
}

export async function getCurrentUser(): Promise<UserResponse> {
  const response = await fetch('/users/currentUser', {
    method: 'GET',
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Current user failed'))
  }

  return response.json()
}

export async function logoutUser(): Promise<void> {
  const response = await fetch('/users/logout', {
    method: 'POST',
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Logout failed'))
  }
}

async function readErrorMessage(response: Response, fallback = 'Registration failed'): Promise<string> {
  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    const body = (await response.json()) as ApiErrorResponse
    return body.error ?? body.message ?? fallback
  }

  const text = await response.text()
  return text || fallback
}
