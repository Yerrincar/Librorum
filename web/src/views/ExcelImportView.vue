<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { getCurrentUser, type UserResponse } from '@/api/auth'
import { importExcelBooks, type ImportExcelBooksResponse } from '@/api/books'

const user = ref<UserResponse | null>(null)
const checkingUser = ref(true)
const loading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')
const selectedFile = ref<File | null>(null)
const result = ref<ImportExcelBooksResponse | null>(null)

const form = reactive({
  spreadsheet: 'Inventario',
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
    errorMessage.value = 'Select an Excel file first'
    return
  }

  loading.value = true
  errorMessage.value = ''
  successMessage.value = ''
  result.value = null

  try {
    result.value = await importExcelBooks({
      file: selectedFile.value,
      spreadsheet: form.spreadsheet,
    })
    successMessage.value = `Imported ${result.value.imported_count} books`
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Excel import failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main>
    <h1>Excel Import</h1>

    <p v-if="checkingUser">Checking session...</p>

    <section v-else-if="!user">
      <p>You need to be logged in to import books.</p>
      <RouterLink to="/login">Login</RouterLink>
      <RouterLink to="/register">Register</RouterLink>
    </section>

    <form v-else @submit.prevent="submitImport">
      <p>Logged in as {{ user.username }}</p>

      <label>
        Excel file
        <input type="file" name="file" accept=".xlsx,.xlsm" @change="selectFile" />
      </label>

      <label>
        Spreadsheet name
        <input v-model.trim="form.spreadsheet" name="spreadsheet" required />
      </label>

      <button type="submit" :disabled="loading">
        {{ loading ? 'Importing...' : 'Search and import metadata' }}
      </button>

      <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
      <p v-if="successMessage" role="status">{{ successMessage }}</p>

      <section v-if="result">
        <h2>Result</h2>
        <p>Imported: {{ result.imported_count }}</p>
        <p>Skipped: {{ result.skipped_count }}</p>

        <section v-if="result.imported.length > 0">
          <h3>Imported books</h3>
          <ul>
            <li v-for="title in result.imported" :key="title">{{ title }}</li>
          </ul>
        </section>

        <section v-if="result.skipped.length > 0">
          <h3>Skipped rows</h3>
          <ul>
            <li v-for="title in result.skipped" :key="title">{{ title }}</li>
          </ul>
        </section>
      </section>
    </form>
  </main>
</template>
