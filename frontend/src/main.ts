import Vue from 'vue'
import App from './App.vue'
import router from './navigation/router'
import Vuex from 'vuex'
import VueRouter from 'vue-router'
import Bootstrap from 'bootstrap-vue'
import VueGtag from 'vue-gtag'
import { PointingSessionStore } from '@/pointing/PointingSessionStore'
import { AppStore } from '@/app/AppStore'

Vue.config.productionTip = false
Vue.use(Vuex)
Vue.use(VueRouter)
Vue.use(Bootstrap)

interface RootState {
}

Vue.use(VueGtag, {
  config: {
    id: process.env.VUE_APP_GOOGLE_ANALYTICS_ID
  },
  bootstrap: false,
  appName: 'pointypoints.com',
  pageTrackerScreenviewEnabled: true
}, router)

const store = new Vuex.Store<RootState>({
  state: {},
  modules: {
    app: AppStore,
    pointingSession: PointingSessionStore
  }
})

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
