<script setup lang="ts">
import Sidebar from './Sidebar.vue'
</script>

<template>
  <div class="layout">
    <Sidebar />
    <main class="main-content">
      <slot />
    </main>
  </div>
</template>