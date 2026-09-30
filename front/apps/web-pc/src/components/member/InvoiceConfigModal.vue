<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { InvoiceConfigItem, SaveInvoiceConfigBody } from '@ai-token-mall/shared'
import { invoiceApi } from '@/api'
import {
  beginRouteNavigationLoading,
  endRouteNavigationLoading,
} from '@/composables/useRouteNavigationLoading'

const open = defineModel<boolean>('open', { default: false })

const configs = ref<InvoiceConfigItem[]>([])
const editingId = ref<number | null>(null)
const saving = ref(false)

const emptyForm = (): SaveInvoiceConfigBody & { profile_type: 'enterprise' | 'personal' } => ({
  profile_type: 'enterprise',
  title: '',
  tax_no: '',
  bank_name: '',
  bank_account: '',
  address: '',
  phone: '',
  is_default: true,
})

const form = reactive(emptyForm())

function resetForm() {
  Object.assign(form, emptyForm())
  editingId.value = null
}

async function loadConfigs() {
  beginRouteNavigationLoading()
  try {
    configs.value = await invoiceApi.listConfigs()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载发票抬头失败')
  } finally {
    endRouteNavigationLoading()
  }
}

function startCreate() {
  resetForm()
  if (!configs.value.length) {
    form.is_default = true
  }
}

function startEdit(row: InvoiceConfigItem) {
  editingId.value = row.id
  form.profile_type = row.profile_type === 'personal' ? 'personal' : 'enterprise'
  form.title = row.title
  form.tax_no = row.tax_no ?? ''
  form.bank_name = row.bank_name ?? ''
  form.bank_account = row.bank_account ?? ''
  form.address = row.address ?? ''
  form.phone = row.phone ?? ''
  form.is_default = row.is_default
}

function bodyFromForm(): SaveInvoiceConfigBody {
  return {
    profile_type: form.profile_type,
    title: form.title.trim(),
    tax_no: form.tax_no?.trim() || undefined,
    bank_name: form.bank_name?.trim() || undefined,
    bank_account: form.bank_account?.trim() || undefined,
    address: form.address?.trim() || undefined,
    phone: form.phone?.trim() || undefined,
    is_default: form.is_default,
  }
}

async function save() {
  if (!form.title.trim()) {
    ElMessage.warning('请填写发票抬头')
    return
  }
  if (form.profile_type === 'enterprise' && !form.tax_no?.trim()) {
    ElMessage.warning('企业抬头请填写税号')
    return
  }
  saving.value = true
  beginRouteNavigationLoading()
  try {
    const body = bodyFromForm()
    if (editingId.value) {
      await invoiceApi.updateConfig(editingId.value, body)
      ElMessage.success('已保存')
    } else {
      await invoiceApi.createConfig(body)
      ElMessage.success('已添加')
    }
    await loadConfigs()
    resetForm()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
    endRouteNavigationLoading()
  }
}

function close() {
  open.value = false
}

