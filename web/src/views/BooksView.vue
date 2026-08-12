<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import FantasySelect from '@/components/FantasySelect.vue'
import {
  bookAuthor,
  bookCoverPath,
  bookCurrentChapter,
  bookDescription,
  bookGenres,
  bookLanguage,
  bookNotes,
  bookOwnershipStatus,
  bookReadAtInput,
  bookRating,
  bookReadMonthYear,
  bookReadingStatus,
  bookTitle,
  bookTotalChapters,
  fetchBooks,
  fetchReadingYears,
  rawBookCoverPath,
  type LibraryItemKind,
  type LibraryItemResponse,
  updateLibraryItem,
} from '@/api/books'

type EditBookForm = {
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

const books = ref<LibraryItemResponse[]>([])
const loading = ref(true)
const errorMessage = ref('')
const readingYearsErrorMessage = ref('')
const activeKind = ref<LibraryItemKind | undefined>()
const activeView = ref<'covers' | 'grid'>('covers')
const selectedBook = ref<LibraryItemResponse>()
const editingBook = ref(false)
const savingBook = ref(false)
const updateErrorMessage = ref('')
const editForm = ref<EditBookForm>(emptyEditBookForm())
const activePage = ref(1)
const hasNextPage = ref(false)
const readingYearFilter = ref('')
const readingYears = ref<number[]>([])
const booksPerPage = 12

const sections: Array<{ label: string; kind?: LibraryItemKind }> = [
  { label: 'Library' },
  { label: 'Books', kind: 'book' },
  { label: 'Manga', kind: 'manga' },
  { label: 'Manhwa', kind: 'manhwa' },
]

function readingYearOptions() {
  return [
    { label: 'All', value: '' },
    ...readingYears.value.map((year) => ({ label: String(year), value: String(year) })),
  ]
}

function activeSectionLabel() {
  return sections.find((section) => section.kind === activeKind.value)?.label ?? 'Library'
}

function activeSectionIsWorkInProgress() {
  return activeKind.value === 'manga' || activeKind.value === 'manhwa'
}

async function hasBooksAfterPage(kind: LibraryItemKind | undefined, page: number, currentBooks: LibraryItemResponse[]) {
  if (currentBooks.length < booksPerPage) {
    return false
  }

  const nextBooks = await fetchBooks({ kind, page: page + 1, limit: booksPerPage, readingYear: readingYearFilter.value })
  return nextBooks.length > 0
}

onMounted(() => {
  void loadSection()
  window.addEventListener('keydown', handleBookDetailsKeydown)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleBookDetailsKeydown)
})

async function loadBooks(kind?: LibraryItemKind, page = 1) {
  activeKind.value = kind
  activePage.value = page
  selectedBook.value = undefined
  loading.value = true
  errorMessage.value = ''

  try {
    const loadedBooks = await fetchBooks({ kind, page, limit: booksPerPage, readingYear: readingYearFilter.value })
    books.value = loadedBooks
    hasNextPage.value = await hasBooksAfterPage(kind, page, loadedBooks)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Fetch books failed'
  } finally {
    loading.value = false
  }
}

async function loadReadingYears(kind?: LibraryItemKind) {
  readingYearsErrorMessage.value = ''
  readingYears.value = await fetchReadingYears(kind)
  if (readingYearFilter.value && !readingYears.value.includes(Number(readingYearFilter.value))) {
    readingYearFilter.value = ''
  }
}

async function loadSection(kind?: LibraryItemKind) {
  try {
    await loadReadingYears(kind)
  } catch (error) {
    readingYears.value = []
    readingYearsErrorMessage.value = error instanceof Error ? error.message : 'Fetch reading years failed'
  }
  await loadBooks(kind, 1)
}

function applyReadingYearFilter() {
  void loadBooks(activeKind.value, 1)
}

function openBookDetails(book: LibraryItemResponse) {
  selectedBook.value = book
  cancelBookEdit()
}

function closeBookDetails() {
  selectedBook.value = undefined
  cancelBookEdit()
}

function handleBookDetailsKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape' && selectedBook.value) {
    closeBookDetails()
  }
}

function startBookEdit() {
  if (!selectedBook.value) {
    return
  }

  editForm.value = editBookForm(selectedBook.value)
  updateErrorMessage.value = ''
  editingBook.value = true
}

function cancelBookEdit() {
  editingBook.value = false
  savingBook.value = false
  updateErrorMessage.value = ''
  editForm.value = emptyEditBookForm()
}

