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

export type ImportMetadataBookRequest = Omit<ImportBookRequest, 'file'> & {
  metadata: BookMetadataCandidateResponse
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

export type MetadataSource = 'openlibrary' | 'google_books'

export type BookMetadataCandidateResponse = {
  source: MetadataSource
  source_id: string
  title: string
  author: string
  description: string
  genres: string[] | null
  language: string
  publication_year: number | null
  cover_id: number
  cover_url: string
  work_key: string
}

type ApiErrorResponse = {
  error?: string
  message?: string
}

export async function importEpubBook(payload: ImportBookRequest): Promise<LibraryItemResponse> {
  const form = bookFormData(payload)
  form.append('file', payload.file)

  const response = await fetch('/books/import/epub', {
    method: 'POST',
    credentials: 'include',
    body: form,
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Book import failed'))
  }

  return response.json()
}

export async function importMetadataBook(
  payload: ImportMetadataBookRequest,
): Promise<LibraryItemResponse> {
  const form = bookFormData(payload)
  form.append('selected_source', payload.metadata.source)
  form.append('selected_source_id', payload.metadata.source_id)
  form.append('selected_title', payload.metadata.title)
  form.append('selected_author', payload.metadata.author)
  form.append('selected_description', payload.metadata.description)
  form.append('selected_language', payload.metadata.language)
  form.append('selected_work_key', payload.metadata.work_key)
  form.append('selected_cover_id', String(payload.metadata.cover_id))
  form.append('selected_cover_url', payload.metadata.cover_url)
  if (payload.metadata.publication_year !== null) {
    form.append('selected_publication_year', String(payload.metadata.publication_year))
  }
  for (const genre of payload.metadata.genres ?? []) {
    form.append('selected_genres', genre)
  }

  const response = await fetch('/books/import/openlibrary', {
    method: 'POST',
    credentials: 'include',
    body: form,
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Metadata import failed'))
  }

  return response.json()
}

export async function searchBookMetadata(
  title: string,
  author: string,
): Promise<BookMetadataCandidateResponse[]> {
  const form = new FormData()
  form.append('title', title.trim())
  if (author.trim() !== '') {
    form.append('author', author.trim())
  }

  const response = await fetch('/books/openlibrary/search', {
    method: 'POST',
    credentials: 'include',
    body: form,
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Metadata search failed'))
  }

  return response.json()
}

function bookFormData(payload: Omit<ImportBookRequest, 'file'>): FormData {
  const form = new FormData()
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

  return form
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
