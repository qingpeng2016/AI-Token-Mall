<script setup lang="ts">
import { computed, reactive, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  DEFAULT_TOPIC_DISCOVERY,
  PAPER_MODULES,
  type PaperModuleId,
  type TopicDiscoveryForm,
} from '@paper/types'

const activeModule = ref<PaperModuleId>('topic-discovery')
const running = ref(false)

const topicForm = reactive<TopicDiscoveryForm>({ ...DEFAULT_TOPIC_DISCOVERY })

const currentMeta = computed(
  () => PAPER_MODULES.find((m) => m.id === activeModule.value) ?? PAPER_MODULES[0],
)

const intensityOptions = [
  { value: 'fast', label: '更快' },
  { value: 'balanced', label: 'Balanced（平衡）' },
  { value: 'deep', label: '更深' },
] as const

const auditOptions = [
  { value: 'standard', label: 'Standard' },
  { value: 'polished', label: 'Polished（精修）' },
  { value: 'strict', label: 'Strict' },
] as const

async function onRun() {
  if (activeModule.value === 'topic-discovery' && !topicForm.direction.trim()) {
    ElMessage.warning('请填写研究方向')
    return
  }
  running.value = true
  try {
    await new Promise((r) => setTimeout(r, 600))
    ElMessage.success(`「${currentMeta.value.label}」工作流已启动（演示）`)
  } finally {
    running.value = false
  }
}
</script>

<template>
  <div class="paper-workbench">
    <aside class="paper-sidebar">
      <div class="paper-sidebar-brand">
        <span class="paper-sidebar-logo" aria-hidden="true">◆</span>
        <div>
          <div class="paper-sidebar-title">Research Desktop</div>
          <div class="paper-sidebar-sub">Research workflows</div>
        </div>
      </div>

      <nav class="paper-nav" aria-label="科研工作流">
        <button
          v-for="mod in PAPER_MODULES"
          :key="mod.id"
          type="button"
          class="paper-nav-item"
          :class="{ 'paper-nav-item--active': activeModule === mod.id }"
          @click="activeModule = mod.id"
        >
          {{ mod.label }}
        </button>
      </nav>

      <RouterLink to="/member" class="paper-sidebar-back">← 返回会员中心</RouterLink>
    </aside>

    <main class="paper-main">
      <header class="paper-main-head">
        <div>
          <h1 class="paper-main-title">{{ currentMeta.label }}</h1>
          <p class="paper-main-desc">{{ currentMeta.description }}</p>
        </div>
        <button
          type="button"
          class="paper-run-btn"
          :disabled="running"
          @click="onRun"
        >
          <span class="paper-run-icon" aria-hidden="true">▶</span>
          {{ running ? '运行中…' : '运行' }}
        </button>
      </header>

      <div class="paper-gate" role="status">
        <span class="paper-gate-icon" aria-hidden="true">📖</span>
        <p>
          <strong>参考文献门禁已启用：</strong>
          未验证文献不能进入最终引用、BibTeX 或论文正文。
        </p>
      </div>

      <!-- 选题发现 -->
      <section v-if="activeModule === 'topic-discovery'" class="paper-panel">
        <h2 class="paper-panel-title">参数</h2>

        <label class="paper-field paper-field--block">
          <span class="paper-label">研究方向 <em class="req">*</em></span>
          <textarea
            v-model="topicForm.direction"
            class="paper-textarea"
            rows="4"
            placeholder="输入要探索的研究主题、问题或关键词。"
          />
        </label>

        <div class="paper-field-grid">
          <label class="paper-field">
            <span class="paper-label">目标会议/期刊</span>
            <input v-model="topicForm.venue" type="text" class="paper-input" />
            <span class="paper-hint">用于约束贡献类型、实验标准与写作风格</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">文献来源</span>
            <input v-model="topicForm.literatureSources" type="text" class="paper-input" />
            <span class="paper-hint">检索结果须进入真实文献校验流程</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">执行强度</span>
            <select v-model="topicForm.intensity" class="paper-select">
              <option v-for="o in intensityOptions" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
            <span class="paper-hint">控制检索数量、迭代轮数与输出深度</span>
          </label>

          <label class="paper-field">
            <span class="paper-label">审计等级</span>
            <select v-model="topicForm.auditLevel" class="paper-select">
              <option v-for="o in auditOptions" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
            <span class="paper-hint">citation / claim / kill argument 门禁强度</span>
          </label>
        </div>

        <label class="paper-check">
          <input v-model="topicForm.humanCheckpoint" type="checkbox" />
          <span>
            <strong>人工检查点</strong>
            <span class="paper-hint paper-hint--inline">在关键阶段暂停，等待人工确认后再继续</span>
          </span>
        </label>
      </section>

      <!-- 其他模块占位（结构一致，后续接 Agent） -->
      <section v-else class="paper-panel paper-panel--placeholder">
        <p class="paper-placeholder-lead">
          「{{ currentMeta.label }}」配置与工作流与 ARIS Desktop 对齐，后端 Agent 接入后在此展示表单与进度。
        </p>
        <ul class="paper-placeholder-list">
          <li>遵守全局参考文献门禁与审计策略</li>
          <li>支持人工检查点（可在环境配置中设为默认）</li>
          <li>运行结果将写入本地工作区 / Research Wiki（规划中）</li>
        </ul>
      </section>
    </main>
  </div>
