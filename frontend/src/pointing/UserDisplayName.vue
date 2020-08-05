<template>
  <span v-if="hasHandle">{{ user.handle }}</span>
  <span v-else>{{ user.name }}</span>
</template>

<script lang="ts">
import { Component, Prop, Vue } from 'vue-property-decorator'
import { User } from '@/pointing/PointingSessionStore'

@Component({
  props: {
    user: Object
  }
})
export default class UserDisplayName extends Vue {
  user: User | undefined

  get hasHandle() {
    return this.user && this.user.handle
  }
}
</script>
