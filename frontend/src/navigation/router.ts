import VueRouter from 'vue-router'
import Home from '@/home/Home.vue'
import About from '@/about/About.vue'
import Search from '@/search/SearchDisplay.vue'
import Articles from '@/articles/Articles.vue'
import ArticlesHome from '@/articles/ArticlesHome.vue'
import Article from '@/articles/Article.vue'
import Privacy from '@/legal/Privacy.vue'

export const HOME_ROUTE_NAME = 'home'

const routes = [
  {
    path: '/',
    name: HOME_ROUTE_NAME,
    component: Home
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

export default router
