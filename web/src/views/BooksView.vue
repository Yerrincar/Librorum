<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  bookAuthor,
  bookCoverPath,
  bookTitle,
  fetchBooks,
  type LibraryItemKind,
  type LibraryItemResponse,
} from '@/api/books'

const books = ref<LibraryItemResponse[]>([])
const loading = ref(true)
const errorMessage = ref('')
const activeKind = ref<LibraryItemKind | undefined>()

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

    <p v-if="loading">Loading books...</p>
    <p v-else-if="errorMessage" role="alert">{{ errorMessage }}</p>

    <section v-else-if="books.length === 0">
      <p>No books found.</p>
      <RouterLink to="/books/import">Import an EPUB</RouterLink>
    </section>

    <section v-else aria-label="Book covers">
      <article v-for="book in books" :key="book.id ?? bookTitle(book)">
        <img v-if="bookCoverPath(book)" :src="bookCoverPath(book)" :alt="bookTitle(book)" loading="lazy" />
        <div v-else>No cover</div>
        <h2>{{ bookTitle(book) }}</h2>
        <p v-if="bookAuthor(book)">{{ bookAuthor(book) }}</p>
      </article>
    </section>
  </main>
</template>

<style scoped>
section[aria-label='Book covers'] {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(auto-fill, minmax(8rem, 1fr));
}

article {
  display: grid;
  gap: 0.5rem;
}

img,
article > div {
  aspect-ratio: 2 / 3;
  background: #eee;
  object-fit: cover;
  width: 100%;
}

h2,
p {
  margin: 0;
}
</style>
