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
      errorMessage.value =
        error instanceof Error ? error.message : 'Current user failed'
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
    errorMessage.value =
      error instanceof Error ? error.message : 'Logout failed'
  }
}
</script>


<template>
  <main class="home-page">

    <div class="home-content">

      <h1 class="logo">
        Librorum
      </h1>


      <nav class="links__group" aria-label="Main navigation">

        <ul>

          <template v-if="user">

            <li class="links">
              <RouterLink to="/books">
                Books
              </RouterLink>
            </li>

            <li class="links">
              <RouterLink to="/books/import">
                Add Books
              </RouterLink>
            </li>

            <li class="links">
              <RouterLink to="/books/import/excel">
                Excel Import
              </RouterLink>
            </li>

            <li class="links">
              <button type="button" @click="submitLogout">
                Logout
              </button>
            </li>

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

    </div>


    <p v-if="message" class="message">
      {{ message }}
    </p>


    <p v-if="errorMessage" class="error">
      {{ errorMessage }}
    </p>


  </main>
</template>


<style scoped>

.home-page {
  box-sizing: border-box;
  min-height: 100vh;
  padding: clamp(1rem, 4vw, 3rem);
  width: 100%;

  display: flex;
  justify-content: center;
  align-items: center;

  background-image:
    linear-gradient(rgba(0, 0, 0, 0.20), rgba(0, 0, 0, 0.20)),
    url('../../images/LOTR-HOME.jpg');
  background-size: cover;
  background-position: center;
  background-repeat: no-repeat;

  overflow-x: hidden;
}


.home-content {
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  max-width: min(100%, 52rem);
  text-align: center;
  transform: translateY(clamp(-9rem, -14vh, -2rem));
}


.logo {
  margin: 0 0 2rem 0;
  font-family: 'Cinzel', serif;
  font-size: clamp(3rem, 10vw, 5rem);
  letter-spacing: clamp(2px, 0.8vw, 5px);
  overflow-wrap: anywhere;

  background: linear-gradient(
    180deg,
    #fff3b0 0%,
    #e7c75a 25%,
    #c18b20 55%,
    #7a4b08 85%,
    #3b2104 100%
  );

  -webkit-background-clip: text;
  background-clip: text;
  color: transparent;

  text-shadow:
    2px 3px 4px rgba(0,0,0,0.9),
    0 0 15px rgba(255,215,100,0.4);

  filter: brightness(1.1);
}


.links__group ul {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1.3rem;
}


.links {
  font-family: 'Cinzel', serif;
  font-size: clamp(1.35rem, 4vw, 2rem);
  font-weight: 400;
  letter-spacing: clamp(1.5px, 0.55vw, 3px);
  max-width: 100%;
}


.links a,
.links button {
  display: inline-block;
  max-width: min(100%, 90vw);
  overflow-wrap: anywhere;

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

  text-decoration: none;

  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 10px rgba(212,175,55,0.75);

  filter: brightness(1.4);

  transition: filter 0.3s ease, transform 0.3s ease;
}


.links button {
  border: none;
  padding: 0;
  background-color: transparent;
  cursor: pointer;
  font-family: inherit;
  font-size: inherit;
  letter-spacing: inherit;
}


.links a:hover,
.links button:hover,
.links a:focus-visible,
.links button:focus-visible {
  transform: translateY(-3px);
  filter: brightness(1.3);
  outline: none;
  text-shadow:
    2px 3px 5px rgba(0,0,0,0.95),
    0 0 14px rgba(212,175,55,0.95);
}


.message {
  position: absolute;
  bottom: 40px;
  left: 1rem;
  right: 1rem;
  font-family: 'Cinzel', serif;
  color: #f5deb3;
  text-align: center;
}


.error {
  position: absolute;
  bottom: 40px;
  left: 1rem;
  right: 1rem;
  color: #b00020;
  text-align: center;
}

@media (max-height: 620px) {
  .home-content {
    transform: translateY(-1rem);
  }

  .logo {
    margin-bottom: 1.25rem;
  }

  .links__group ul {
    gap: 0.8rem;
  }
}

</style>
