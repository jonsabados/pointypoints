<template>
  <div id="app">
    <main-nav />
    <div>
      <router-view/>
    </div>
    <b-modal v-model="hasRemoteError" role="alert" ok-only id="remoteErrorDialog" title="An error has occurred">
      <p id="remoteErrorMessage">{{ currentError }}</p>
    </b-modal>
    <footer class="footer bg-light fixed-bottom" id="footer">
      <nav role="navigation" class="navbar navbar-expand-lg navbar-light bg-light justify-content-sm-center" id="footerNav">
        <ul class="navbar-nav">
          <li class="nav-item">
            <router-link :to="{name: 'privacy'}" class="nav-link">Privacy Policy</router-link>
          </li>
        </ul>
      </nav>
    </footer>
  </div>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'
import MainNav from './navigation/MainNav.vue'
import { AppStore } from '@/app/AppStore'
import { PointingSessionStore } from '@/pointing/PointingSessionStore'

@Component({
  components: {
    MainNav
  }
})
export default class App extends Vue {
  created() {
    this.$store.dispatch(PointingSessionStore.ACTION_INITIALIZE)
  }

  get currentError(): null | string {
    return this.$store.state.app.errorToAck
  }

  get hasRemoteError(): boolean {
    return this.$store.state.app.errorToAck != null
  }

  set hasRemoteError(error) {
    if (!error) {
      this.$store.dispatch(AppStore.ACTION_ACK_REMOTE_ERROR)
    }
  }
}
</script>

<style lang="scss">
@import "~bootstrap/scss/bootstrap";
html {
  position: relative;
  min-height: 100%;
}
</style>
