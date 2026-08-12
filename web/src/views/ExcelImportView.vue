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
  <main class="excel-import-page">
    <div class="import-shell">
      <template v-if="checkingUser || !user">
        <nav class="heading-nav" aria-label="Import navigation">
          <RouterLink to="/" class="heading-link">Home</RouterLink>
          <RouterLink to="/books" class="heading-link">Library</RouterLink>
        </nav>
        <h1>Excel Import</h1>
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
            <h1>Excel Import</h1>
          </div>

          <section class="source-section">
            <h2>Excel file</h2>
            <label class="file-field">
              File
              <input type="file" name="file" accept=".xlsx,.xlsm" @change="selectFile" />
              <span class="file-action">Choose Excel file</span>
              <span>{{ selectedFile?.name ?? 'No file selected' }}</span>
            </label>
          </section>

          <section class="source-section spreadsheet-section">
            <h2>Spreadsheet</h2>
            <label>
              Sheet name
              <input v-model.trim="form.spreadsheet" name="spreadsheet" required />
            </label>
          </section>
        </div>

        <div class="import-actions">
          <button type="submit" :disabled="loading">
            {{ loading ? 'Importing…' : 'Search and import metadata' }}
          </button>

          <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
          <p v-if="successMessage" role="status">{{ successMessage }}</p>
        </div>

        <section v-if="result" class="result-panel">
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
    </div>
  </main>
</template>

<style scoped>
.excel-import-page{
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
  padding: 1.25rem;
  width: 100%;
}

h1,
h2,
h3,
.heading-link,
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
.result-panel li {
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
.result-panel,
.result-panel section {
  display: grid;
  gap: 0.85rem;
}

form {
  align-content: start;
  align-items: stretch;
  box-sizing: border-box;
  gap: 0.55rem;
  max-width: 100%;
  width: 100%;
}

.import-top-row {
  align-items: stretch;
  display: grid;
  gap: 0.65rem;
  grid-template-columns: minmax(11rem, 0.7fr) minmax(18rem, 2fr) minmax(16rem, 1fr);
  width: 100%;
}

.import-heading,
.source-section,
.auth-panel,
.result-panel,
.result-panel section {
  background: rgba(5, 3, 1, 0.32);
  border: 1px solid rgba(212, 175, 55, 0.36);
  box-sizing: border-box;
  max-width: 100%;
  padding: 1rem;
}

.import-heading,
.source-section {
  align-content: start;
  display: grid;
  gap: 0.45rem;
  min-height: 8.25rem;
  min-width: 0;
  padding: 0.75rem 0.85rem;
}

.import-heading {
  gap: 0.1rem;
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

.import-top-row h2 {
  font-size: 1rem;
  line-height: 1.1;
  margin-bottom: 0.25rem;
}

.import-top-row label {
  font-size: 0.86rem;
  gap: 0.25rem;
}

label {
  display: grid;
  gap: 0.35rem;
  letter-spacing: 1px;
}

input {
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
  font-family: 'Cinzel', serif;
  font-size: 1rem;
  letter-spacing: 1px;
  padding: 0.5rem 0.75rem;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 8px rgba(212,175,55,0.55);
  width: fit-content;
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

.import-actions {
  align-items: center;
  display: flex;
  flex-wrap: wrap;
  gap: 0.85rem 1rem;
}

.import-actions p,
.result-panel p,
.result-panel ul {
  margin: 0;
}

.result-panel ul {
  display: grid;
  gap: 0.35rem;
  padding-left: 1.25rem;
}

@media (max-width: 1050px) {
  .import-top-row {
    grid-template-columns: 1fr 1fr;
  }

  .spreadsheet-section {
    grid-column: 1 / -1;
  }
}

@media (max-width: 680px) {
  .excel-import-page {
    padding: 1rem;
  }

  .import-shell {
    padding: 0.85rem;
  }

  .import-top-row {
    grid-template-columns: 1fr;
  }
}
</style>
