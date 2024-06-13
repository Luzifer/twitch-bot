import { createRouter, createMemoryHistory } from 'vue-router'

//import AuthView from './components/auth.vue'
//import ChatView from './components/chatview.vue'

const routes = [
  //  {
  //    component: AuthView,
  //    path: '/',
  //  },
  //  {
  //    component: ChatView,
  //    path: '/chat',
  //  },
]

const router = createRouter({
  history: createMemoryHistory(),
  routes,
})

export default router