function onCoverFileChange(event: Event) {
  const input = event.target as HTMLInputElement
  editForm.value.cover_file = input.files?.[0]
}

async function confirmBookEdit() {
  const bookId = selectedBook.value?.id
  if (!bookId) {
    updateErrorMessage.value = 'Book id is missing'
    return
  }

  savingBook.value = true
  updateErrorMessage.value = ''

  try {
    await updateLibraryItem({ id: bookId, ...editForm.value })
    await loadReadingYears(activeKind.value)
    const refreshedBooks = await fetchBooks({
      kind: activeKind.value,
      page: activePage.value,
      limit: booksPerPage,
      readingYear: readingYearFilter.value,
    })
    books.value = refreshedBooks
    hasNextPage.value = await hasBooksAfterPage(activeKind.value, activePage.value, refreshedBooks)
    selectedBook.value = refreshedBooks.find((book) => book.id === bookId)
    editingBook.value = false
  } catch (error) {
    updateErrorMessage.value = error instanceof Error ? error.message : 'Book update failed'
  } finally {
    savingBook.value = false
  }
}

function emptyEditBookForm(): EditBookForm {
  return {
    title: '',
    author: '',
    rating: '',
    cover_path: '',
    read_at: '',
    description: '',
    language: '',
    genres: '',
    ownership_status: 'none',
    reading_status: 'unread',
    current_chapter: '',
    total_chapters: '',
    notes: '',
  }
}

function editBookForm(book: LibraryItemResponse): EditBookForm {
  return {
    title: bookTitle(book),
    author: bookAuthor(book),
    rating: bookRating(book),
    cover_path: rawBookCoverPath(book),
    read_at: bookReadAtInput(book),
    description: bookDescription(book),
    language: bookLanguage(book),
    genres: bookGenres(book),
    ownership_status: bookOwnershipStatus(book) || 'none',
    reading_status: bookReadingStatus(book),
    current_chapter: bookCurrentChapter(book),
    total_chapters: bookTotalChapters(book),
    notes: bookNotes(book),
  }
}
</script>

