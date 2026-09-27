<script setup lang="ts">
import { RouterView } from 'vue-router'
import FloatingCustomerService from '@/components/layout/FloatingCustomerService.vue'
import SiteFooter from '@/components/layout/SiteFooter.vue'
import SiteHeader from '@/components/layout/SiteHeader.vue'
</script>

<template>
  <div class="layout">
    <SiteHeader />
    <main class="main">
      <RouterView />
    </main>
    <SiteFooter />
    <FloatingCustomerService />
  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.main {
  flex: 1;
  background: var(--atm-bg);
}
</style>
