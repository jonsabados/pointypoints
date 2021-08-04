import Vue from 'vue'
import App from './App.vue'
import router from './navigation/router'
import Vuex from 'vuex'
import VueRouter from 'vue-router'
import { BootstrapVue, BootstrapVueIcons } from 'bootstrap-vue'
import VueGtag from 'vue-gtag'
import { PointingSessionStore } from '@/pointing/PointingSessionStore'
import { AppStore } from '@/app/AppStore'
import { ProfileStore } from '@/profile/ProfileStore'
// @ts-ignore
import { LoaderPlugin } from 'vue-google-login'

Vue.config.productionTip = false
Vue.use(Vuex)
Vue.use(VueRouter)
Vue.use(BootstrapVue)
Vue.use(BootstrapVueIcons)

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

Vue.use(LoaderPlugin, {
  client_id: process.env.VUE_APP_GOOGLE_OAUTH_CLIENT_ID
})

const store = new Vuex.Store<RootState>({
  state: {},
  modules: {
    app: AppStore,
    pointingSession: PointingSessionStore,
    profile: ProfileStore
  }
})

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
