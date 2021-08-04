<template>
  <main class="container-fluid" role="main">
    <h1>New Pointing Session</h1>
    <h3 v-if="!isSignedIn">Tip: Log in to save time and avoid having to enter your name!</h3>
    <div v-if="creatingSession || !hasConnectionId">
      <loading id="searchResultLoadingIndicator" />
    </div>
    <div v-else>
      <p>Once the session has been started you will be provided a link which participants may use to join.</p>
      <form @submit.prevent="startSession">
        <div v-if="!isSignedIn" class="form-group">
          <label for="facilitatorName">Facilitator Name:</label>
          <input type="text" class="form-control" id="facilitatorName" aria-describedby="facilitatorHelp" placeholder="Jane Doe" v-model="facilitatorName" />
          <small id="facilitatorHelp" class="form-text text-muted">Name of the individual running the session. Required.</small>
        </div>
        <div v-if="!isSignedIn" class="form-group">
          <label for="facilitatorHandle">Facilitator Handle:</label>
          <input type="text" class="form-control" id="facilitatorHandle" aria-describedby="facilitatorHandleHelp" placeholder="PointMaster2020" v-model="facilitatorHandle" />
          <small id="facilitatorHandleHelp" class="form-text text-muted">If specified this will be the display name of the facilitator, otherwise the value for Facilitator Name will be displayed.</small>
        </div>
        <div class="form-check">
          <input type="radio" name="facilitatorPointing" class="form-check-input" id="facilitatorPointingNo" aria-describedby="facilitatorPointingNoHelp" value="false" v-model="facilitatorPoints" />
          <label class="form-check-label" for="facilitatorPointingNo">Facilitator will not be pointing</label>
          <small id="facilitatorPointingNoHelp" class="form-text text-muted">When selected the facilitator will only control when votes are shown and cleared.</small>
        </div>
        <div class="form-check">
          <input type="radio" name="facilitatorPointing" class="form-check-input" id="facilitatorPointingYes" aria-describedby="facilitatorPointingYesHelp" value="true" v-model="facilitatorPoints"/>
          <label class="form-check-label" for="facilitatorPointingYes">Facilitator will be pointing</label>
          <small id="facilitatorPointingYesHelp" class="form-text text-muted">When selected the facilitator will also have the option to point issues along with the ability to control when votes are shown and cleared.</small>
        </div>
        <button type="submit" class="btn btn-primary" :disabled="disableSubmit" id="startSessionButton">Start Session</button>
      </form>
    </div>
  </main>
</template>

<script lang="ts">
import { Vue, Component, Watch } from 'vue-property-decorator'
import { PointingSessionStore } from '@/pointing/PointingSessionStore'
import Loading from '@/app/Loading.vue'
import { FACILITATE_ROUTE_NAME } from '@/navigation/router'
import { newUser } from '@/user/user'
import { AppStore } from '@/app/AppStore'
import { createSession } from '@/pointing/pointing'

@Component({
  components: {
    Loading
  }
})
export default class NewSession extends Vue {
  facilitatorName: string = ''

  facilitatorHandle: string = ''

  facilitatorPoints: string = 'false'

  creatingSession: boolean = false

  get isSignedIn(): boolean {
    return this.$store.state.profile.signedIn
  }

  get sessionActive(): boolean {
    return !!this.$store.state.pointingSession.currentSession
  }

  get disableSubmit(): boolean {
    return !this.isSignedIn && this.facilitatorName === ''
  }

  get hasConnectionId(): boolean {
    return !!this.$store.state.pointingSession.connectionId
  }

  mounted() {
    this.$store.dispatch(PointingSessionStore.ACTION_END_SESSION)
  }

  async startSession() {
    this.creatingSession = true
    const facilitatorName = this.isSignedIn ? this.$store.state.profile.remoteProfile.name : this.facilitatorName
    const facilitatorHandle = this.isSignedIn ? this.$store.state.profile.remoteProfile.handle : this.facilitatorHandle
    const facilitatorPoints = this.facilitatorPoints === 'true'
    const request = {
      connectionId: this.$store.state.pointingSession.connectionId,
      facilitator: newUser(facilitatorName, facilitatorHandle),
      facilitatorPoints: facilitatorPoints
    }
    try {
      const session = await createSession(request)
      await this.$store.commit(PointingSessionStore.MUTATION_SET_SESSION_ID, session.sessionId)
      await this.$store.commit(PointingSessionStore.MUTATION_SET_SESSION, session)
    } catch (e) {
      this.creatingSession = false
      await this.$store.dispatch(AppStore.ACTION_REGISTER_REMOTE_ERROR, 'Error creating session')
    }
  }

  @Watch('sessionActive')
  currentSessionChanged() {
    if (this.sessionActive) {
      this.$router.push({
        name: FACILITATE_ROUTE_NAME,
        params: {
          sessionId: this.$store.state.pointingSession.currentSession.sessionId,
          facilitatorSessionKey: this.$store.state.pointingSession.currentSession.facilitatorSessionKey
        }
      })
    }
  }
}
</script>

<style lang="scss">

</style>
