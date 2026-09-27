<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>

<template>
  <el-container class="layout">
    <el-header height="56px" class="header">
      <RouterLink to="/" class="brand">AI Token Mall</RouterLink>
      <nav>
        <RouterLink to="/login">登录</RouterLink>
        <RouterLink to="/register">注册</RouterLink>
      </nav>
    </el-header>
    <el-main>
      <RouterView />
    </el-main>
  </el-container>
</template>

<style scoped>
.layout {
  min-height: 100vh;
}
.header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--el-border-color);
}
.brand {
  font-weight: 600;
  font-size: 18px;
  color: inherit;
  text-decoration: none;
}
nav a {
  margin-left: 16px;
  color: var(--el-color-primary);
  text-decoration: none;
}
</style>
