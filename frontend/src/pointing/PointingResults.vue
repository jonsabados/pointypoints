<template>
  <div class="container-fluid" role="application">
    <h4>Results</h4>
    <table class="table table-striped table-bordered">
      <thead>
      <tr>
        <th scope="col">Participant</th>
        <th scope="col">Vote</th>
      </tr>
      </thead>
      <tbody>
      <tr v-for="user in users" :key="user.userId">
        <td><user-display-name :user="user" /></td>
        <td>
          <div v-if="user.currentVote">
            {{ user.currentVote }}
          </div>
          <div v-else>
            -
          </div>
        </td>
      </tr>
      </tbody>
    </table>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { PointingSession } from './PointingSessionStore'
import UserDisplayName from '../user/UserDisplayName.vue'
import { User } from '@/user/user'

@Component({
  components: { UserDisplayName },
  props: {
    session: Object
  }
})
export default class PointingResults extends Vue {
  session?: PointingSession

  get users():Array<User> {
    if (!this.session) {
      return []
    }
    const ret = this.session.participants
    if (this.session.facilitatorPoints) {
      ret.push(this.session.facilitator)
    }
    return ret
  }
}
</script>

<style lang="scss">

</style>
