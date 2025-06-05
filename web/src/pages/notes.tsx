import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useNotes, useCreateNote, useUpdateNote, useDeleteNote } from '@/hooks/useNotes'
import { Plus, Search, Edit3, Trash2, FileText, Clock, Sparkles } from 'lucide-react'
import type { Note } from '@/types/api'

type NotesProps = {
  onLogout: () => void
}

export default function Notes({ onLogout }: NotesProps) {
  const [selectedNote, setSelectedNote] = useState<Note | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [isEditing, setIsEditing] = useState(false)

  const { data: notes = [], isLoading, error } = useNotes()
  const createNoteMutation = useCreateNote()
  const updateNoteMutation = useUpdateNote()
  const deleteNoteMutation = useDeleteNote()

  // Filter notes based on search query
  const filteredNotes = notes.filter(note =>
    note.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
    note.content.toLowerCase().includes(searchQuery.toLowerCase())
  )

  useEffect(() => {
    if (selectedNote) {
      setTitle(selectedNote.title)
      setContent(selectedNote.content)
      setIsEditing(false)
    }
  }, [selectedNote])

  // Auto-save with debounce
  useEffect(() => {
    if (!selectedNote || !isEditing) return

    const timer = setTimeout(() => {
      if (title.trim() && (title !== selectedNote.title || content !== selectedNote.content)) {
        updateNoteMutation.mutate({
          id: selectedNote.id,
          title: title.trim(),
          content: content.trim(),
        })
      }
    }, 1000)

    return () => clearTimeout(timer)
  }, [title, content, selectedNote, isEditing, updateNoteMutation])

  const handleCreateNote = () => {
    createNoteMutation.mutate(
      { title: 'Untitled Note', content: '' },
      {
        onSuccess: (response) => {
          const newNote = notes.find(note => note.id === response.id)
          if (newNote) {
            setSelectedNote(newNote)
            setIsEditing(true)
          }
        }
      }
    )
  }

  const handleDeleteNote = (noteId: string) => {
    deleteNoteMutation.mutate(noteId, {
      onSuccess: () => {
        if (selectedNote?.id === noteId) {
          setSelectedNote(null)
          setTitle('')
          setContent('')
        }
      }
    })
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('en-US', { 
      month: 'short', 
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const getPreview = (content: string) => {
    return content.length > 80 ? content.substring(0, 80) + '...' : content
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-violet-50 via-pink-50 to-cyan-50 flex items-center justify-center p-4">
        <div className="bg-white/80 backdrop-blur-xl rounded-2xl shadow-2xl border border-white/20 p-8 text-center">
          <div className="text-red-500 mb-4">
            <FileText className="w-12 h-12 mx-auto mb-4" />
            <h2 className="text-xl font-semibold">Error loading notes</h2>
            <p className="text-sm text-gray-600 mt-2">{error.message}</p>
          </div>
          <Button onClick={onLogout} variant="outline">
            Back to Login
          </Button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-violet-50 via-pink-50 to-cyan-50">
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute -top-1/2 -right-1/2 w-96 h-96 bg-gradient-to-br from-purple-400/10 to-pink-400/10 rounded-full blur-3xl"></div>
        <div className="absolute -bottom-1/2 -left-1/2 w-96 h-96 bg-gradient-to-tr from-blue-400/10 to-cyan-400/10 rounded-full blur-3xl"></div>
      </div>

      <div className="relative h-screen flex">
        <div className="w-80 bg-white/80 backdrop-blur-xl border-r border-white/20 flex flex-col">
          <div className="p-6 border-b border-gray-100">
            <div className="flex items-center justify-between mb-4">
              <div className="flex items-center space-x-2">
                <div className="p-2 bg-gradient-to-r from-pink-500 to-purple-600 rounded-full shadow-lg">
                  <Sparkles className="w-5 h-5 text-white" />
                </div>
                <h1 className="text-xl font-bold bg-gradient-to-r from-gray-900 to-gray-600 bg-clip-text text-transparent">
                  Quanta Notes
                </h1>
              </div>
              <Button
                onClick={onLogout}
                variant="outline"
                size="sm"
                className="text-gray-500 hover:text-gray-700"
              >
                Logout
              </Button>
            </div>

            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
              <Input
                placeholder="Search notes..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="pl-10 bg-gray-50/50 border-gray-200 focus:border-pink-300 focus:ring-pink-200"
              />
            </div>

            <Button
              onClick={handleCreateNote}
              disabled={createNoteMutation.isPending}
              className="w-full mt-4 bg-gradient-to-r from-pink-500 to-purple-600 hover:from-pink-600 hover:to-purple-700 text-white shadow-lg transition-all duration-200"
            >
              <Plus className="w-4 h-4 mr-2" />
              {createNoteMutation.isPending ? 'Creating...' : 'New Note'}
            </Button>
          </div>

          <div className="flex-1 overflow-y-auto">
            {isLoading ? (
              <div className="p-6 space-y-4">
                {[...Array(3)].map((_, i) => (
                  <div key={i} className="animate-pulse">
                    <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                    <div className="h-3 bg-gray-200 rounded w-full mb-1"></div>
                    <div className="h-3 bg-gray-200 rounded w-2/3"></div>
                  </div>
                ))}
              </div>
            ) : filteredNotes.length === 0 ? (
              <div className="p-6 text-center text-gray-500">
                <FileText className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                <p className="text-sm">
                  {searchQuery ? 'No notes match your search' : 'No notes yet'}
                </p>
                {!searchQuery && (
                  <p className="text-xs mt-2">Create your first note to get started!</p>
                )}
              </div>
            ) : (
              <div className="p-4 space-y-2">
                {filteredNotes.map((note) => (
                  <button
                    key={note.id}
                    onClick={() => setSelectedNote(note)}
                    className={`w-full text-left p-4 rounded-xl transition-all duration-200 group ${
                      selectedNote?.id === note.id
                        ? 'bg-gradient-to-r from-pink-500/10 to-purple-600/10 border border-pink-200'
                        : 'hover:bg-gray-50 border border-transparent'
                    }`}
                  >
                    <div className="flex items-start justify-between mb-2">
                      <h3 className="font-medium text-gray-900 truncate flex-1">
                        {note.title}
                      </h3>
                      <button
                        onClick={(e) => {
                          e.stopPropagation()
                          handleDeleteNote(note.id)
                        }}
                        className="opacity-0 group-hover:opacity-100 p-1 hover:bg-red-100 rounded text-red-500 transition-all"
                      >
                        <Trash2 className="w-3 h-3" />
                      </button>
                    </div>
                    <p className="text-sm text-gray-600 mb-2 line-clamp-2">
                      {getPreview(note.content)}
                    </p>
                    <div className="flex items-center text-xs text-gray-400">
                      <Clock className="w-3 h-3 mr-1" />
                      {formatDate(note.updated_at)}
                    </div>
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
        <div className="flex-1 flex flex-col">
          {selectedNote ? (
            <>
              <div className="p-6 bg-white/50 backdrop-blur-sm border-b border-gray-100">
                <div className="flex items-center justify-between">
                  <div className="flex-1 max-w-md">
                    <Label htmlFor="title" className="text-sm font-medium text-gray-700 mb-2 block">
                      Title
                    </Label>
                    <Input
                      id="title"
                      value={title}
                      onChange={(e) => {
                        setTitle(e.target.value)
                        setIsEditing(true)
                      }}
                      className="text-lg font-semibold border-0 bg-transparent p-0 focus:ring-0 focus:border-0"
                      placeholder="Note title..."
                    />
                  </div>
                  <div className="flex items-center space-x-2">
                    {updateNoteMutation.isPending && (
                      <div className="flex items-center text-sm text-gray-500">
                        <div className="w-3 h-3 border-2 border-gray-300 border-t-pink-500 rounded-full animate-spin mr-2"></div>
                        Saving...
                      </div>
                    )}
                    <Edit3 className="w-5 h-5 text-gray-400" />
                  </div>
                </div>
              </div>
              <div className="flex-1 p-6">
                <Label htmlFor="content" className="text-sm font-medium text-gray-700 mb-2 block">
                  Content
                </Label>
                <textarea
                  id="content"
                  value={content}
                  onChange={(e) => {
                    setContent(e.target.value)
                    setIsEditing(true)
                  }}
                  placeholder="Start writing your note..."
                  className="w-full h-full resize-none border-0 bg-transparent p-0 focus:ring-0 focus:outline-none text-gray-700 leading-relaxed"
                />
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center">
              <div className="text-center">
                <div className="p-4 bg-gradient-to-r from-pink-500/10 to-purple-600/10 rounded-full inline-block mb-4">
                  <FileText className="w-12 h-12 text-gray-400" />
                </div>
                <h2 className="text-xl font-semibold text-gray-900 mb-2">
                  Select a note to start editing
                </h2>
                <p className="text-gray-500 mb-6">
                  Choose a note from the sidebar or create a new one
                </p>
                <Button
                  onClick={handleCreateNote}
                  className="bg-gradient-to-r from-pink-500 to-purple-600 hover:from-pink-600 hover:to-purple-700 text-white"
                >
                  <Plus className="w-4 h-4 mr-2" />
                  Create Your First Note
                </Button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
