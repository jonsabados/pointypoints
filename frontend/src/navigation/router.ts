import VueRouter from 'vue-router'
import Home from '@/home/Home.vue'
import NewSession from '@/pointing/NewSession.vue'
import Session from '@/pointing/Session.vue'
import Facilitate from '@/pointing/FacilitateSession.vue'
import Profile from '@/profile/Profile.vue'
import Privacy from '@/Privacy.vue'

export const HOME_ROUTE_NAME = 'home'
export const NEW_SESSION_ROUTE_NAME = 'newSession'
export const SESSION_ROUTE_NAME = 'session'
export const FACILITATE_ROUTE_NAME = 'facilitate'
export const PROFILE_ROUTE_NAME = 'profile'
export const PRIVACY_ROUTE_NAME = 'privacy'

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
  },
  {
    path: '/session/:sessionId',
    name: SESSION_ROUTE_NAME,
    component: Session
  },
  {
    path: '/facilitate/:sessionId/:facilitatorSessionKey',
    name: FACILITATE_ROUTE_NAME,
    component: Facilitate
  },
  {
    path: '/profile',
    name: PROFILE_ROUTE_NAME,
    component: Profile
  },
  {
    path: '/privacy',
    name: PRIVACY_ROUTE_NAME,
    component: Privacy
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