</template>

<style scoped>
.paper-workbench {
  display: flex;
  min-height: 100vh;
  background: var(--atm-bg, #f5f3ff);
}

.paper-sidebar {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  width: 220px;
  padding: 20px 12px 16px;
  color: #fff;
  background: linear-gradient(180deg, #1e1b4b 0%, #5b21b6 55%, #6366f1 100%);
}

.paper-sidebar-brand {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  padding: 4px 10px 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.12);
}

.paper-sidebar-logo {
  font-size: 18px;
  color: #c4b5fd;
}

.paper-sidebar-title {
  font-size: 15px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.paper-sidebar-sub {
  margin-top: 2px;
  font-size: 11px;
  opacity: 0.75;
}

.paper-nav {
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 4px;
  margin-top: 16px;
}

.paper-nav-item {
  padding: 10px 12px;
  font-size: 14px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.88);
  text-align: left;
  background: transparent;
  border: none;
  border-radius: 10px;
  cursor: pointer;
  transition: background 0.12s;
}

.paper-nav-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.paper-nav-item--active {
  color: #fff;
  background: rgba(255, 255, 255, 0.18);
  box-shadow: inset 0 0 0 1px rgba(255, 255, 255, 0.12);
}

.paper-sidebar-back {
  margin-top: 12px;
  padding: 12px 12px;
  font-size: 15px;
  font-weight: 500;
  color: #fff;
  text-decoration: none;
  border-radius: 10px;
}

.paper-sidebar-back:hover {
  background: rgba(255, 255, 255, 0.1);
}

.paper-main {
  flex: 1;
  min-width: 0;
  padding: 28px 32px 40px;
}

.paper-main-head {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px 24px;
  margin-bottom: 20px;
}

.paper-main-title {
  margin: 0;
  font-size: 26px;
  font-weight: 800;
  color: var(--atm-text, #1e1b4b);
  letter-spacing: -0.03em;
}

.paper-main-desc {
  margin: 8px 0 0;
  max-width: 640px;
  font-size: 14px;
  line-height: 1.55;
  color: var(--atm-text-muted, #64748b);
}

.paper-run-btn {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  padding: 12px 22px;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  background: var(--atm-gradient, linear-gradient(135deg, #7c3aed, #6366f1));
  border: none;
  border-radius: 12px;
  cursor: pointer;
  box-shadow: 0 8px 24px rgba(91, 33, 182, 0.28);
  transition:
    transform 0.12s,
    opacity 0.12s;
}

.paper-run-btn:hover:not(:disabled) {
  transform: translateY(-1px);
}

.paper-run-btn:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.paper-run-icon {
  font-size: 12px;
}

.paper-gate {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  margin-bottom: 22px;
  padding: 14px 16px;
  font-size: 14px;
  line-height: 1.5;
  color: #713f12;
  background: linear-gradient(90deg, #fef9c3 0%, #fef3c7 100%);
  border: 1px solid #fde68a;
  border-radius: 12px;
}

.paper-gate-icon {
  flex-shrink: 0;
  font-size: 18px;
}

.paper-gate p {
  margin: 0;
}

.paper-panel {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.paper-panel-title {
  margin: 0 0 20px;
  font-size: 16px;
  font-weight: 700;
  color: var(--atm-text, #1e1b4b);
}

.paper-field-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px 20px;
  margin-top: 18px;
}

@media (max-width: 900px) {
  .paper-field-grid {
    grid-template-columns: 1fr;
  }

  .paper-sidebar {
    width: 100%;
  }

  .paper-workbench {
    flex-direction: column;
  }
}

.paper-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.paper-field--block {
  margin-bottom: 4px;
}

.paper-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-text, #1e1b4b);
}

.paper-label .req {
  color: #e53935;
  font-style: normal;
}

.paper-input,
.paper-select,
.paper-textarea {
  width: 100%;
  padding: 10px 12px;
  font-size: 14px;
  color: var(--atm-text, #1e1b4b);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  transition: border-color 0.12s;
}

.paper-textarea {
  resize: vertical;
  min-height: 100px;
  font-family: inherit;
  line-height: 1.5;
}

.paper-input:focus,
.paper-select:focus,
.paper-textarea:focus {
  outline: none;
  border-color: var(--atm-primary, #7c3aed);
  box-shadow: 0 0 0 3px rgba(124, 58, 237, 0.12);
}

.paper-hint {
  font-size: 12px;
  line-height: 1.45;
  color: var(--atm-text-muted, #64748b);
}

.paper-hint--inline {
  display: block;
  margin-top: 2px;
  font-weight: 400;
}

.paper-check {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  margin-top: 22px;
  font-size: 14px;
  color: var(--atm-text, #1e1b4b);
  cursor: pointer;
}

.paper-check input {
  margin-top: 3px;
  width: 16px;
  height: 16px;
  accent-color: var(--atm-primary, #7c3aed);
}

.paper-panel--placeholder {
  color: var(--atm-text-muted, #64748b);
}

.paper-placeholder-lead {
  margin: 0 0 16px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--atm-text, #1e1b4b);
}

.paper-placeholder-list {
  margin: 0;
  padding-left: 20px;
  font-size: 14px;
  line-height: 1.7;
}
</style>
