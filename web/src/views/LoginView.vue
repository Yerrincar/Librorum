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
const successMessage = ref('')

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

function hasValue(value: string) {
  return value.length > 0
}
</script>

<template>
  <main class="login-page">

    <form autocomplete="on" @submit.prevent="submitLogin">

      <div class="form__group"
      :class="{ filled: hasValue(form.username) }">
        <input
          id="login-username"
          v-model.trim="form.username"
          class="form__field"
          name="username"
          autocomplete="username"
          spellcheck="false"
          placeholder=" "
          required
        />
        <label class="form__label" for="login-username">
          Username
        </label>
      </div>

      <div class="form__group"
      :class="{ filled: hasValue(form.password) }">
        <input
          id="login-password"
          v-model="form.password"
          class="form__field"
          name="password"
          type="password"
          autocomplete="current-password"
          minlength="8"
          placeholder=" "
          required
        />
        <label class="form__label" for="login-password">
          Password
        </label>
      </div>


      <button type="submit" :disabled="loading">
        {{ loading ? 'Signing in…' : 'Login' }}
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
.login-page {
  min-height: 100vh;
  width: 100%;

  display: flex;
  align-items: center;
  justify-content: flex-start;

  background-image:
    linear-gradient(rgba(0, 0, 0, 0.24), rgba(0, 0, 0, 0.24)),
    url('../../images/ARGONATH.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;

  overflow-x: hidden;
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
  border-bottom: 1px solid rgba(212, 175, 55, 0.82);

  outline: none;

  color: #fff0ad;

  padding: 7px 0;

  background: transparent;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.98),
    0 0 8px rgba(0,0,0,0.72);

  transition: border-color 0.2s;
}


.form__field::placeholder {
  color: transparent;
}


.form__label {
  position: absolute;
  font-family: 'Cinzel', serif;
  font-size: 1.2rem;
  top: 20px;
  left: 0;

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
  filter: brightness(1.28);

  pointer-events: none;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.98),
    0 0 10px rgba(212,175,55,0.45);

  transition: color 0.2s, font-weight 0.2s, top 0.2s;
}


.form__field:focus {
  border-width: 3px;

  border-image: linear-gradient(
    to right,
    #ffe9a3,
    #d4af37,
    #8c5a18
  );

  border-image-slice: 1;
}

.form__field:focus-visible {
  outline: none;
}


.form__field:focus ~ .form__label {
  top: 0;

  filter: brightness(1.45);
  font-weight: 700;
}


.form__field:placeholder-shown ~ .form__label {
  top: 20px;
}

.form__group.filled .form__label {
  top: 0;
  filter: brightness(1.45);
  font-weight: 700;
}

button {
  justify-self: center;
  padding: 0.5rem 1rem;
  border: 1px solid rgba(212, 175, 55, 0.72);
  border-radius: 0;
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
  cursor: pointer;
  font-family: 'Cinzel', serif;
  font-size: 0.9rem;
  letter-spacing: 0.8px;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.98),
    0 0 10px rgba(212,175,55,0.45);
  filter: brightness(1.22);
}


button:disabled {
  cursor: wait;
  opacity: 0.7;
}


button:focus-visible {
  outline: 2px solid rgba(158, 101, 18, 0.9);
  outline-offset: 4px;
}


.error {
  color: #ff9b85;
  font-family: 'Cinzel', serif;
  text-shadow: 1px 2px 4px rgba(0,0,0,0.95);
}


.success {
  color: #fff0ad;
  font-family: 'Cinzel', serif;
  text-shadow: 1px 2px 4px rgba(0,0,0,0.95);
}
</style>
