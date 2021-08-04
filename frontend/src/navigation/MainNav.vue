<template>
  <nav role="navigation" class="navbar navbar-expand-lg navbar-light bg-light">
    <ul class="navbar-nav mr-auto">
      <li class="nav-item">
        <router-link :to="{name: 'home'}" class="nav-link">Home</router-link>
      </li>
      <li class="nav-item">
        <router-link :to="{name: 'newSession'}" class="nav-link">Start New Pointing Session</router-link>
      </li>
      <li class="nav-item" v-if="isSignedIn">
        <router-link :to="{name: 'profile'}" class="nav-link">Profile</router-link>
      </li>
    </ul>
    <div class="my-2 my-lg-0">
      <sign-in />
    </div>
  </nav>
</template>

<script lang="ts">
import SignIn from '@/profile/SignIn.vue'
import { Component, Vue } from 'vue-property-decorator'

@Component({
  components: {
    SignIn
  }
})
export default class MainNav extends Vue {
  get isSignedIn(): boolean {
    return this.$store.state.profile.signedIn
  }
}
</script>
