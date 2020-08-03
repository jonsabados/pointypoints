import VueRouter from 'vue-router'
import Home from '@/home/Home.vue'
import NewSession from '@/pointing/NewSession.vue'

export const HOME_ROUTE_NAME = 'home'
export const NEW_SESSION_ROUTE_NAME = 'newSession'

const routes = [
  {
    path: '/',
    name: HOME_ROUTE_NAME,
    component: Home
  },
  {
    path: '/session/new',
    name: NEW_SESSION_ROUTE_NAME,
    component: NewSession
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
