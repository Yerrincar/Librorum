<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { loginUser } from '@/api/auth'

const router = useRouter()

const form = reactive({
  username: '',
  password: '',
})

const loading = ref(false)
const errorMessage = ref('')

async function submitLogin() {
  loading.value = true
  errorMessage.value = ''

  try {
    await loginUser({ ...form })
    form.password = ''
    await router.push('/')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main>
    <form autocomplete="on" @submit.prevent="submitLogin">
      <h1>Login</h1>

      <label>
        Username
        <input v-model.trim="form.username" name="username" autocomplete="username" required />
      </label>

      <label>
        Password
        <input
          v-model="form.password"
          name="password"
          type="password"
          autocomplete="current-password"
          required
        />
      </label>

      <button type="submit" :disabled="loading">
        {{ loading ? 'Logging in...' : 'Login' }}
      </button>

      <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
    </form>
  </main>
</template>
