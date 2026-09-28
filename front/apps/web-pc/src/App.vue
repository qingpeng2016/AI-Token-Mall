<script setup lang="ts">
import { useRouteNavigationLoading } from '@/composables/useRouteNavigationLoading'

const { active: routeLoading } = useRouteNavigationLoading()
</script>

<template>
  <div v-if="routeLoading" class="route-loading" aria-live="polite" aria-busy="true">
    <div class="route-loading-bar" />
  </div>
  <router-view />
</template>

<style scoped>
.route-loading {
  position: fixed;
  inset: 0;
  z-index: 9999;
  pointer-events: none;
  background: rgba(248, 250, 252, 0.35);
}
.route-loading-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: linear-gradient(90deg, transparent, var(--atm-primary, #7c3aed), transparent);
  animation: route-loading-slide 0.9s ease-in-out infinite;
}
@keyframes route-loading-slide {
  0% {
    transform: translateX(-100%);
  }
  100% {
    transform: translateX(100%);
  }
}
</style>
