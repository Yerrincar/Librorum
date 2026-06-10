<script setup lang="ts">
import { ref } from 'vue'
import { getCurrentUser, logoutUser, type UserResponse } from '@/api/auth'

const user = ref<UserResponse | null>(null)
const message = ref('')
const errorMessage = ref('')

async function loadCurrentUser() {
  message.value = ''
  errorMessage.value = ''

  try {
    user.value = await getCurrentUser()
    message.value = `Logged in as ${user.value.username}`
  } catch (error) {
    user.value = null
    errorMessage.value = error instanceof Error ? error.message : 'Current user failed'
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
        <li><RouterLink to="/register">Register</RouterLink></li>
        <li><RouterLink to="/login">Login</RouterLink></li>
      </ul>
    </nav>

    <section aria-label="Auth test controls">
      <h2>Auth test</h2>
      <button type="button" @click="loadCurrentUser">Current user</button>
      <button type="button" @click="submitLogout">Logout</button>

      <p v-if="message" role="status">{{ message }}</p>
      <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
    </section>
  </main>
</template>
