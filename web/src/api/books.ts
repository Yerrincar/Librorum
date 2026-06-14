export type ImportBookRequest = {
  file: File
  kind: string
  rating: string
  ownership_status: string
  reading_status: string
  publication_status: string
  current_chapter: string
  read_at: string
  notes: string
}

export type LibraryItemResponse = {
  id: number
  title: string
  author: string
  kind: string
}

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

  if (payload.rating.trim() !== '') {
    form.append('rating', payload.rating.trim())
  }
  if (payload.current_chapter.trim() !== '') {
    form.append('current_chapter', payload.current_chapter.trim())
  }
  if (payload.read_at !== '') {
    form.append('read_at', new Date(payload.read_at).toISOString())
  }
  if (payload.notes.trim() !== '') {
    form.append('notes', payload.notes.trim())
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

async function readErrorMessage(response: Response, fallback: string): Promise<string> {
  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    const body = (await response.json()) as ApiErrorResponse
    return body.error ?? body.message ?? fallback
  }

  const text = await response.text()
  return text || fallback
}
