<script setup lang="ts">
import { reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import AuthShell from '@/components/auth/AuthShell.vue'
import { userApi } from '@/api'

const router = useRouter()
const loading = ref(false)
const form = reactive({
  email: '',
  phone: '',
  password: '',
  nickname: '',
  mode: 'email' as 'email' | 'phone',
})

async function onSubmit() {
  loading.value = true
  try {
    const body =
      form.mode === 'email'
        ? { email: form.email, password: form.password, nickname: form.nickname || undefined }
        : { phone: form.phone, password: form.password, nickname: form.nickname || undefined }
    await userApi.register(body)
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <AuthShell title="创建账号" subtitle="一分钟完成注册，即可选购套餐。">
    <div class="auth-mode-pills">
      <button
        type="button"
        class="auth-mode-pill"
        :class="{ active: form.mode === 'email' }"
        @click="form.mode = 'email'"
      >
        邮箱注册
      </button>
      <button
        type="button"
        class="auth-mode-pill"
        :class="{ active: form.mode === 'phone' }"
        @click="form.mode = 'phone'"
      >
        手机注册
      </button>
    </div>

    <form class="auth-form" @submit.prevent="onSubmit">
      <label v-if="form.mode === 'email'" class="auth-field">
        <span class="auth-field-label">邮箱</span>
        <el-input
          v-model="form.email"
          type="email"
          placeholder="you@example.com"
          size="large"
        />
      </label>
      <label v-else class="auth-field">
        <span class="auth-field-label">手机号</span>
        <el-input v-model="form.phone" placeholder="11 位手机号" size="large" />
      </label>

      <label class="auth-field">
        <span class="auth-field-label"
          >昵称<span class="auth-field-optional">选填</span></span
        >
        <el-input v-model="form.nickname" placeholder="显示在会员中心" size="large" />
      </label>

      <label class="auth-field">
        <span class="auth-field-label">密码</span>
        <el-input
          v-model="form.password"
          type="password"
          placeholder="至少 8 位"
          show-password
          size="large"
        />
      </label>

      <button type="submit" class="auth-submit" :disabled="loading">
        {{ loading ? '提交中…' : '免费注册' }}
      </button>
    </form>

    <p class="auth-foot">
      已有账号？
      <RouterLink to="/login">直接登录</RouterLink>
    </p>
  </AuthShell>
</template>
