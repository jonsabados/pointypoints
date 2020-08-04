<template>
  <main class="container-fluid" role="main">
    <h1>New Pointing Session</h1>
    <p>Once the session has been started you will be provided a link which participants may use to join.</p>
    <form @submit.prevent="startSession">
      <div class="form-group">
        <label for="facilitatorName">Facilitator Name:</label>
        <input type="text" class="form-control" id="facilitatorName" aria-describedby="facilitatorHelp" placeholder="Jane Doe" v-model="facilitatorName" />
        <small id="facilitatorHelp" class="form-text text-muted">Name of the individual running the session. Required.</small>
      </div>
      <div class="form-group">
        <label for="facilitatorHandle">Facilitator Handle:</label>
        <input type="text" class="form-control" id="facilitatorHandle" aria-describedby="facilitatorHandleHelp" placeholder="PointMaster2020" v-model="facilitatorHandle" />
        <small id="facilitatorHandleHelp" class="form-text text-muted">If specified this will be the display name of the facilitator, otherwise the value for Facilitator Name will be displayed.</small>
      </div>
      <div class="form-check">
        <input type="radio" name="facilitatorPointing" class="form-check-input" id="facilitatorPointingNo" aria-describedby="facilitatorPointingNoHelp" value="false" v-model="facilitatorParticipating" />
        <label class="form-check-label" for="facilitatorPointingNo">Facilitator will not be pointing</label>
        <small id="facilitatorPointingNoHelp" class="form-text text-muted">When selected the facilitator will only control when votes are shown and cleared.</small>
      </div>
      <div class="form-check">
        <input type="radio" name="facilitatorPointing" class="form-check-input" id="facilitatorPointingYes" aria-describedby="facilitatorPointingYesHelp" value="true" v-model="facilitatorParticipating"/>
        <label class="form-check-label" for="facilitatorPointingYes">Facilitator will be pointing</label>
        <small id="facilitatorPointingYesHelp" class="form-text text-muted">When selected the facilitator will also have the option to point issues along with the ability to control when votes are shown and cleared.</small>
      </div>
      <button type="submit" class="btn btn-primary" :disabled="disableSubmit" id="startSessionButton">Start Session</button>
    </form>
  </main>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'
import { PointingSessionStore } from '@/pointing/PointingSessionStore'

@Component
export default class NewSession extends Vue {
  facilitatorName: string = ''

  facilitatorHandle: string = ''

  facilitatorParticipating: boolean = false

  startSession() {
    const facilitatorName = this.facilitatorName
    const facilitatorHandle = this.facilitatorHandle
    const facilitatorParticipating = this.facilitatorParticipating
    this.$store.dispatch(PointingSessionStore.ACTION_BEGIN_SESSION, {
      facilitator: {
        name: facilitatorName,
        handle: facilitatorHandle
      },
      facilitatorParticipating: facilitatorParticipating
    })
  }

  get disableSubmit(): boolean {
    return this.facilitatorName === ''
  }
}
</script>

<style lang="scss">

</style>
