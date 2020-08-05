<template>
  <main class="container-fluid" role="main">
    <h1>Facilitating Session</h1>
    <div v-if="isSessionReady">

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

@Component({
  components: { Loading }
})
export default class Session extends Vue {
  get isSessionReady(): boolean {
    return this.$store.state.pointingSession.sessionActive
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
