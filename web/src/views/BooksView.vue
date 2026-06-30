<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
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
const activeKind = ref<LibraryItemKind | undefined>()
const activeView = ref<'covers' | 'grid'>('covers')
const selectedBook = ref<LibraryItemResponse>()
const editingBook = ref(false)
const savingBook = ref(false)
const updateErrorMessage = ref('')
const editForm = ref<EditBookForm>(emptyEditBookForm())

const sections: Array<{ label: string; kind?: LibraryItemKind }> = [
  { label: 'Library' },
  { label: 'Books', kind: 'book' },
  { label: 'Manga', kind: 'manga' },
  { label: 'Manhwa', kind: 'manhwa' },
]

onMounted(() => {
  void loadBooks()
})

async function loadBooks(kind?: LibraryItemKind) {
  activeKind.value = kind
  selectedBook.value = undefined
  loading.value = true
  errorMessage.value = ''

  try {
    books.value = await fetchBooks(kind)
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Fetch books failed'
  } finally {
    loading.value = false
  }
}

function openBookDetails(book: LibraryItemResponse) {
  selectedBook.value = book
  cancelBookEdit()
}

function closeBookDetails() {
  selectedBook.value = undefined
  cancelBookEdit()
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
    const refreshedBooks = await fetchBooks(activeKind.value)
    books.value = refreshedBooks
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
  <main>
    <h1>Library</h1>

    <nav aria-label="Library sections">
      <button
        v-for="section in sections"
        :key="section.label"
        type="button"
        :disabled="activeKind === section.kind"
        @click="loadBooks(section.kind)"
      >
        {{ section.label }}
      </button>
    </nav>

    <nav aria-label="Library view">
      <button type="button" :disabled="activeView === 'covers'" @click="activeView = 'covers'">Covers</button>
      <button type="button" :disabled="activeView === 'grid'" @click="activeView = 'grid'">Grid</button>
    </nav>

    <p v-if="loading">Loading books...</p>
    <p v-else-if="errorMessage" role="alert">{{ errorMessage }}</p>

    <section v-else-if="books.length === 0">
      <p>No books found.</p>
      <RouterLink to="/books/import">Import an EPUB</RouterLink>
    </section>

    <section v-else-if="activeView === 'covers'" aria-label="Book covers">
      <article v-for="book in books" :key="book.id ?? bookTitle(book)">
        <button type="button" class="cover-button" :aria-label="`Show details for ${bookTitle(book)}`" @click="openBookDetails(book)">
          <span class="cover-frame">
            <img v-if="bookCoverPath(book)" :src="bookCoverPath(book)" :alt="bookTitle(book)" loading="lazy" />
            <span v-else>No cover</span>
          </span>
        </button>
      </article>
    </section>

    <section v-else aria-label="Book grid">
      <div class="grid-header" role="row">
        <span>Title</span>
        <span>Author</span>
        <span>Rating</span>
        <span>Read at</span>
      </div>
      <div v-for="book in books" :key="book.id ?? bookTitle(book)" class="grid-row" role="row">
        <span>{{ bookTitle(book) }}</span>
        <span>{{ bookAuthor(book) }}</span>
        <span>{{ bookRating(book) }}</span>
        <span>{{ bookReadMonthYear(book) }}</span>
      </div>
    </section>

    <div v-if="selectedBook" class="details-backdrop" @click.self="closeBookDetails">
      <section class="details-panel" role="dialog" aria-modal="true" aria-labelledby="book-details-title">
        <div class="details-actions">
          <template v-if="editingBook">
            <button type="button" :disabled="savingBook" @click="confirmBookEdit">{{ savingBook ? 'Saving...' : 'Confirm' }}</button>
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
              loading="lazy"
            />
            <span v-else>No cover</span>
          </div>
          <label v-if="editingBook" class="edit-field">
            Cover
            <input type="file" accept="image/jpeg,image/png" @change="onCoverFileChange" />
          </label>
        </div>
        <div class="details-body">
          <h2 v-if="!editingBook" id="book-details-title">{{ bookTitle(selectedBook) }}</h2>
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
              <input v-model="editForm.genres" name="genres" placeholder="Fantasy, Sci-Fi" />
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
          <dl v-else>
            <div v-if="bookRating(selectedBook)">
              <dt>Rating</dt>
              <dd>{{ bookRating(selectedBook) }}</dd>
            </div>
            <div v-if="bookDescription(selectedBook)">
              <dt>Description</dt>
              <dd class="description">{{ bookDescription(selectedBook) }}</dd>
            </div>
            <div v-if="bookGenres(selectedBook)">
              <dt>Genres</dt>
              <dd>{{ bookGenres(selectedBook) }}</dd>
            </div>
            <div v-if="bookAuthor(selectedBook)">
              <dt>Author</dt>
              <dd>{{ bookAuthor(selectedBook) }}</dd>
            </div>
            <div v-if="bookOwnershipStatus(selectedBook)">
              <dt>Ownership status</dt>
              <dd>{{ bookOwnershipStatus(selectedBook) }}</dd>
            </div>
            <div v-if="bookReadMonthYear(selectedBook)">
              <dt>Read at</dt>
              <dd>{{ bookReadMonthYear(selectedBook) }}</dd>
            </div>
          </dl>
        </div>
      </section>
    </div>
  </main>
</template>

<style scoped>
main {
  display: grid;
  gap: 1rem;
}

nav {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

section[aria-label='Book covers'] {
  display: grid;
  gap: 1.25rem;
  grid-template-columns: repeat(auto-fill, minmax(9rem, 1fr));
}

article {
  display: grid;
  justify-items: center;
  text-align: center;
}

.cover-button {
  background: transparent;
  border: 0;
  cursor: pointer;
  display: block;
  padding: 0;
  width: 100%;
}

.cover-button:focus-visible {
  outline: 3px solid #335cff;
  outline-offset: 4px;
}

.cover-frame {
  aspect-ratio: 2 / 3;
  background: #eee;
  width: 100%;
  max-width: 10rem;
  overflow: hidden;
}

.cover-frame img {
  display: block;
  height: 100%;
  object-fit: cover;
  width: 100%;
}

.cover-frame span {
  align-items: center;
  display: flex;
  height: 100%;
  justify-content: center;
  padding: 0.75rem;
}

section[aria-label='Book grid'] {
  display: grid;
  overflow-x: auto;
}

.grid-header,
.grid-row {
  align-items: center;
  border-bottom: 1px solid #ddd;
  display: grid;
  gap: 1rem;
  grid-template-columns: minmax(12rem, 2fr) minmax(10rem, 1.5fr) 5rem 6rem;
  min-width: 34rem;
  padding: 0.55rem 0;
}

.grid-header {
  font-weight: 700;
}

.grid-row span:nth-child(3),
.grid-row span:nth-child(4) {
  color: #555;
}

.details-backdrop {
  align-items: center;
  background: rgb(0 0 0 / 0.45);
  display: grid;
  inset: 0;
  padding: 1rem;
  position: fixed;
  z-index: 10;
}

.details-panel {
  background: #fff;
  box-shadow: 0 1rem 3rem rgb(0 0 0 / 0.3);
  display: grid;
  gap: 1.5rem;
  grid-template-columns: minmax(10rem, 14rem) minmax(0, 1fr);
  max-height: calc(100vh - 2rem);
  max-width: 52rem;
  overflow: auto;
  padding: 1.25rem;
  position: relative;
  width: min(100%, 52rem);
}

.details-actions {
  display: flex;
  gap: 0.5rem;
  position: absolute;
  right: 0.75rem;
  top: 0.75rem;
}

.details-cover .cover-frame {
  max-width: none;
}

.details-body {
  display: grid;
  gap: 0.75rem;
  padding-right: 3rem;
}

.details-body dl,
.details-body dd {
  margin: 0;
}

.details-body dl {
  display: grid;
  gap: 0.75rem;
}

.details-body dt {
  font-weight: 700;
}

.edit-form {
  display: grid;
  gap: 0.75rem;
}

.edit-field {
  display: grid;
  gap: 0.25rem;
}

.edit-field input,
.edit-field select,
.edit-field textarea {
  font: inherit;
  max-width: 100%;
  padding: 0.4rem;
}

.description {
  line-height: 1.5;
}

@media (max-width: 42rem) {
  .details-panel {
    grid-template-columns: 1fr;
  }

  .details-cover {
    max-width: 12rem;
  }

  .details-body {
    padding-right: 0;
  }
}

h2,
p {
  margin: 0;
}
</style>
