export type ImportBookRequest = {
  file: File
  kind: string
  rating: string | number
  ownership_status: string
  reading_status: string
  publication_status: string
  current_chapter: string | number
  read_at: string
  notes: string
}

export type LibraryItemResponse = {
  id?: number
  title?: string
  author?: string
  kind?: string
  cover_path?: string
  Title?: string
  Author?: string
  Kind?: string
  Cover_path?: string
}

export type LibraryItemKind = 'book' | 'manga' | 'manhwa'

type ApiErrorResponse = {
  error?: string
  message?: string
}

export async function importEpubBook(payload: ImportBookRequest): Promise<LibraryItemResponse> {
  const form = new FormData()
  form.append('file', payload.file)
  form.append('kind', payload.kind)
  form.append('ownership_status', payload.ownership_status)
  form.append('reading_status', payload.reading_status)
  form.append('publication_status', payload.publication_status)

  const rating = String(payload.rating).trim()
  const currentChapter = String(payload.current_chapter).trim()
  const notes = payload.notes.trim()

  if (rating !== '') {
    form.append('rating', rating)
  }
  if (currentChapter !== '') {
    form.append('current_chapter', currentChapter)
  }
  if (payload.read_at !== '') {
    form.append('read_at', new Date(payload.read_at).toISOString())
  }
  if (notes !== '') {
    form.append('notes', notes)
  }

  const response = await fetch('/books/insert', {
    method: 'POST',
    credentials: 'include',
    body: form,
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Book import failed'))
  }

  return response.json()
}

export async function fetchBooks(kind?: LibraryItemKind): Promise<LibraryItemResponse[]> {
  const params = new URLSearchParams()
  if (kind) {
    params.set('kind', kind)
  }
  const query = params.toString()
  const response = await fetch(`/books/library-items${query ? `?${query}` : ''}`, {
    method: 'GET',
    credentials: 'include',
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Fetch books failed'))
  }

  return response.json()
}

export function bookTitle(book: LibraryItemResponse): string {
  return book.title ?? book.Title ?? 'Untitled'
}

export function bookAuthor(book: LibraryItemResponse): string {
  return book.author ?? book.Author ?? ''
}

export function bookCoverPath(book: LibraryItemResponse): string {
  const coverPath = (book.cover_path ?? book.Cover_path ?? '').trim()
  if (coverPath === '' || coverPath.startsWith('http://') || coverPath.startsWith('https://')) {
    return coverPath
  }

  const normalized = coverPath.replaceAll('\\', '/')
  if (normalized.startsWith('/covers/')) {
    return normalized
  }
  if (normalized.startsWith('covers/')) {
    return `/${normalized}`
  }

  const coversIndex = normalized.lastIndexOf('/covers/')
  if (coversIndex >= 0) {
    return normalized.slice(coversIndex)
  }

  const filename = normalized.split('/').pop()
  return filename ? `/covers/${encodeURIComponent(filename)}` : ''
}

async function readErrorMessage(response: Response, fallback: string): Promise<string> {
  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    const body = (await response.json()) as ApiErrorResponse
    return body.error ?? body.message ?? fallback
  }

  const text = await response.text()
  return text || fallback
}
