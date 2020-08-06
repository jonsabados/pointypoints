<template>
  <main class="container-fluid" role="main">
    <h1>Facilitating Session</h1>
    <div v-if="isSessionReady">
      <div v-if="teamEmpty">
        <p>No team members have joined the session. They may do so by going to the following URL: <strong>{{ userURL }}</strong></p>
      </div>
      <div v-else>
        <h4>Active Team Members</h4>
        <table class="table table-striped table-bordered">
          <thead>
            <tr>
              <th scope="col">Name</th>
              <th scope="col">Handle</th>
              <th v-if="votesShown" scope="col">Vote</th>
              <th v-else scope="col">Vote Ready</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in currentUsers" :key="user.userId">
              <td>{{ user.name }}</td>
              <td>{{ user.handle }}</td>
              <td v-if="votesShown">
                <div v-if="user.currentVote">
                  {{ user.currentVote }}
                </div>
                <div v-else>
                  -
                </div>
              </td>
              <td v-else>
                <div v-if="user.currentVote">
                  Yes
                </div>
                <div v-else>
                  No
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        <p>Additional team members may join by going to the following URL: <strong>{{ userURL }}</strong></p>
      </div>
    </div>
    <div v-else>
      <loading />
    </div>
  </main>
</template>

<script lang="ts">
import { Vue, Component, Watch } from 'vue-property-decorator'
import { PointingSessionStore } from '@/pointing/PointingSessionStore'
import Loading from '@/app/Loading.vue'
import { User } from '@/user/user'

@Component({
  components: { Loading }
})
export default class Session extends Vue {
  get teamEmpty(): boolean {
    return this.$store.state.pointingSession.currentSession.participants.length === 0
  }

  get currentUsers(): Array<User> {
    return this.$store.state.pointingSession.currentSession.participants
  }

  get isSessionReady(): boolean {
    return this.$store.state.pointingSession.sessionActive
  }

  get votesShown(): boolean {
    return this.$store.state.pointingSession.currentSession.votesShown
  }

  get userURL(): string {
    let port = ''
    if (window.location.protocol === 'http:' && window.location.port !== '80') {
      port = `:${window.location.port}`
    }
    return `${window.location.protocol}//${window.location.hostname}${port}/session/${this.$store.state.pointingSession.currentSession.sessionId}`
  }

  mounted() {
    this.routeParamsChanged()
  }

  @Watch('$route')
  routeParamsChanged() {
    const sessionId = this.$route.params.sessionId
    const facilitatorSessionKey = this.$route.params.facilitatorSessionKey
    const markActive = true
    this.$store.dispatch(PointingSessionStore.ACTION_LOAD_FACILITATOR_SESSION, { sessionId, facilitatorSessionKey, markActive })
  }
}
</script>

<style lang="scss">

</style>
