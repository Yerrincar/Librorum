<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import FantasyDateTimePicker from '@/components/FantasyDateTimePicker.vue'
import FantasySelect, { type FantasySelectOption } from '@/components/FantasySelect.vue'
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

const importMethodOptions: FantasySelectOption[] = [
  { label: 'EPUB file', value: 'epub' },
  { label: 'Metadata search', value: 'metadata' },
]

const kindOptions: FantasySelectOption[] = [
  { label: 'Book', value: 'book' },
  { label: 'Manga', value: 'manga' },
  { label: 'Manhwa', value: 'manhwa' },
]

const ownershipStatusOptions: FantasySelectOption[] = [
  { label: 'None', value: 'none' },
  { label: 'Owned physical', value: 'owned_physical' },
  { label: 'Owned digital', value: 'owned_digital' },
  { label: 'Owned physical and digital', value: 'owned_physical_and_digital' },
  { label: 'Wishlist', value: 'wishlist' },
]

const readingStatusOptions: FantasySelectOption[] = [
  { label: 'Unread', value: 'unread' },
  { label: 'To read', value: 'to_read' },
  { label: 'Reading', value: 'reading' },
  { label: 'Read', value: 'read' },
  { label: 'Dropped', value: 'dropped' },
]

const publicationStatusOptions: FantasySelectOption[] = [
  { label: 'Unknown', value: 'unknown' },
  { label: 'Finished', value: 'finished' },
  { label: 'Ongoing', value: 'ongoing' },
  { label: 'Hiatus', value: 'hiatus' },
]

function showComicFields() {
  return form.kind === 'manga' || form.kind === 'manhwa'
}

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
  if (candidate.source === 'calibre') return 'Calibre'
  if (candidate.source === 'google_books') return 'Google Books'
  return 'OpenLibrary'
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
  <main class="import-page">
    <div class="import-shell">
      <template v-if="checkingUser || !user">
        <nav class="heading-nav" aria-label="Import navigation">
          <RouterLink to="/" class="heading-link">Home</RouterLink>
          <RouterLink to="/books" class="heading-link">Library</RouterLink>
        </nav>
        <h1>Add Books</h1>
      </template>

      <p v-if="checkingUser" class="status-message">Checking session…</p>

      <section v-else-if="!user" class="auth-panel">
        <p>You need to be logged in to import books.</p>
        <RouterLink to="/login">Login</RouterLink>
        <RouterLink to="/register">Register</RouterLink>
      </section>

      <form v-else @submit.prevent="submitImport">
        <div class="import-top-row">
          <div class="import-heading">
            <nav class="heading-nav" aria-label="Import navigation">
              <RouterLink to="/" class="heading-link">Home</RouterLink>
              <RouterLink to="/books" class="heading-link">Library</RouterLink>
            </nav>
            <h1>Add Books</h1>
          </div>

          <section class="method-section">
            <FantasySelect
              id="import-method"
              v-model="importMethod"
              label="Import method"
              layout="stacked"
              name="import_method"
              :options="importMethodOptions"
            />
          </section>

          <section v-if="importMethod === 'epub'" class="source-section">
            <h2>EPUB file</h2>
            <label class="file-field">
              File
              <input type="file" name="file" accept=".epub,application/epub+zip" @change="selectFile" />
              <span class="file-action">Choose EPUB file</span>
              <span>{{ selectedFile?.name ?? 'No file selected' }}</span>
            </label>
          </section>

          <section v-else class="source-section metadata-section">
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
              {{ searchingMetadata ? 'Searching…' : 'Search metadata' }}
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
              <p v-if="selectedMetadata.isbn">ISBN: {{ selectedMetadata.isbn }}</p>
              <p v-if="selectedMetadata.language">Language: {{ selectedMetadata.language }}</p>
              <p v-if="selectedMetadata.genres?.length">Genres: {{ selectedMetadata.genres.join(', ') }}</p>
              <p v-if="selectedMetadata.description">{{ selectedMetadata.description }}</p>
            </article>
          </section>
        </div>

        <section v-if="importMethod === 'epub' || selectedMetadata" class="library-fields">
          <h2>Library fields</h2>

          <div class="library-field-grid">
            <FantasySelect
              class="kind-field"
              id="book-kind"
              v-model="form.kind"
              label="Kind"
              layout="stacked"
              name="kind"
              :options="kindOptions"
            />

            <label class="rating-field">
              Rating
              <input v-model="form.rating" name="rating" type="number" min="0" max="5" step="0.1" />
            </label>

            <FantasySelect
              class="ownership-field"
              id="ownership-status"
              v-model="form.ownership_status"
              label="Ownership status"
              layout="stacked"
              name="ownership_status"
              :options="ownershipStatusOptions"
            />

            <FantasySelect
              class="reading-field"
              id="reading-status"
              v-model="form.reading_status"
              label="Reading status"
              layout="stacked"
              name="reading_status"
              :options="readingStatusOptions"
            />

            <FantasySelect
              v-if="showComicFields()"
              class="publication-field"
              id="publication-status"
              v-model="form.publication_status"
              label="Publication status"
              layout="stacked"
              name="publication_status"
              :options="publicationStatusOptions"
            />

            <label v-if="showComicFields()" class="chapter-field">
              Current chapter
              <input v-model="form.current_chapter" name="current_chapter" type="number" min="0" step="0.01" />
            </label>

            <FantasyDateTimePicker
              id="read-at"
              v-model="form.read_at"
              class="read-at-field"
              label="Read at"
              name="read_at"
            />

            <label class="notes-field">
              Notes
              <textarea v-model="form.notes" name="notes" rows="4" />
            </label>
          </div>
        </section>

        <div class="import-actions">
          <button type="submit" :disabled="loading">
            {{ loading ? 'Importing…' : importMethod === 'epub' ? 'Import EPUB' : 'Import from metadata' }}
          </button>

          <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
          <p v-if="successMessage" role="status">{{ successMessage }}</p>
        </div>
      </form>
    </div>
  </main>
