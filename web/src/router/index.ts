import { createRouter, createWebHistory } from 'vue-router'
import BooksView from '@/views/BooksView.vue'
import ExcelImportView from '@/views/ExcelImportView.vue'
import HomeView from '@/views/HomeView.vue'
import ImportBooksView from '@/views/ImportBooksView.vue'
import LoginView from '@/views/LoginView.vue'
import RegisterView from '@/views/RegisterView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/books',
      name: 'books',
      component: BooksView,
    },
    {
      path: '/books/import',
      name: 'import-books',
      component: ImportBooksView,
    },
    {
      path: '/books/import/excel',
      name: 'excel-import',
      component: ExcelImportView,
    },
    {
      path: '/register',
      name: 'register',
      component: RegisterView,
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
    },
  ],
})

export default router
