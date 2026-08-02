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
    errorMessage.value =
      error instanceof Error ? error.message : 'Registration failed'
  } finally {
    loading.value = false
  }
}

function hasValue(value: string) {
  return value.length > 0
}

</script>

<template>
  <main class="register-page">

    <form autocomplete="on" @submit.prevent="submitRegister">

      <div class="form__group"
      :class="{ filled: hasValue(form.username) }">
        <input
          v-model.trim="form.username"
          class="form__field"
          name="username"
          autocomplete="username"
          placeholder=" "
          required
        />
        <label class="form__label">
          Username
        </label>
      </div>


      <div class="form__group"
      :class="{ filled: hasValue(form.email) }">
        <input
          v-model.trim="form.email"
          class="form__field"
          name="email"
          type="email"
          autocomplete="email"
          placeholder=" "
          required
        />
        <label class="form__label">
          Email
        </label>
      </div>


      <div class="form__group"
      :class="{ filled: hasValue(form.display_name) }">
        <input
          v-model.trim="form.display_name"
          class="form__field"
          name="display_name"
          autocomplete="name"
          placeholder=" "
          required
        />
        <label class="form__label">
          Display name
        </label>
      </div>


      <div class="form__group"
      :class="{ filled: hasValue(form.password) }">
        <input
          v-model="form.password"
          class="form__field"
          name="password"
          type="password"
          autocomplete="new-password"
          minlength="8"
          placeholder=" "
          required
        />
        <label class="form__label">
          Password
        </label>
      </div>


      <button type="submit" :disabled="loading">
        {{ loading ? 'Signing up...' : 'Sign Up' }}
      </button>


      <p v-if="errorMessage" class="error" role="alert">
        {{ errorMessage }}
      </p>

      <p v-if="successMessage" class="success" role="status">
        {{ successMessage }}
      </p>

    </form>

  </main>
</template>


<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Cinzel:wght@400;700&display=swap');
.register-page {
  min-height: 100vh;
  width: 100%;

  display: flex;
  align-items: center;
  justify-content: flex-start;

  background-image: url('../../images/ARGONATH.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;

  overflow: hidden;
}


form {
  display: grid;
  gap: 1.2rem;
  padding: 0 0 200px 100px;

  width: 280px;
  margin-left: 40px;
}

.form__group {
  position: relative;
  padding-top: 20px;
}


.form__field {
  font-family: 'Cinzel', serif;
  font-size: 1.2rem;
  width: 100%;

  border: none;
  border-bottom: 2px solid #140b01;

  outline: none;

  color: #140b01;

  padding: 7px 0;

  background: transparent;

  transition: border-color 0.2s;
}


.form__field::placeholder {
  color: transparent;
}


.form__label {
  position: absolute;

  top: 20px;
  left: 0;

  font-family: 'Cinzel', serif;
  font-size: 1.2rem;

  color: #140b01;

  pointer-events: none;
  transition: 0.2s;
}

.form__field:focus {
  border-width: 3px;

  border-image: linear-gradient(
    to right,
    #7B3F00,
    #FF8C00
  );

  border-image-slice: 1;
}


.form__field:focus ~ .form__label {
  top: 0;

  color: #9c570b;
  font-weight: 700;
}


.form__field:placeholder-shown ~ .form__label {
  top: 20px;
}

.form__group.filled .form__label {
  top: 0;
  color: #9c570b;
  font-weight: 700;
}

button {
  width: 120px;

  justify-self: center;

  padding: 0.5rem 1rem;

  border: none;
  border-radius: 0.4rem;

  background: linear-gradient(
    to right,
    #7B3F00,
    #7B3F00
  );

  color: white;

  cursor: pointer;

  font-family: 'Cinzel', serif;
  font-size: 0.9rem;
  letter-spacing: 0.8px;
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
