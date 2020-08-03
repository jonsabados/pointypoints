<template>
  <div id="app">
    <main-nav />
    <div>
      <router-view/>
    </div>
  </div>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'
import MainNav from './navigation/MainNav.vue'

@Component({
  components: {
    MainNav
  }
})
export default class App extends Vue {

}
</script>

<style lang="scss">
@import "~bootstrap/scss/bootstrap";
html {
  position: relative;
  min-height: 100%;
}
</style>