</template>

<style scoped>
.import-page{
  min-height: 100vh;
  width: 100%;
  box-sizing: border-box;
  padding: 2rem;

  display: flex;
  justify-content: flex-start;
  overflow-x: hidden;

  background-image:
    linear-gradient(rgba(0, 0, 0, 0.22), rgba(0, 0, 0, 0.22)),
    url('../../images/IMPORT-BOOKS.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  background-attachment: fixed;
}

.import-shell {
  align-content: start;
  background: rgba(18, 10, 3, 0.66);
  border: 1px solid rgba(212, 175, 55, 0.62);
  box-sizing: border-box;
  display: grid;
  gap: 0.85rem;
  max-width: none;
  overflow: visible;
  padding: 1.25rem;
  width: 100%;
}

h1,
h2,
h3,
.heading-link,
form > p,
.auth-panel p,
.auth-panel a,
.status-message {
  background: linear-gradient(
    180deg,
    #ffe9a3,
    #d4af37 45%,
    #8c5a18 75%,
    #4b2805
  );
  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;
  font-family: 'Cinzel', serif;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 10px rgba(212,175,55,0.75);
}

label,
form p,
.auth-panel p,
section[aria-label='Metadata results'] span,
article p {
  color: #ffe9a3;
  font-family: 'Cinzel', serif;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 8px rgba(212,175,55,0.55);
}

h1 {
  filter: brightness(1.18);
  font-size: clamp(2.1rem, 4vw, 3rem);
  letter-spacing: 4px;
  margin: 0;
  white-space: nowrap;
}

h2,
h3 {
  letter-spacing: 2px;
  margin: 0 0 0.75rem;
}

.heading-nav {
  display: grid;
  gap: 0.2rem;
}

.heading-link {
  filter: brightness(1.16);
  font-size: 1.15rem;
  letter-spacing: 2px;
  text-decoration: none;
  width: fit-content;
}

form,
.auth-panel,
article,
section[aria-label='Metadata results'] {
  display: grid;
  gap: 0.85rem;
}

form {
  align-content: start;
  align-items: stretch;
  box-sizing: border-box;
  display: grid;
  gap: 0.55rem;
  max-width: 100%;
  width: 100%;
}

.import-top-row {
  align-items: stretch;
  display: grid;
  gap: 0.65rem;
  grid-template-columns: minmax(11rem, 0.7fr) minmax(13rem, 0.75fr) minmax(22rem, 2fr);
  width: 100%;
}

.import-heading {
  align-content: start;
  background: rgba(5, 3, 1, 0.32);
  border: 1px solid rgba(212, 175, 55, 0.36);
  box-sizing: border-box;
  display: grid;
  gap: 0.1rem;
  height: 100%;
  min-height: 8.25rem;
  min-width: 0;
  padding: 0.75rem 0.85rem;
}

.import-heading h1 {
  font-size: clamp(1.55rem, 2.8vw, 2.15rem);
  letter-spacing: 3px;
  line-height: 1.05;
}

.import-heading .heading-link {
  font-size: 1rem;
  line-height: 1.1;
}

form section {
  align-content: start;
  display: grid;
  gap: 0.8rem;
  min-width: 0;
}

.method-section {
  min-width: 0;
}

form section,
.auth-panel,
article,
section[aria-label='Metadata results'] {
  background: rgba(5, 3, 1, 0.32);
  border: 1px solid rgba(212, 175, 55, 0.36);
  box-sizing: border-box;
  max-width: 100%;
  padding: 1rem;
}

.import-top-row > section {
  gap: 0.45rem;
  min-height: 8.25rem;
  padding: 0.75rem 0.85rem;
}

.import-top-row h2 {
  font-size: 1rem;
  line-height: 1.1;
  margin-bottom: 0.25rem;
}

.import-top-row label {
  font-size: 0.86rem;
  gap: 0.25rem;
}

.method-section :deep(.fantasy-select) {
  gap: 0.3rem;
}

.method-section :deep(.fantasy-select-label) {
  font-size: 0.82rem;
  letter-spacing: 1.5px;
}

.method-section :deep(.fantasy-select-button) {
  padding-block: 0.25rem;
}

.source-section .file-field {
  gap: 0.25rem;
}

.source-section {
  min-height: 0;
}

.metadata-section {
  grid-template-columns: repeat(2, minmax(0, 1fr)) auto;
}

.metadata-section h2,
.metadata-section section,
.metadata-section article {
  grid-column: 1 / -1;
}

.metadata-section > button {
  align-self: end;
}

.library-fields {
  box-sizing: border-box;
  min-height: 17rem;
  width: 100%;
}

.library-field-grid {
  display: grid;
  gap: 0.85rem;
  grid-template-areas:
    'kind rating notes notes'
    'ownership reading notes notes'
    'publication read-at notes notes'
    'chapter . notes notes';
  grid-template-columns: minmax(10rem, 0.95fr) minmax(10rem, 0.95fr) minmax(14rem, 1.25fr) minmax(14rem, 1.25fr);
  min-width: 0;
}

.kind-field {
  grid-area: kind;
}

.rating-field {
  grid-area: rating;
}

.ownership-field {
  grid-area: ownership;
}

.reading-field {
  grid-area: reading;
}

.publication-field {
  grid-area: publication;
}

.chapter-field {
  grid-area: chapter;
}

.read-at-field {
  grid-area: read-at;
}

.notes-field {
  align-self: stretch;
  gap: 0.12rem;
  grid-area: notes;
  grid-template-rows: auto minmax(0, 1fr);
  min-width: 0;
}

.notes-field textarea {
  align-self: start;
  height: 100%;
  min-height: 0;
  resize: vertical;
}

label {
  display: grid;
  gap: 0.35rem;
  letter-spacing: 1px;
}

input,
textarea {
  background: rgba(20, 11, 1, 0.82);
  border: 1px solid rgba(212, 175, 55, 0.5);
  box-sizing: border-box;
  color: #ffe9a3;
  font-family: 'Cinzel', serif;
  max-width: 100%;
  padding: 0.45rem 0.55rem;
  width: 100%;
}

input:focus-visible,
select:focus-visible,
textarea:focus-visible,
button:focus-visible,
a:focus-visible {
  outline: 2px solid rgba(212, 175, 55, 0.85);
  outline-offset: 0.25rem;
}

button {
  background: transparent;
  border: 1px solid rgba(212, 175, 55, 0.55);
  color: #ffe9a3;
  cursor: pointer;
  font-size: 1rem;
  font-family: 'Cinzel', serif;
  letter-spacing: 1px;
  padding: 0.5rem 0.75rem;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 8px rgba(212,175,55,0.55);
  width: fit-content;
}

.import-actions {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 0.85rem 1rem;
}

.import-actions p {
  margin: 0;
}

button:disabled {
  cursor: wait;
  opacity: 0.65;
}

.file-field input[type='file'] {
  height: 1px;
  opacity: 0;
  position: absolute;
  width: 1px;
}

.file-action {
  background: transparent;
  border: 1px solid rgba(212, 175, 55, 0.55);
  color: #ffe9a3;
  cursor: pointer;
  font-family: 'Cinzel', serif;
  padding: 0.3rem 0.55rem;
  width: fit-content;
}

.file-field input[type='file']:focus-visible + .file-action {
  outline: 2px solid rgba(212, 175, 55, 0.85);
  outline-offset: 0.25rem;
}

.file-field span {
  color: #ffe9a3;
  font-family: 'Cinzel', serif;
  font-size: 0.82rem;
  letter-spacing: 1px;
}

@media (max-width: 1050px) {
  .import-top-row {
    grid-template-columns: 1fr 1fr;
  }

  .source-section {
    grid-column: 1 / -1;
  }

  .library-field-grid {
    grid-template-areas:
      'kind notes'
      'ownership notes'
      'rating notes'
      'reading notes'
      'read-at notes'
      'publication chapter';
    grid-template-columns: minmax(10rem, 1fr) minmax(14rem, 1.35fr);
  }
}

@media (max-width: 680px) {
  .import-page {
    padding: 1rem;
  }

  .import-shell {
    padding: 0.85rem;
  }

  .import-top-row,
  .metadata-section,
  .library-field-grid {
    grid-template-columns: 1fr;
  }

  .library-field-grid {
    grid-template-areas:
      'kind'
      'rating'
      'ownership'
      'reading'
      'publication'
      'chapter'
      'read-at'
      'notes';
  }

  .notes-field {
    grid-column: auto;
  }
}
</style>
