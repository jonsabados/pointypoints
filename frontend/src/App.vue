<template>
  <div id="app">
    <div>
      <router-view/>
    </div>
  </div>
</template>

<script lang="ts">
import { Vue, Component } from 'vue-property-decorator'

@Component({
  components: {
  }
})
export default class App extends Vue {

}
</script>

<style lang="scss">
$dark: #223e5d;
@import "~bootstrap/scss/bootstrap";
html {
  position: relative;
  min-height: 100%;
}
body {
  margin-bottom: 60px;
}
</style>
