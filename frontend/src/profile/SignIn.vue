<template>
  <div v-if="isReady" id="signInWidget">
    <google-login v-if="!isSignedIn" :params="googleLoginParams">Sign In With Google</google-login>
    <google-login v-else :params="googleLoginParams" :logoutButton="true">Logout</google-login>
  </div>
</template>

<script lang="ts">
// @ts-ignore
import GoogleLogin from 'vue-google-login'
import { Component, Vue } from 'vue-property-decorator'

@Component({
  components: {
    GoogleLogin
  }
})
export default class SignIn extends Vue {
  get isReady(): boolean {
    return this.$store.state.profile.isReady
  }

  get isSignedIn(): boolean {
    return this.$store.state.profile.signedIn
  }

  googleLoginParams = {
    client_id: process.env.VUE_APP_GOOGLE_OAUTH_CLIENT_ID
  }
}
</script>

<style lang="scss">
@import "~bootstrap/scss/bootstrap";

#signInWidget  {
  display: inline;
  white-space: nowrap;
  button {
    border: none;
    background: none;
    color: $navbar-light-color;
    font-weight: bold;
  }
  button:hover {
    color: $navbar-light-hover-color;
  }
}
</style>
