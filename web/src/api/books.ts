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

export type ImportExcelBooksRequest = {
  file: File
  spreadsheet: string
}

export type ImportExcelBooksResponse = {
  imported_count: number
  skipped_count: number
  imported: string[]
  skipped: string[]
}

export type UpdateLibraryItemRequest = {
  id: number
  title: string
  author: string
  rating: string
  cover_path: string
  read_at: string
  description: string
  language: string
  genres: string
  ownership_status: string
  reading_status: string
  current_chapter: string
  total_chapters: string
  notes: string
  cover_file?: File
}

export type LibraryItemResponse = {
  id?: number
  title?: string
  author?: string
  kind?: string
  description?: string
  genres?: string[] | null
  language?: string
  rating?: number | null
  ownership_status?: string
  reading_status?: string
  read_at?: string | null
  current_chapter?: number | null
  total_chapters?: number | null
  notes?: string
  cover_path?: string
  Title?: string
  Author?: string
  Kind?: string
  Description?: string
  Genres?: string[] | null
  Language?: string
  Rating?: number | null
  Ownership_status?: string
  Reading_status?: string
  Read_at?: string | null
  Current_chapter?: number | null
  Total_chapters?: number | null
  Notes?: string
  Cover_path?: string
}

export type LibraryItemKind = 'book' | 'manga' | 'manhwa'

export type MetadataSource = 'calibre' | 'openlibrary' | 'google_books'

export type BookMetadataCandidateResponse = {
  source: MetadataSource
  source_id: string
  title: string
  author: string
  description: string
  genres: string[] | null
  language: string
  publication_year: number | null
  isbn: string
  cover_id: number
  cover_url: string
  cover_path: string
  work_key: string
}

type ApiErrorResponse = {
  error?: string
  message?: string
}

const monthYearFormatter = new Intl.DateTimeFormat(undefined, { month: 'long', year: 'numeric' })

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
  if (payload.metadata.cover_path) {
    form.append('selected_cover_path', payload.metadata.cover_path)
  }
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

export async function importExcelBooks(
  payload: ImportExcelBooksRequest,
): Promise<ImportExcelBooksResponse> {
  const form = new FormData()
  form.append('file', payload.file)
  form.append('spreadsheet', payload.spreadsheet.trim())

  const response = await fetch('/books/import/excelImport', {
    method: 'POST',
    credentials: 'include',
    body: form,
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Excel import failed'))
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

export async function updateLibraryItem(payload: UpdateLibraryItemRequest): Promise<LibraryItemResponse> {
  const form = new FormData()
  form.append('title', payload.title.trim())
  form.append('author', payload.author.trim())
  form.append('rating', payload.rating.trim())
  form.append('cover_path', payload.cover_path.trim())
  form.append('description', payload.description.trim())
  form.append('language', payload.language.trim())
  form.append('genres', payload.genres.trim())
  form.append('ownership_status', payload.ownership_status)
  form.append('reading_status', payload.reading_status)
  form.append('current_chapter', payload.current_chapter.trim())
  form.append('total_chapters', payload.total_chapters.trim())
  form.append('notes', payload.notes.trim())
  if (payload.read_at.trim() !== '') {
    form.append('read_at', new Date(payload.read_at).toISOString())
  }
  if (payload.cover_file) {
    form.append('file', payload.cover_file)
  }

  const response = await fetch(`/books/library-items/${payload.id}`, {
    method: 'PUT',
    credentials: 'include',
    body: form,
  })

  if (!response.ok) {
    throw new Error(await readErrorMessage(response, 'Book update failed'))
  }

  return response.json()
}

export function bookTitle(book: LibraryItemResponse): string {
  return book.title ?? book.Title ?? 'Untitled'
}

export function bookAuthor(book: LibraryItemResponse): string {
  return book.author ?? book.Author ?? ''
}

export function bookLanguage(book: LibraryItemResponse): string {
  return book.language ?? book.Language ?? ''
}

export function bookDescription(book: LibraryItemResponse): string {
  return book.description ?? book.Description ?? ''
}

export function bookGenres(book: LibraryItemResponse): string {
  return (book.genres ?? book.Genres ?? []).filter(Boolean).join(', ')
}

export function bookNotes(book: LibraryItemResponse): string {
  return book.notes ?? book.Notes ?? ''
}

export function bookOwnershipStatus(book: LibraryItemResponse): string {
  return book.ownership_status ?? book.Ownership_status ?? ''
}

export function bookReadingStatus(book: LibraryItemResponse): string {
  return book.reading_status ?? book.Reading_status ?? 'unread'
}

export function bookCurrentChapter(book: LibraryItemResponse): string {
  const currentChapter = book.current_chapter ?? book.Current_chapter
  return typeof currentChapter === 'number' ? String(currentChapter) : ''
}

export function bookTotalChapters(book: LibraryItemResponse): string {
  const totalChapters = book.total_chapters ?? book.Total_chapters
  return typeof totalChapters === 'number' ? String(totalChapters) : ''
}

export function rawBookCoverPath(book: LibraryItemResponse): string {
  return book.cover_path ?? book.Cover_path ?? ''
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

export function bookRating(book: LibraryItemResponse): string {
  const rating = book.rating ?? book.Rating
  return typeof rating === 'number' ? String(rating) : ''
}

export function bookReadMonthYear(book: LibraryItemResponse): string {
  const readAt = book.read_at ?? book.Read_at
  if (!readAt) {
    return ''
  }

  const readDate = new Date(readAt)
  return Number.isNaN(readDate.getTime()) ? '' : monthYearFormatter.format(readDate)
}

export function bookReadAtInput(book: LibraryItemResponse): string {
  const readAt = book.read_at ?? book.Read_at
  if (!readAt) {
    return ''
  }

  const readDate = new Date(readAt)
  if (Number.isNaN(readDate.getTime())) {
    return ''
  }

  const localDate = new Date(readDate.getTime() - readDate.getTimezoneOffset() * 60_000)
  return localDate.toISOString().slice(0, 16)
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
