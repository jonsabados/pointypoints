<template>
  <nav role="navigation" class="navbar navbar-expand-lg navbar-light bg-light">
    <ul class="navbar-nav mr-auto">
      <li class="nav-item">
        <router-link :to="{name: 'home'}" class="nav-link">Home</router-link>
      </li>
      <li class="nav-item">
        <router-link :to="{name: 'newSession'}" class="nav-link">Start New Pointing Session</router-link>
      </li>
    </ul>
  </nav>
</template>
