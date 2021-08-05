<template>
  <div class="container-fluid" role="application">
    <div v-if="loading">
      <loading/>
    </div>
    <div v-else>
      <h4 v-if="currentVote">Your current vote is {{ currentVote }}</h4>
      <h4 v-else>Please Vote</h4>
      <div class="voteButtons">
        <button type="button" class="btn btn-primary" v-on:click="vote('0')">0</button>
        <button type="button" class="btn btn-primary" v-on:click="vote('.5')">½</button>
        <button type="button" class="btn btn-primary" v-on:click="vote('1')">1</button>
        <button type="button" class="btn btn-primary" v-on:click="vote('2')">2</button>
        <button type="button" class="btn btn-primary" v-on:click="vote('3')">3</button>
        <button type="button" class="btn btn-primary" v-on:click="vote('5')">5</button>
        <button type="button" class="btn btn-primary" v-on:click="vote('8')">8</button>
        <button type="button" class="btn btn-primary" v-on:click="vote('13')">13</button>
        <button type="button" class="btn btn-primary" v-on:click="vote('21')">21</button>
        <button type="button" class="btn btn-primary" v-on:click="vote('∞')">∞</button>
        <button type="button" class="btn btn-primary" v-on:click="vote('')">Clear Vote</button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { PointingSession } from './PointingSessionStore'
import { AppStore } from '@/app/AppStore'
import { User } from '@/user/user'
import Loading from '@/app/Loading.vue'
import { vote } from '@/pointing/pointing'

@Component({
  components: { Loading },
  props: {
    session: Object,
    userId: String
  }
})
export default class Pointing extends Vue {
  session?: PointingSession
  userId?: string
  loading: boolean = false

  get user(): User | undefined {
    if (!this.session) {
      return undefined
    }
    if (this.session.facilitator.userId === this.userId) {
      return this.session.facilitator
    }
    return this.session.participants.find((u) => {
      return u.userId === this.userId
    })
  }

  get currentVote(): string {
    if (!this.user || !this.user.currentVote) {
      return ''
    }
    return this.user.currentVote
  }

  async vote(value: string) {
    if (!this.session) {
      throw new Error('attempt to vote without a session')
    }
    if (!this.userId || !this.user) {
      throw new Error('attempt to vote without a user id')
    }
    this.loading = true
    try {
      await vote(this.$store.state.profile.authToken, this.session.sessionId, this.userId, value)
      this.user.currentVote = value
    } catch (e) {
      await this.$store.dispatch(AppStore.ACTION_REGISTER_REMOTE_ERROR, e)
    }
    this.loading = false
  }
}
</script>

<style lang="scss">
.voteButtons button {
  margin: .25em;
}
</style>
