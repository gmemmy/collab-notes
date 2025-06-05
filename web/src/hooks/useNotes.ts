import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { notesApi } from '@/services/api'
import type { UpdateNoteRequest } from '@/types/api'

const NOTES_QUERY_KEY = ['notes'] as const

export function useNotes() {
  return useQuery({
    queryKey: NOTES_QUERY_KEY,
    queryFn: notesApi.getNotes,
    staleTime: 5 * 60 * 1000, 
  })
}

export function useCreateNote() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: notesApi.createNote,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: NOTES_QUERY_KEY })
    },
  })
}

export function useUpdateNote() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, ...note }: { id: string } & UpdateNoteRequest) =>
      notesApi.updateNote(id, note),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: NOTES_QUERY_KEY })
    },
  })
}

export function useDeleteNote() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: notesApi.deleteNote,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: NOTES_QUERY_KEY })
    },
  })
} 