<template>
  <span v-if="hasHandle">{{ user.handle }}</span>
  <span v-else>{{ user.name }}</span>
</template>

<script lang="ts">
import { Component, Vue } from 'vue-property-decorator'
import { User } from '@/user/user'

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
