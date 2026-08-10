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
  <main class="home-page">
    <h1 class="logo">Librorum</h1>

<nav class="links__group" aria-label="Test pages">
  <ul>

    <template v-if="user">
      <li class="links">
        <RouterLink to="/books">Books</RouterLink>
      </li>

      <li class="links">
        <RouterLink to="/books/import">
          Import books
        </RouterLink>
      </li>

      <li class="links">
        <RouterLink to="/books/import/excel">
          Excel Import
        </RouterLink>
      </li>
    <section aria-label="Auth test controls">
      <button class="links" type="button" @click="submitLogout">Logout</button>

      <p v-if="message" role="status">{{ message }}</p>
      <p v-if="errorMessage" role="alert">{{ errorMessage }}</p>
    </section>
    </template>


    <template v-else>
      <li class="links">
        <RouterLink to="/register">
          Sign Up
        </RouterLink>
      </li>

      <li class="links">
        <RouterLink to="/login">
          Login
        </RouterLink>
      </li>
    </template>

  </ul>
</nav>
  </main>
</template>

<style scoped>
.home-page {
  min-height: 100vh;
  width: 100%;

  display: flex;
  align-items: center;
  justify-content: flex-start;

  background-image: url('../../images/LOTR-HOME.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;

  overflow: hidden;
}

.logo {
  position: absolute;
  top: 200px;
  right: 250px;

  margin: 0;
  padding: 0;

  font-family: 'Cinzel', serif;
  font-size: 5rem;
  letter-spacing: 3px;
 background: linear-gradient(
    to bottom,
    #fff1a8 0%,
    #d4af37 35%,
    #a66a16 65%,
    #6b3f0a 100%
  );

  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;

  text-shadow:
    2px 2px 3px rgba(0, 0, 0, 0.8),
    0 0 12px rgba(212, 175, 55, 0.35);

  text-decoration: none;
}

.links__group{
  position: absolute;
  top: 220px;
  left: 150px;
}

.links__group ul{
  list-style: none;

  margin: 0;
  padding: 0;

  display: flex;
  flex-direction: column;

  gap: 1.5rem;
}

.links {
  font-family: 'Cinzel', serif;
  font-size: 3rem;
  letter-spacing: 3px;
}

.links a {
 background: linear-gradient(
    to bottom,
    #fff1a8 0%,
    #d4af37 35%,
    #a66a16 65%,
    #6b3f0a 100%
  );

  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;

  text-shadow:
    2px 2px 3px rgba(0, 0, 0, 0.8),
    0 0 12px rgba(212, 175, 55, 0.35);

  text-decoration: none;
}

.error {
  color: #b00020;
}


.success {
  color: #176b36;
}
</style>
}
