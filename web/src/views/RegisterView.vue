<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { registerUser } from '@/api/auth'

const router = useRouter()

const form = reactive({
  username: '',
  email: '',
  display_name: '',
  password: '',
})

const loading = ref(false)
const errorMessage = ref('')
const successMessage = ref('')

async function submitRegister() {
  loading.value = true
  errorMessage.value = ''
  successMessage.value = ''

  try {
    const user = await registerUser({ ...form })
    successMessage.value = `Registered ${user.username}`
    form.password = ''
    await router.push('/')
  } catch (error) {
    errorMessage.value = error instanceof Error ? error.message : 'Registration failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="register-page">
    <form class="register-form" autocomplete="on" @submit.prevent="submitRegister">
      <h1>Register</h1>

      <label>
        Username
        <input v-model.trim="form.username" name="username" autocomplete="username" required />
      </label>

      <label>
        Email
        <input v-model.trim="form.email" name="email" type="email" autocomplete="email" required />
      </label>

      <label>
        Display name
        <input v-model.trim="form.display_name" name="display_name" autocomplete="name" required />
      </label>

      <label>
        Password
        <input
          v-model="form.password"
          name="password"
          type="password"
          autocomplete="new-password"
          minlength="8"
          required
        />
      </label>

      <button type="submit" :disabled="loading">
        {{ loading ? 'Registering...' : 'Register' }}
      </button>

      <p v-if="errorMessage" class="error" role="alert">{{ errorMessage }}</p>
      <p v-if="successMessage" class="success" role="status">{{ successMessage }}</p>
    </form>
  </main>
</template>

<style scoped>
.register-page {
  display: grid;
  min-height: 100vh;
  place-items: center;
  padding: 1rem;
}

.register-form {
  display: grid;
  gap: 0.875rem;
  width: min(100%, 24rem);
}

label {
  display: grid;
  gap: 0.35rem;
}

input,
button {
  border: 1px solid #c9c9c9;
  border-radius: 0.4rem;
  font: inherit;
  padding: 0.65rem 0.75rem;
}

button {
  cursor: pointer;
}

button:disabled {
  cursor: wait;
  opacity: 0.7;
}

.error {
  color: #b00020;
}

.success {
  color: #176b36;
}
</style>