watch(open, (v) => {
  if (v) {
    void loadConfigs()
    resetForm()
  }
})
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="icm-backdrop" @click.self="close">
      <div class="icm-panel" role="dialog" aria-labelledby="invoice-config-title">
        <header class="icm-head">
          <h3 id="invoice-config-title">发票抬头</h3>
          <button type="button" class="icm-close" aria-label="关闭" @click="close">×</button>
        </header>

        <div class="icm-toolbar">
          <button type="button" class="atm-btn-primary btn-xs" @click="startCreate">新增抬头</button>
        </div>

        <div v-if="configs.length" class="icm-table-wrap">
          <table class="icm-table">
            <thead>
              <tr>
                <th>抬头</th>
                <th>税号</th>
                <th>默认</th>
                <th />
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in configs" :key="row.id">
                <td>{{ row.title }}</td>
                <td class="icm-mono">{{ row.tax_no || '—' }}</td>
                <td>{{ row.is_default ? '是' : '—' }}</td>
                <td>
                  <button type="button" class="icm-link" @click="startEdit(row)">编辑</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else class="icm-hint">暂无抬头，请下方填写并保存。</p>

        <form class="icm-form" @submit.prevent="save">
          <p class="icm-label icm-form-title">
            {{ editingId ? '编辑抬头' : '新增抬头' }}
          </p>
          <div class="icm-form-grid">
            <label class="icm-field">
              <span class="icm-label">类型</span>
              <select v-model="form.profile_type" class="icm-input">
                <option value="enterprise">企业</option>
                <option value="personal">个人</option>
              </select>
            </label>
            <label class="icm-field">
              <span class="icm-label">税号</span>
              <input v-model="form.tax_no" class="icm-input" type="text" maxlength="64" />
            </label>
            <label class="icm-field icm-field--full">
              <span class="icm-label">发票抬头</span>
              <input v-model="form.title" class="icm-input" type="text" maxlength="256" />
            </label>
            <label class="icm-field">
              <span class="icm-label">开户银行</span>
              <input v-model="form.bank_name" class="icm-input" type="text" maxlength="128" />
            </label>
            <label class="icm-field">
              <span class="icm-label">银行账号</span>
              <input v-model="form.bank_account" class="icm-input" type="text" maxlength="64" />
            </label>
            <label class="icm-field">
              <span class="icm-label">地址</span>
              <input v-model="form.address" class="icm-input" type="text" maxlength="512" />
            </label>
            <label class="icm-field">
              <span class="icm-label">电话</span>
              <input v-model="form.phone" class="icm-input" type="text" maxlength="32" />
            </label>
          </div>
          <label class="icm-check">
            <input v-model="form.is_default" type="checkbox" />
            设为默认抬头
          </label>
          <div class="icm-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="close">关闭</button>
            <button type="submit" class="atm-btn-primary btn-xs" :disabled="saving">
              {{ saving ? '保存中…' : '保存' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.icm-backdrop {
  position: fixed;
  inset: 0;
  z-index: 2050;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px 16px;
  background: rgba(15, 10, 40, 0.52);
  backdrop-filter: blur(4px);
}

.icm-panel {
  width: 100%;
  max-width: 640px;
  max-height: min(90vh, 720px);
  overflow-y: auto;
  padding: 22px 24px 24px;
  background: #fff;
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(30, 27, 75, 0.22);
}

.icm-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.icm-head h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 800;
  color: var(--atm-text);
}

.icm-close {
  width: 36px;
  height: 36px;
  font-size: 22px;
  line-height: 1;
  color: var(--atm-text-muted);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
}

.icm-toolbar {
  margin-bottom: 12px;
}

.icm-hint {
  margin: 0 0 8px;
  font-size: 14px;
  color: var(--atm-text-muted);
}

.icm-table-wrap {
  overflow-x: auto;
}

.icm-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.icm-table th {
  padding: 10px 12px;
  text-align: left;
  font-weight: 600;
  color: var(--atm-text-muted);
  border-bottom: 1px solid #e2e8f0;
}

.icm-table td {
  padding: 10px 12px;
  border-bottom: 1px solid #f1f5f9;
}

.icm-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: var(--atm-text-muted);
}

.icm-link {
  padding: 0;
  font-size: 13px;
  color: var(--atm-primary, #7c3aed);
  background: none;
  border: none;
  cursor: pointer;
}

.icm-link:hover {
  text-decoration: underline;
}

.icm-form {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.icm-form-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px 16px;
}

@media (max-width: 560px) {
  .icm-form-grid {
    grid-template-columns: 1fr;
  }

  .icm-field--full {
    grid-column: auto;
  }
}

.icm-field--full {
  grid-column: 1 / -1;
}

.icm-form-title {
  margin: 0;
  font-weight: 600;
}

.icm-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.icm-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-text-muted);
}

.icm-input {
  width: 100%;
  box-sizing: border-box;
  padding: 10px 12px;
  font-size: 13px;
  color: var(--atm-text);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.icm-check {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--atm-text);
}

.icm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 4px;
}
</style>
