<template>
  <main class="container-fluid" role="main">
    <loading v-if="!profile" />
    <div v-else>
      {{ profile }}
    </div>
  </main>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { Profile } from '@/pointing/pointing'
import Loading from '@/app/Loading.vue'

@Component({
  components: { Loading }
})
export default class Pointing extends Vue {
  get profile(): Profile | undefined {
    return this.$store.state.profile.profile
  }
}
</script>
