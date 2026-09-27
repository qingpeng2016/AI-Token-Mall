<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { userApi } from '@/api'

const router = useRouter()
const loading = ref(false)
const form = reactive({
  email: '',
  phone: '',
  password: '',
  mode: 'email' as 'email' | 'phone',
})

async function onSubmit() {
  loading.value = true
  try {
    const body =
      form.mode === 'email'
        ? { email: form.email, password: form.password }
        : { phone: form.phone, password: form.password }
    const res = await userApi.login(body)
    localStorage.setItem('atm_user', JSON.stringify(res.data))
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <el-card class="card" header="登录">
    <el-radio-group v-model="form.mode" style="margin-bottom: 16px">
      <el-radio-button value="email">邮箱</el-radio-button>
      <el-radio-button value="phone">手机</el-radio-button>
    </el-radio-group>
    <el-form label-width="80px" @submit.prevent="onSubmit">
      <el-form-item v-if="form.mode === 'email'" label="邮箱">
        <el-input v-model="form.email" type="email" autocomplete="username" />
      </el-form-item>
      <el-form-item v-else label="手机">
        <el-input v-model="form.phone" autocomplete="tel" />
      </el-form-item>
      <el-form-item label="密码">
        <el-input v-model="form.password" type="password" autocomplete="current-password" show-password />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading">登录</el-button>
      </el-form-item>
    </el-form>
  </el-card>
</template>

<style scoped>
.card {
  max-width: 378px;
  margin: 48px auto;
}
</style>
