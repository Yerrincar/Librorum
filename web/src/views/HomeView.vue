<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getCurrentUser, logoutUser, type UserResponse } from '@/api/auth'

const user = ref<UserResponse | null>(null)
const message = ref('')
const errorMessage = ref('')

onMounted(() => {
  void loadCurrentUser(true)
})

async function loadCurrentUser(silent = false) {
  message.value = ''
  if (!silent) {
    errorMessage.value = ''
  }

  try {
    user.value = await getCurrentUser()
    if (!silent) {
      message.value = `Logged in as ${user.value.username}`
    }
  } catch (error) {
    user.value = null
    if (!silent) {
      errorMessage.value = error instanceof Error ? error.message : 'Current user failed'
    }
  }
}

async function submitLogout() {
  message.value = ''
  errorMessage.value = ''

  try {
    await logoutUser()
    user.value = null
    message.value = 'Logged out'
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Logout failed'
  }
}
</script>

<template>
  <main>
    <h1>Librorum</h1>

    <nav aria-label="Test pages">
      <ul>
        <li><RouterLink to="/books">Books</RouterLink></li>
        <li v-if="user"><RouterLink to="/books/import">Import books</RouterLink></li>
        <li v-if="user"><RouterLink to="/books/import/excel">Excel Import</RouterLink></li>
        <li><RouterLink to="/register">Register</RouterLink></li>
        <li><RouterLink to="/login">Login</RouterLink></li>
      </ul>
    </nav>

    <section aria-label="Auth test controls">
      <h2>Auth test</h2>
      <button type="button" @click="loadCurrentUser(false)">Current user</button>
      <button type="button" @click="submitLogout">Logout</button>

      <p v-if="message" role="status">{{ message }}</p>
      <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
    </section>
  </main>
</template>
