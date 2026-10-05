<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div class="workbench-layout">
    <RouterView />
  </div>
</template>

<style scoped>
.workbench-layout {
  min-height: 100vh;
  background: var(--atm-bg, #f5f3ff);
}
</style>
