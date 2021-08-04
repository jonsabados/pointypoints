<template>
  <main class="container-fluid" role="main">
    <loading v-if="!isReady" />
    <div v-else>
      <h1>Profile</h1>
      <form @submit.prevent="saveProfile">
        <div class="form-group">
          <label for="email">Email Address:</label>
          <input type="text" class="form-control" id="email" aria-describedby="emailHelp" v-model="email" disabled="disabled" />
          <small id="emailHelp" class="form-text text-muted">Your email address. This is based on your credentials and cannot be changed.</small>
        </div>
        <div class="form-group">
          <label for="email">Name:</label>
          <input type="text" class="form-control" id="name" aria-describedby="nameHelp" v-model="name" disabled="disabled" />
          <small id="nameHelp" class="form-text text-muted">Your name. This is based on your credentials and cannot be changed.</small>
        </div>
        <div class="form-group">
          <label for="email">Display Name:</label>
          <input type="text" class="form-control" id="handle" aria-describedby="handleHelp" placeholder="PointMaster 2000" v-model="handle" />
          <small id="handleHelp" class="form-text text-muted">How you are displayed to other participants. Note: facilitators may always see your name.</small>
        </div>
        <button type="submit" class="btn btn-primary" :disabled="!changed" id="saveProfile">Save Changes</button>
      </form>
    </div>
  </main>
</template>

<script lang="ts">
import { Component, Vue, Watch } from 'vue-property-decorator'
import Loading from '@/app/Loading.vue'
import { HOME_ROUTE_NAME } from '@/navigation/router'
import { ProfileStore } from '@/profile/ProfileStore'

@Component({
  components: { Loading }
})
export default class Pointing extends Vue {
  handle: string = this.remoteHandle

  get isReady(): boolean {
    return this.$store.state.profile.isReady && this.$store.state.profile.remoteProfile
  }

  get email(): string {
    return this.$store.state.profile.remoteProfile.email
  }

  get name(): string {
    return this.$store.state.profile.remoteProfile.name
  }

  get changed(): boolean {
    return this.handle !== this.remoteHandle
  }

  get remoteHandle(): string {
    return this.$store.state.profile.remoteProfile && this.$store.state.profile.remoteProfile.handle ? this.$store.state.profile.remoteProfile.handle : ''
  }

  saveProfile() {
    this.$store.dispatch(ProfileStore.ACTION_UPDATE_PROFILE, {
      handle: this.handle
    })
  }

  @Watch('$store.state.profile.remoteProfile')
  setupProfileCopy() {
    this.handle = this.remoteHandle
  }

  @Watch('$store.state.profile.signedIn')
  watchForLogout() {
    if (!this.$store.state.profile.signedIn) {
      this.$router.push({ name: HOME_ROUTE_NAME })
    }
  }
}
</script>

<style lang="scss">

</style>