<template>
  <main class="books-page">

    <!-- Left Sidebar with Home Link & Section Options -->
    <aside class="sidebar">
      <nav aria-label="Library sections" class="sections-nav">
        <RouterLink to="/" class="nav-option home-link">
          Home
        </RouterLink>

        <button
          v-for="section in sections"
          :key="section.label"
          type="button"
          class="nav-option"
          :class="{ active: activeKind === section.kind }"
          :disabled="activeKind === section.kind"
          @click="loadSection(section.kind)"
        >
          {{ section.label }}
        </button>
      </nav>
    </aside>

    <!-- Main Content Area -->
    <div class="content-area">

      <!-- Header with View Switcher (Positioned on top of books) -->
      <header class="content-header">
        <div class="title-tools">
          <h1 class="page-title">{{ activeSectionLabel() }}</h1>
          <FantasySelect
            id="reading-year-filter"
            v-model="readingYearFilter"
            label="Read year"
            name="reading_year"
            :options="readingYearOptions()"
            @change="applyReadingYearFilter"
          />
        </div>

        <nav aria-label="Library view" class="view-nav">
          <button
            type="button"
            class="view-option"
            :class="{ active: activeView === 'covers' }"
            :disabled="activeView === 'covers'"
            @click="activeView = 'covers'"
          >
            Covers
          </button>
          <button
            type="button"
            class="view-option"
            :class="{ active: activeView === 'grid' }"
            :disabled="activeView === 'grid'"
            @click="activeView = 'grid'"
          >
            Grid
          </button>
        </nav>
      </header>

      <!-- Status messages -->
      <p v-if="loading" class="status-message">Loading books…</p>
      <p v-else-if="errorMessage" role="alert" class="status-message error">{{ errorMessage }}</p>
      <p v-if="readingYearsErrorMessage" role="alert" class="status-message error">{{ readingYearsErrorMessage }}</p>

      <section v-else-if="activeSectionIsWorkInProgress()" class="work-section" aria-label="Work in progress">
        <p>Work in progress</p>
        <span>{{ activeSectionLabel() }} will be added in a future version.</span>
      </section>

      <section v-else-if="books.length === 0" class="empty-section">
        <p class="status-message">No books found.</p>
        <RouterLink to="/books/import" class="import-link">Import an EPUB</RouterLink>
      </section>

      <!-- Covers View -->
      <section v-else-if="activeView === 'covers'" aria-label="Book covers">
        <article v-for="book in books" :key="book.id ?? bookTitle(book)">
          <button
            type="button"
            class="cover-button"
            :aria-label="`Show details for ${bookTitle(book)}`"
            @click="openBookDetails(book)"
          >
            <span class="cover-frame">
              <img v-if="bookCoverPath(book)" :src="bookCoverPath(book)" :alt="bookTitle(book)" width="160" height="240" loading="lazy" />
              <span v-else class="no-cover-text">No cover</span>
            </span>
          </button>
        </article>
      </section>

      <!-- Grid View -->
      <section v-else aria-label="Book grid" class="grid-table" role="table">
        <div class="grid-header" role="row">
          <span role="columnheader">Title</span>
          <span role="columnheader">Author</span>
          <span role="columnheader">Rating</span>
          <span role="columnheader">Read at</span>
        </div>
        <div v-for="book in books" :key="book.id ?? bookTitle(book)" class="grid-row" role="row">
          <span role="cell">{{ bookTitle(book) }}</span>
          <span role="cell">{{ bookAuthor(book) }}</span>
          <span role="cell">{{ bookRating(book) ? `${bookRating(book)}/5` : '' }}</span>
          <span role="cell">{{ bookReadMonthYear(book) }}</span>
        </div>
      </section>

      <nav v-if="!loading && !errorMessage && books.length > 0" aria-label="Library pages" class="page-nav">
        <button
          type="button"
          class="page-option"
          :disabled="activePage <= 1"
          @click="loadBooks(activeKind, activePage - 1)"
        >
          Previous
        </button>
        <span class="page-indicator">Page {{ activePage }}</span>
        <button
          type="button"
          class="page-option"
          :disabled="!hasNextPage"
          @click="loadBooks(activeKind, activePage + 1)"
        >
          Next
        </button>
      </nav>

    </div>

    <!-- Details Modal -->
    <div v-if="selectedBook" class="details-backdrop" @click.self="closeBookDetails">
      <section class="details-panel" role="dialog" aria-modal="true" aria-labelledby="book-details-title">
        <div class="details-actions">
          <template v-if="editingBook">
            <button type="button" :disabled="savingBook" @click="confirmBookEdit">{{ savingBook ? 'Saving…' : 'Confirm' }}</button>
            <button type="button" :disabled="savingBook" @click="cancelBookEdit">Cancel</button>
          </template>
          <template v-else>
            <button type="button" @click="startBookEdit">Edit</button>
            <button type="button" aria-label="Close details" @click="closeBookDetails">Close</button>
          </template>
        </div>
        <div class="details-cover">
          <div class="cover-frame">
            <img
              v-if="bookCoverPath(selectedBook)"
              :src="bookCoverPath(selectedBook)"
              :alt="bookTitle(selectedBook)"
              width="160"
              height="240"
              loading="lazy"
            />
            <span v-else>No cover</span>
          </div>
          <label v-if="editingBook" class="edit-field">
            Cover
            <input name="file" type="file" accept="image/jpeg,image/png" @change="onCoverFileChange" />
          </label>
        </div>
        <div class="details-body">
          <form v-if="editingBook" class="edit-form" @submit.prevent="confirmBookEdit">
            <h2 id="book-details-title">Edit book</h2>
            <p v-if="updateErrorMessage" role="alert">{{ updateErrorMessage }}</p>
            <label class="edit-field">
              Title
              <input v-model="editForm.title" name="title" required />
            </label>
            <label class="edit-field">
              Author
              <input v-model="editForm.author" name="author" />
            </label>
            <label class="edit-field">
              Rating
              <input v-model="editForm.rating" name="rating" type="number" min="0" max="5" step="0.1" />
            </label>
            <label class="edit-field">
              Description
              <textarea v-model="editForm.description" name="description" rows="5" />
            </label>
            <label class="edit-field">
              Genres
              <input v-model="editForm.genres" name="genres" placeholder="Fantasy, Sci-Fi…" autocomplete="off" />
            </label>
            <label class="edit-field">
              Ownership status
              <select v-model="editForm.ownership_status" name="ownership_status">
                <option value="none">None</option>
                <option value="owned_physical">Owned physical</option>
                <option value="owned_digital">Owned digital</option>
                <option value="owned_physical_and_digital">Owned physical and digital</option>
                <option value="wishlist">Wishlist</option>
              </select>
            </label>
            <label class="edit-field">
              Read at
              <input v-model="editForm.read_at" name="read_at" type="datetime-local" />
            </label>
          </form>
          <dl v-else class="details-fields">
            <div>
              <dt>Title</dt>
              <dd id="book-details-title">{{ bookTitle(selectedBook) }}</dd>
            </div>
            <div>
              <dt>Author</dt>
              <dd>{{ bookAuthor(selectedBook) || 'Unknown' }}</dd>
            </div>
            <div>
              <dt>Rating</dt>
              <dd>{{ bookRating(selectedBook) ? `${bookRating(selectedBook)}/5` : 'Unrated' }}</dd>
            </div>
            <div>
              <dt>Read at</dt>
              <dd>{{ bookReadMonthYear(selectedBook) || bookReadingStatus(selectedBook) }}</dd>
            </div>
            <div>
              <dt>Ownership status</dt>
              <dd>{{ bookOwnershipStatus(selectedBook) || 'none' }}</dd>
            </div>
            <div>
              <dt>Genres</dt>
              <dd>{{ bookGenres(selectedBook) || 'No genres' }}</dd>
            </div>
            <div>
              <dt>Description</dt>
              <dd>
                <details class="description-box">
                  <summary>Open description</summary>
                  <p>{{ bookDescription(selectedBook) || 'No description' }}</p>
                </details>
              </dd>
            </div>
          </dl>
        </div>
      </section>
    </div>

  </main>
