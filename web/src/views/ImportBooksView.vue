<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getCurrentUser, type UserResponse } from '@/api/auth'
import { importEpubBook } from '@/api/books'

const user = ref<UserResponse | null>(null)
const checkingUser = ref(true)
const loading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const selectedFile = ref<File | null>(null)

const form = reactive({
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

async function submitImport() {
  if (!selectedFile.value) {
    errorMessage.value = 'Select an EPUB file first'
    return
  }

  loading.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const book = await importEpubBook({
      file: selectedFile.value,
      ...form,
    })
    successMessage.value = `Imported ${book.title}`
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
        <h2>EPUB file</h2>
        <label>
          File
          <input type="file" name="file" accept=".epub,application/epub+zip" required @change="selectFile" />
        </label>
      </section>

      <section>
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
        {{ loading ? 'Importing...' : 'Import EPUB' }}
      </button>

      <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
      <p v-if="successMessage" role="status">{{ successMessage }}</p>
    </form>
  </main>
</template>
