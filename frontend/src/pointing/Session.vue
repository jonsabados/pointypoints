<template>
  <main class="container-fluid" role="main">
    <h1>Pointing Session</h1>
    <div v-if="sessionLoaded">
      <div v-if="isParticipating">
        <pointing :session="currentSession" :user-id="userId" />
        <pointing-results v-if="currentSession.votesShown" :session="currentSession" />
        <p v-else>
          Votes are currently hidden. Once the facilitator chooses
          <span v-if="currentSession.facilitatorPoints">their vote and</span>
          the votes of the
          {{ currentSession.participants.length }} participants will be shown.
        </p>
      </div>
      <div v-else>
        <h4>This session currently has {{ currentSession.participants.length }} participants.</h4>
        <p>
          The session facilitator is
          <user-display-name :user="currentSession.facilitator"/>,
          <span v-if="currentSession.facilitatorPoints">
            and they are participating in pointing
          </span>
          <span v-else>
            and they are not participating in pointing
          </span>
        </p>
        <div v-if="needDetails">
          <p>You must enter some details before you can join the session:</p>
          <form @submit.prevent="joinSession">
            <div class="form-group">
              <label for="Name">Name:</label>
              <input type="text" class="form-control" id="name" aria-describedby="nameHelp" placeholder="Jane Doe" v-model="name"/>
              <small id="Help" class="form-text text-muted">Name of the individual running the session. Required.</small>
            </div>
            <div class="form-group">
              <label for="Handle"> Handle:</label>
              <input type="text" class="form-control" id="Handle" aria-describedby="HandleHelp"
                     placeholder="PointMaster2020" v-model="handle"/>
              <small id="HandleHelp" class="form-text text-muted">If specified this will be the name displayed to other
                participants, otherwise the value for Name will be displayed.</small>
            </div>
            <button type="submit" class="btn btn-primary" :disabled="detailsIncomplete" id="startSessionButton">Join Session</button>
          </form>
        </div>
        <div v-else>
          <loading/>
        </div>
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
import UserDisplayName from '@/user/UserDisplayName.vue'
import { v4 as uuidv4 } from 'uuid'
import Pointing from '@/pointing/Pointing.vue'
import PointingResults from '@/pointing/PointingResults.vue'

@Component({
  components: { PointingResults, Pointing, UserDisplayName, Loading }
})
export default class Session extends Vue {
  userId: string = uuidv4()
  name: string = ''
  handle: string = ''
  detailsSet: boolean = false

  // TODO - this should come from being logged in or something but will work for POC stuff
  get needDetails(): boolean {
    return !this.detailsSet
  }

  get detailsIncomplete(): boolean {
    return this.name === ''
  }

  get sessionLoaded(): boolean {
    const sessionId = this.$route.params.sessionId
    return this.$store.state.pointingSession.knownSessions.find((s: PointingSession) => {
      return s.sessionId === sessionId
    })
  }

  get isParticipating(): boolean {
    return this.sessionLoaded && this.currentSession.participants.find((u) => {
      return u.userId === this.userId
    }) !== undefined
  }

  get currentSession(): PointingSession {
    const sessionId = this.$route.params.sessionId
    return this.$store.state.pointingSession.knownSessions.find((s: PointingSession) => {
      return s.sessionId === sessionId
    })
  }

  joinSession() {
    this.detailsSet = true
    this.$store.dispatch(PointingSessionStore.ACTION_JOIN_SESSION, {
      sessionId: this.$route.params.sessionId,
      user: {
        userId: this.userId,
        name: this.name,
        handle: this.handle
      }
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
