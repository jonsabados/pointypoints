<template>
  <div id="app">
    <main-nav />
    <div>
      <router-view/>
      Current session: {{ currentSession }}
    </div>
  </div>
</template>

<script lang="ts">
import { Vue, Component, Watch } from 'vue-property-decorator'
import MainNav from './navigation/MainNav.vue'
import { PointingSession, PointingSessionStore } from '@/pointing/PointingSessionStore'
import { HOME_ROUTE_NAME } from '@/navigation/router'
import { apiSocket } from '@/api/api'

const EVENT_TYPE_SESSION_CREATED = 'SESSION_CREATED'

@Component({
  components: {
    MainNav
  }
})
export default class App extends Vue {
  created() {
    apiSocket().onmessage = (ev) => {
      const eventData = JSON.parse(ev.data)
      if (!eventData.type) {
        console.log('event does not match expected interface')
        console.log(eventData)
        return
      }
      switch (eventData.type) {
        case EVENT_TYPE_SESSION_CREATED:
          this.$store.commit(PointingSessionStore.MUTATION_ADD_SESSION, {
            isFacilitator: true,
            facilitatorSessionKey: eventData.body.facilitatorSessionKey,
            sessionId: eventData.body.sessionId,
            facilitator: eventData.body.facilitator,
            participants: eventData.body.participants
          })
          this.$store.commit(PointingSessionStore.MUTATION_SET_SESSION, eventData.body.sessionId)
          break
        default:
          console.log('unknown event')
          console.log(eventData)
      }
    }
  }

  get currentSession(): PointingSession | undefined {
    return this.$store.state.pointingSession.currentSession
  }

  currentSessionChanged() {
    if (this.$store.state.pointingSession.currentSession) {
      console.log('would open session')
    } else {
      console.log('current session cleared, redirecting to home')
      this.$router.push(({
        name: HOME_ROUTE_NAME
      }))
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
