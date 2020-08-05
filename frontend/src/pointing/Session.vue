<template>
  <main class="container-fluid" role="main">
    <h1>Pointing Session</h1>
    <div v-if="sessionLoaded">
      <div v-if="isParticipating">

      </div>
      <div v-else>
        <h1>This session currently has {{ currentSession.participants.length }} participants.</h1>
        <p>
          The session facilitator is <user-display-name :user="currentSession.facilitator" />,
          <span v-if="currentSession.facilitatorPoints">
            and they are participating in pointing
          </span>
          <span v-else>
            and they are not participating in pointing
          </span>
        </p>
      </div>
    </div>
    <div v-else>
      <loading/>
    </div>
  </main>
</template>

<script lang="ts">
import { Vue, Component, Watch } from 'vue-property-decorator'
import { PointingSession, PointingSessionStore } from '@/pointing/PointingSessionStore'
import Loading from '@/app/Loading.vue'
import UserDisplayName from '@/pointing/UserDisplayName.vue'

@Component({
  components: { UserDisplayName, Loading }
})
export default class Session extends Vue {
  get sessionLoaded(): boolean {
    const sessionId = this.$route.params.sessionId
    return this.$store.state.pointingSession.knownSessions.find((s: PointingSession) => {
      return s.sessionId === sessionId
    })
  }

  get isParticipating(): boolean {
    const sessionId = this.$route.params.sessionId
    return this.$store.state.pointingSession.currentSession && this.$store.state.pointingSession.currentSession.sessionId === sessionId
  }

  get currentSession(): PointingSession {
    const sessionId = this.$route.params.sessionId
    return this.$store.state.pointingSession.knownSessions.find((s: PointingSession) => {
      return s.sessionId === sessionId
    })
  }

  mounted() {
    this.routeParamsChanged()
  }

  @Watch('$route')
  routeParamsChanged() {
    const sessionId = this.$route.params.sessionId
    this.$store.dispatch(PointingSessionStore.ACTION_LOAD_SESSION, { sessionId })
  }
}
</script>

<style lang="scss">

</style>