</template>

<style scoped>
/* Main Page Container (Edge-to-edge background) */
.books-page {
  min-height: 100vh;
  width: 100%;
  box-sizing: border-box;
  padding: 3rem 2rem;

  display: grid;
  gap: 2rem;
  grid-template-columns: 14rem 1fr;

  background-image:
    linear-gradient(rgba(0, 0, 0, 0.20), rgba(0, 0, 0, 0.20)),
    url('../../images/BOOKS-VIEW.jpeg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;
  background-attachment: fixed;
}

/* -----------------------------
   REUSABLE GOLD METALLIC TEXT
----------------------------- */
.nav-option,
.view-option,
.page-option,
.page-indicator,
.page-title,
.grid-header span,
.grid-row span,
.status-message,
.import-link {
  font-family: 'Cinzel', serif;
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

  text-shadow:
    2px 3px 5px rgba(0,0,0,0.9),
    0 0 10px rgba(212,175,55,0.35);

  filter: brightness(1.1);
  transition: filter 0.3s ease, transform 0.3s ease;
}

/* -----------------------------
   LEFT SIDEBAR (OPTIONS)
----------------------------- */
.sidebar {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
  padding-top: 0.4rem;
}

.sections-nav {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1.25rem;
}

.home-link {
  font-weight: bold;
}

.nav-option {
  background: linear-gradient(
    180deg,
    #ffe9a3,
    #d4af37 45%,
    #8c5a18 75%,
    #4b2805
  );
  -webkit-background-clip: text;
  background-clip: text;
  border: none;
  color: transparent;
  cursor: pointer;
  display: inline-block;
  filter: brightness(1.4);
  font-family: 'Cinzel', serif;
  font-size: 2rem;
  font-weight: 400;
  letter-spacing: 3px;
  padding: 0;
  text-decoration: none;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 10px rgba(212,175,55,0.75);
  text-align: left;
  transition: filter 0.3s ease, transform 0.3s ease;
}

.nav-option:hover {
  transform: translateY(-3px);
  filter: brightness(1.3);
}

.nav-option:focus-visible {
  outline: none;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 14px rgba(212,175,55,0.95);
}

.nav-option.active,
.nav-option:disabled,
.view-option.active,
.view-option:disabled,
.page-option:disabled {
  filter: brightness(1.4);
  cursor: default;
  transform: none;
}

/* -----------------------------
   CONTENT AREA & HEADER
----------------------------- */
.content-area {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  padding: 0 70px 0 0;
}

.content-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid rgba(212, 175, 55, 0.3);
  padding-bottom: 1rem;
}

.title-tools {
  align-items: center;
  display: flex;
  gap: 1.5rem;
}

.page-title {
  margin: 0;
  font-size: 3rem;
  letter-spacing: 4px;
}

.view-nav {
  display: flex;
  gap: 1.5rem;
}

.page-nav {
  align-items: center;
  display: flex;
  gap: 1.5rem;
  justify-content: center;
}

