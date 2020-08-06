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
import { Component, Vue, Watch } from 'vue-property-decorator'
import { PointingSession, PointingSessionStore } from './PointingSessionStore'
import { User } from '@/user/user'
import Loading from '@/app/Loading.vue'

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
  newVote: string = ''
  votesShownState: boolean = false

  get loading(): boolean {
    return this.newVote !== this.currentVote
  }

  get user(): User | undefined {
    if (!this.session) {
      return undefined
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

  mounted() {
    this.votesShownState = this.session !== undefined && this.session.votesShown
  }

  vote(value: string) {
    if (!this.session) {
      throw Error('attempt to vote without a session')
    }
    this.newVote = value
    this.$store.dispatch(PointingSessionStore.ACTION_VOTE, { sessionId: this.session.sessionId, vote: value })
  }

  @Watch('session.votesShown')
  watchForClearVotes() {
    if (this.votesShownState && this.session && !this.session.votesShown) {
      this.newVote = ''
    }
    this.votesShownState = (this.session !== undefined && this.session.votesShown)
  }
}
</script>

<style lang="scss">
.voteButtons button {
  margin: .25em;
}
</style>
