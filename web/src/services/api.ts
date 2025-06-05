import type { LoginRequest, AuthSuccessResponse, ApiErrorResponse, Note, CreateNoteRequest, CreateNoteResponse, UpdateNoteRequest } from '@/types/api'
import { ApiError } from '@/types/api'

async function apiRequest<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const token = localStorage.getItem('token')
  
  const response = await fetch(`/api${endpoint}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    },
    ...options,
  })

  let data: unknown
  try {
    data = await response.json() as ApiErrorResponse | T
  } catch {
    throw new ApiError(`Server error (${response.status}): ${response.statusText}`, response.status)
  }

  if (!response.ok) {
    const errorData = data as ApiErrorResponse
    throw new ApiError(errorData.error || `Request failed (${response.status})`, response.status)
  }

  return data as T
}


export const authApi = {
  login: (credentials: LoginRequest): Promise<AuthSuccessResponse> =>
    apiRequest<AuthSuccessResponse>('/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    }),

  signUp: (credentials: LoginRequest): Promise<AuthSuccessResponse> =>
    apiRequest<AuthSuccessResponse>('/signup', {
      method: 'POST',
      body: JSON.stringify(credentials),
    }),
}

export const notesApi = {
  getNotes: (): Promise<Note[]> =>
    apiRequest<Note[]>('/notes'),

  createNote: (note: CreateNoteRequest): Promise<CreateNoteResponse> =>
    apiRequest<CreateNoteResponse>('/notes', {
      method: 'POST',
      body: JSON.stringify(note),
    }),

  updateNote: (id: string, note: UpdateNoteRequest): Promise<void> =>
    apiRequest<void>(`/notes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(note),
    }),

  deleteNote: (id: string): Promise<void> =>
    apiRequest<void>(`/notes/${id}`, {
      method: 'DELETE',
    }),
} 