.page-indicator {
  font-size: 1.25rem;
  letter-spacing: 2px;
}

.view-option,
.page-option {
  background: linear-gradient(
    180deg,
    #ffe9a3,
    #d4af37 45%,
    #8c5a18 75%,
    #4b2805
  );
  -webkit-background-clip: text;
  background-clip: text;
  border: none;
  color: transparent;
  cursor: pointer;
  display: inline-block;
  filter: brightness(1.4);
  font-family: 'Cinzel', serif;
  font-size: 1.25rem;
  letter-spacing: 2px;
  padding: 0.2rem 0.5rem;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 10px rgba(212,175,55,0.75);
  transition: filter 0.3s ease, transform 0.3s ease;
}

.view-option:hover {
  transform: translateY(-3px);
  filter: brightness(1.3);
}

.page-option:hover {
  transform: translateY(-3px);
  filter: brightness(1.3);
}

.view-option:focus-visible,
.page-option:focus-visible {
  outline: none;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 14px rgba(212,175,55,0.95);
}

/* -----------------------------
   COVERS VIEW
----------------------------- */
section[aria-label='Book covers'] {
  display: grid;
  gap: 2.25rem;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  align-items: start;
}

article {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.cover-button {
  background: transparent;
  border: 0;
  cursor: pointer;
  padding: 0;
  width: 100%;

  transition: transform 0.3s ease, filter 0.3s ease;
}

.cover-button:hover {
  transform: translateY(-6px);
  filter: drop-shadow(0 8px 15px rgba(0, 0, 0, 0.7));
}

.cover-button:focus-visible {
  outline: 2px solid rgba(212, 175, 55, 0.85);
  outline-offset: 0.35rem;
}

.cover-frame {
  aspect-ratio: 2 / 3;
  background: rgba(0, 0, 0, 0.3);
  border: 1px solid rgba(212, 175, 55, 0.4);
  border-radius: 4px;
  width: 100%;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cover-frame img {
  display: block;
  height: 100%;
  width: 100%;
  object-fit: cover;
}

.no-cover-text {
  font-size: 0.9rem;
}

/* -----------------------------
   GRID VIEW (METALLIC FONT TABLE)
----------------------------- */
.grid-table {
  display: flex;
  flex-direction: column;
  width: 100%;
  overflow-x: auto;
}

.grid-header,
.grid-row {
  display: grid;
  grid-template-columns: minmax(12rem, 2fr) minmax(10rem, 1.5fr) 6rem 8rem;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0.5rem;
  border-bottom: 1px solid rgba(212, 175, 55, 0.2);
}

.grid-header {
  border-bottom: 2px solid rgba(212, 175, 55, 0.5);
}

.grid-header span {
  font-size: 1.35rem;
  font-weight: bold;
  letter-spacing: 2px;
}

.grid-row span {
  font-size: 1.15rem;
  letter-spacing: 1px;
}

/* -----------------------------
   MESSAGES & DETAILS MODAL
----------------------------- */
.status-message {
  font-size: 1.2rem;
  letter-spacing: 1px;
}

.status-message.error {
  color: #ff6b6b;
  background: none;
  -webkit-background-clip: unset;
}

.work-section {
  background: rgba(18, 10, 3, 0.62);
  border: 1px solid rgba(212, 175, 55, 0.62);
  display: grid;
  gap: 0.75rem;
  justify-items: start;
  padding: 1.5rem;
}

.work-section p,
.work-section span {
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
  margin: 0;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 10px rgba(212,175,55,0.75);
}

.work-section p {
  filter: brightness(1.45);
  font-size: 2rem;
  letter-spacing: 3px;
}

.work-section span {
  filter: brightness(1.2);
  font-size: 1.1rem;
  letter-spacing: 1px;
}

.details-backdrop {
  position: fixed;
  inset: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(7px);
  background: rgba(7, 5, 2, 0.28);
  padding: 1.5rem;
  box-sizing: border-box;
  overscroll-behavior: contain;
}

.details-panel {
  position: relative;
  width: 100%;
  max-width: 50rem;
  max-height: calc(100vh - 3rem);

  display: grid;
  grid-template-columns: 12rem minmax(0, 1fr);
  gap: 2rem;

  background: rgba(18, 10, 3, 0.88);
  border: 1px solid rgba(212, 175, 55, 0.72);
  color: #ffe9a3;
  padding: 2rem;
  border-radius: 0;
  box-shadow: 0 1rem 3rem rgba(0, 0, 0, 0.6);
  overflow-y: auto;
}

.details-actions {
  position: absolute;
  top: 1rem;
  right: 1rem;
  display: flex;
  gap: 0.5rem;
}

.details-actions button {
  background: linear-gradient(
    180deg,
    #ffe9a3,
    #d4af37 45%,
    #8c5a18 75%,
    #4b2805
  );
  -webkit-background-clip: text;
  background-clip: text;
  border: 1px solid rgba(212, 175, 55, 0.55);
  border-radius: 0;
  color: transparent;
  cursor: pointer;
  filter: brightness(1.35);
  font-family: 'Cinzel', serif;
  letter-spacing: 1px;
  padding: 0.35rem 0.6rem;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 10px rgba(212,175,55,0.75);
}

.details-body {
  display: flex;
  flex-direction: column;
  gap: 1rem;
  padding-top: 1rem;
}

.edit-form,
.edit-field {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.edit-form h2,
.edit-form p {
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
  letter-spacing: 2px;
  margin: 0;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 10px rgba(212,175,55,0.75);
}

.edit-field {
  color: #cfa34a;
  font-family: 'Cinzel', serif;
  letter-spacing: 2px;
  text-shadow: 1px 2px 4px rgba(0,0,0,0.9);
  text-transform: uppercase;
}

.edit-field input,
.edit-field select,
.edit-field textarea {
  background: rgba(20, 11, 1, 0.82);
  border: 1px solid rgba(212, 175, 55, 0.5);
  border-radius: 0;
  box-sizing: border-box;
  color: #ffe9a3;
  font-family: 'Cinzel', serif;
  letter-spacing: 1px;
  padding: 0.45rem 0.55rem;
  text-shadow: 1px 2px 4px rgba(0,0,0,0.9);
  width: 100%;
}

.edit-field input::placeholder,
.edit-field textarea::placeholder {
  color: rgba(255, 233, 163, 0.62);
}

.edit-field select option {
  background: #140b01;
  color: #ffe9a3;
}

.edit-field input[type='file'] {
  color: #ffe9a3;
}

.edit-field input[type='file']::file-selector-button {
  background: transparent;
  border: 1px solid rgba(212, 175, 55, 0.55);
  color: #ffe9a3;
  cursor: pointer;
  font-family: 'Cinzel', serif;
  letter-spacing: 1px;
  margin-right: 0.75rem;
  padding: 0.3rem 0.55rem;
}

.edit-field input:focus-visible,
.edit-field select:focus-visible,
.edit-field textarea:focus-visible {
  outline: 2px solid rgba(212, 175, 55, 0.85);
  outline-offset: 0.25rem;
}

.details-fields {
  display: grid;
  gap: 0.8rem;
  margin: 0;
}

.details-fields div {
  display: grid;
  gap: 0.2rem;
}

.description-box summary,
.description-box p {
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

.details-fields dt {
  color: #cfa34a;
  filter: none;
  font-family: 'Cinzel', serif;
  font-size: 1rem;
  letter-spacing: 2.5px;
  text-shadow: 1px 2px 4px rgba(0,0,0,0.9);
  text-transform: uppercase;
}

.details-fields dd {
  color: #ffe9a3;
  filter: brightness(1.2);
  font-size: 1.25rem;
  font-family: 'Cinzel', serif;
  letter-spacing: 1px;
  margin: 0;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 10px rgba(212,175,55,0.75);
}

.description-box {
  background: rgba(5, 3, 1, 0.38);
  border: 1px solid rgba(212, 175, 55, 0.45);
  padding: 0.65rem 0.75rem;
}

.description-box summary {
  cursor: pointer;
  filter: brightness(1.35);
  font-size: 1.1rem;
}

.description-box p {
  background: none;
  -webkit-background-clip: unset;
  background-clip: unset;
  color: #ffe9a3;
  filter: brightness(1.15);
  font-size: 1.12rem;
  line-height: 1.55;
  margin: 0.75rem 0 0;
}

/* Responsive breakpoint for mobile */
@media (max-width: 52rem) {
  .books-page {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }

  .sidebar {
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: space-between;
    align-items: center;
  }

  .sections-nav {
    flex-direction: row;
    flex-wrap: wrap;
  }

  section[aria-label='Book covers'] {
    grid-template-columns: repeat(auto-fill, minmax(9rem, 1fr));
  }

  .details-panel {
    grid-template-columns: 1fr;
  }
}
</style>
