<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getCurrentUser, type UserResponse } from '@/api/auth'
import {
  bookTitle,
  importMetadataBook,
  importEpubBook,
  searchBookMetadata,
  type BookMetadataCandidateResponse,
} from '@/api/books'

const user = ref<UserResponse | null>(null)
const checkingUser = ref(true)
const loading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const selectedFile = ref<File | null>(null)
const importMethod = ref<'epub' | 'metadata'>('epub')
const searchingMetadata = ref(false)
const metadataCandidates = ref<BookMetadataCandidateResponse[]>([])
const selectedMetadata = ref<BookMetadataCandidateResponse | null>(null)

const form = reactive({
  title: '',
  author: '',
  kind: 'book',
  rating: '',
  ownership_status: 'none',
  reading_status: 'unread',
  publication_status: 'unknown',
  current_chapter: '',
  read_at: '',
  notes: '',
})

onMounted(async () => {
  try {
    user.value = await getCurrentUser()
  } catch {
    user.value = null
  } finally {
    checkingUser.value = false
  }
})

function selectFile(event: Event) {
  const input = event.target as HTMLInputElement
  selectedFile.value = input.files?.[0] ?? null
}

async function searchMetadata() {
  if (form.title.trim() === '') {
    errorMessage.value = 'Enter a title first'
    return
  }

  searchingMetadata.value = true
  errorMessage.value = ''
  successMessage.value = ''
  metadataCandidates.value = []
  selectedMetadata.value = null

  try {
    metadataCandidates.value = await searchBookMetadata(form.title, form.author)
    selectedMetadata.value = metadataCandidates.value[0] ?? null
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Metadata search failed'
  } finally {
    searchingMetadata.value = false
  }
}

function metadataSourceLabel(candidate: BookMetadataCandidateResponse): string {
  return candidate.source === 'google_books' ? 'Google Books' : 'OpenLibrary'
}

function metadataCandidateKey(candidate: BookMetadataCandidateResponse): string {
  return `${candidate.source}:${candidate.source_id || candidate.work_key || candidate.title}`
}

async function submitImport() {
  if (importMethod.value === 'epub' && !selectedFile.value) {
    errorMessage.value = 'Select an EPUB file first'
    return
  }
  if (importMethod.value === 'metadata' && !selectedMetadata.value) {
    errorMessage.value = 'Search metadata and select a result first'
    return
  }

  loading.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const book =
      importMethod.value === 'epub'
        ? await importEpubBook({
            file: selectedFile.value as File,
            ...form,
          })
        : await importMetadataBook({
            ...form,
            metadata: selectedMetadata.value as BookMetadataCandidateResponse,
          })
    successMessage.value = `Imported ${bookTitle(book)}`
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Book import failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main>
    <h1>Import books</h1>

    <p v-if="checkingUser">Checking session...</p>

    <section v-else-if="!user">
      <p>You need to be logged in to import books.</p>
      <RouterLink to="/login">Login</RouterLink>
      <RouterLink to="/register">Register</RouterLink>
    </section>

    <form v-else @submit.prevent="submitImport">
      <p>Logged in as {{ user.username }}</p>

      <section>
        <label>
          Import method
          <select v-model="importMethod" name="import_method">
            <option value="epub">EPUB file</option>
            <option value="metadata">Metadata search</option>
          </select>
        </label>
      </section>

      <section v-if="importMethod === 'epub'">
        <h2>EPUB file</h2>
        <label>
          File
          <input type="file" name="file" accept=".epub,application/epub+zip" @change="selectFile" />
        </label>
      </section>

      <section v-else>
        <h2>Metadata search</h2>
        <label>
          Title
          <input v-model.trim="form.title" name="title" required />
        </label>
        <label>
          Author
          <input v-model.trim="form.author" name="author" />
        </label>

        <button type="button" :disabled="searchingMetadata" @click="searchMetadata">
          {{ searchingMetadata ? 'Searching...' : 'Search metadata' }}
        </button>

        <section v-if="metadataCandidates.length > 0" aria-label="Metadata results">
          <h3>Choose a result</h3>
          <label v-for="candidate in metadataCandidates" :key="metadataCandidateKey(candidate)">
            <input v-model="selectedMetadata" type="radio" name="metadata_result" :value="candidate" />
            [{{ metadataSourceLabel(candidate) }}]
            {{ candidate.title }}
            <span v-if="candidate.author">by {{ candidate.author }}</span>
            <span v-if="candidate.publication_year">({{ candidate.publication_year }})</span>
          </label>
        </section>

        <article v-if="selectedMetadata">
          <h3>{{ selectedMetadata.title }}</h3>
          <p>{{ metadataSourceLabel(selectedMetadata) }}</p>
          <p v-if="selectedMetadata.author">{{ selectedMetadata.author }}</p>
          <p v-if="selectedMetadata.publication_year">{{ selectedMetadata.publication_year }}</p>
          <p v-if="selectedMetadata.language">Language: {{ selectedMetadata.language }}</p>
          <p v-if="selectedMetadata.genres?.length">Genres: {{ selectedMetadata.genres.join(', ') }}</p>
          <p v-if="selectedMetadata.description">{{ selectedMetadata.description }}</p>
        </article>
      </section>

      <section v-if="importMethod === 'epub' || selectedMetadata">
        <h2>Library fields</h2>

        <label>
          Kind
          <select v-model="form.kind" name="kind">
            <option value="book">Book</option>
            <option value="manga">Manga</option>
            <option value="manhwa">Manhwa</option>
          </select>
        </label>

        <label>
          Rating
          <input v-model="form.rating" name="rating" type="number" min="0" max="5" step="0.1" />
        </label>

        <label>
          Ownership status
          <select v-model="form.ownership_status" name="ownership_status">
            <option value="none">None</option>
            <option value="owned_physical">Owned physical</option>
            <option value="owned_digital">Owned digital</option>
            <option value="owned_physical_and_digital">Owned physical and digital</option>
            <option value="wishlist">Wishlist</option>
          </select>
        </label>

        <label>
          Reading status
          <select v-model="form.reading_status" name="reading_status">
            <option value="unread">Unread</option>
            <option value="to_read">To read</option>
            <option value="reading">Reading</option>
            <option value="read">Read</option>
            <option value="dropped">Dropped</option>
          </select>
        </label>

        <label>
          Publication status
          <select v-model="form.publication_status" name="publication_status">
            <option value="unknown">Unknown</option>
            <option value="finished">Finished</option>
            <option value="ongoing">Ongoing</option>
            <option value="hiatus">Hiatus</option>
          </select>
        </label>

        <label>
          Current chapter
          <input v-model="form.current_chapter" name="current_chapter" type="number" min="0" step="0.01" />
        </label>

        <label>
          Read at
          <input v-model="form.read_at" name="read_at" type="datetime-local" />
        </label>

        <label>
          Notes
          <textarea v-model="form.notes" name="notes" rows="4" />
        </label>
      </section>

      <button type="submit" :disabled="loading">
        {{ loading ? 'Importing...' : importMethod === 'epub' ? 'Import EPUB' : 'Import from metadata' }}
      </button>

      <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
      <p v-if="successMessage" role="status">{{ successMessage }}</p>
    </form>
  </main>
</template>
