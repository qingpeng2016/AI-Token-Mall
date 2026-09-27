<script setup lang="ts">
import { onMounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { getSessionUser, isLoggedIn, userAccountLabel } from '@/composables/useSessionUser'

const router = useRouter()
const user = getSessionUser()

onMounted(() => {
  if (!isLoggedIn()) {
    router.replace({ path: '/login', query: { redirect: '/member' } })
  }
})
</script>

<template>
  <div v-if="user" class="member-page atm-container">
    <h1>会员中心</h1>
    <p class="member-account">当前账号：{{ userAccountLabel(user) }}</p>
    <p class="member-hint">套餐额度、订单与开票功能开发中。</p>
    <RouterLink to="/#catalog" class="atm-btn-primary">继续选购套餐</RouterLink>
  </div>
</template>

<style scoped>
.member-page {
  padding: 48px 24px 64px;
}

.member-page h1 {
  margin: 0 0 12px;
  font-size: 28px;
  font-weight: 800;
  color: var(--atm-text);
}

.member-account {
  margin: 0 0 8px;
  font-size: 15px;
  color: var(--atm-text);
}

.member-hint {
  margin: 0 0 24px;
  font-size: 14px;
  color: var(--atm-text-muted);
}
</style>
