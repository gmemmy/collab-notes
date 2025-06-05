export type ApiErrorResponse = {
  error: string
}

export type AuthSuccessResponse = {
  token: string
}

export type LoginRequest = {
  email: string
  password: string
}

export type Note = {
  id: string
  user_id: string
  title: string
  content: string
  created_at: string
  updated_at: string
}

export type CreateNoteRequest = {
  title: string
  content: string
}

export type CreateNoteResponse = {
  id: string
}

export type UpdateNoteRequest = {
  title: string
  content: string
}

export class ApiError extends Error {
  constructor(message: string, public status: number) {
    super(message)
    this.name = 'ApiError'
  }
}

 