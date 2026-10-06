<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  DEFAULT_ENV_PREFERENCE,
  DEFAULT_TOPIC_DISCOVERY,
  DEMO_PAPER_MANUSCRIPTS,
  DISCIPLINE_OPTIONS,
  LITERATURE_SOURCE_OPTIONS,
  PAPER_MODULES,
  TOPIC_DISCOVERY_FLOW_STEPS,
  type EnvironmentPreferenceForm,
  type PaperManuscriptItem,
  type PaperModuleId,
  type TopicCheckpointKey,
  type TopicDiscoveryForm,
  type TopicFlowStepStatus,
} from '@paper/types'
import PaperWorkflowPanels from './PaperWorkflowPanels.vue'

type TopicFlowStepRuntime = {
  stageCode: string
  label: string
  checkpointKey?: TopicCheckpointKey
  status: TopicFlowStepStatus
}

type TopicCheckpointView = {
  key: TopicCheckpointKey
  title: string
  lines: string[]
}

type TopicRunDemo = {
  status: 'idle' | 'running' | 'checkpoint' | 'completed' | 'failed'
  steps: TopicFlowStepRuntime[]
  checkpoint: TopicCheckpointView | null
}

function createIdleTopicRun(): TopicRunDemo {
  return {
    status: 'idle',
    steps: TOPIC_DISCOVERY_FLOW_STEPS.map((s) => ({
      stageCode: s.stageCode,
      label: s.label,
      checkpointKey: s.checkpointKey,
      status: 'pending' as TopicFlowStepStatus,
    })),
    checkpoint: null,
  }
}

const CHECKPOINT_COPY: Record<TopicCheckpointKey, Omit<TopicCheckpointView, 'key'>> = {
  ideas_ready: {
    title: '候选选题与新颖性结论',
    lines: [
      'Idea A：稀疏注意力 + 动态路由 — 新颖性：与 Static Sparse 系列有明确区分（中等风险）',
      'Idea B：层级 KV 压缩 — 新颖性：需加强与 H2O / SnapKV 对比（偏高风险）',
      'Idea C：训练无关的 token 合并 — 新颖性：检索命中较少（低风险，建议深检索）',
    ],
  },
  plan_ready: {
    title: '实验计划草案',
    lines: [
      '主实验：LongBench + 自建 32k 上下文任务；对比 Full、StreamingLLM、代表稀疏法',
      '消融：路由粒度 / 预算 B / 是否重排 KV',
      '指标：准确率、吞吐、显存峰值；报告 3 种子均值±方差',
    ],
  },
}

const ENV_STORAGE_KEY = 'atm:paper:environment:v3'
const MANUSCRIPTS_STORAGE_KEY = 'atm:paper:manuscripts:v1'
const LEGACY_PROJECTS_STORAGE_KEY = 'atm:paper:projects:v1'

const activeModule = ref<PaperModuleId>('topic-discovery')
const running = ref(false)
const topicRunToken = ref(0)
const topicRunsByManuscript = ref<Record<string, TopicRunDemo>>({})
const workflowPanelsRef = ref<InstanceType<typeof PaperWorkflowPanels> | null>(null)

const topicForm = reactive<TopicDiscoveryForm>({ ...DEFAULT_TOPIC_DISCOVERY })
const envPreference = reactive<EnvironmentPreferenceForm>({ ...DEFAULT_ENV_PREFERENCE })

const manuscripts = ref<PaperManuscriptItem[]>([...DEMO_PAPER_MANUSCRIPTS])
const activeManuscriptId = ref<string>(DEMO_PAPER_MANUSCRIPTS[0]?.id ?? '')

function loadEnvFromStorage() {
  try {
    const raw = localStorage.getItem(ENV_STORAGE_KEY)
    if (!raw) return
    const data = JSON.parse(raw) as { preference?: EnvironmentPreferenceForm }
    if (data.preference) Object.assign(envPreference, data.preference)
  } catch {
    /* ignore */
  }
}

function loadManuscriptsFromStorage() {
  try {
    let raw = localStorage.getItem(MANUSCRIPTS_STORAGE_KEY)
    if (!raw) raw = localStorage.getItem(LEGACY_PROJECTS_STORAGE_KEY)
    if (!raw) return
    const data = JSON.parse(raw) as {
      manuscripts?: PaperManuscriptItem[]
      activeManuscriptId?: string
      projects?: PaperManuscriptItem[]
      activeProjectId?: string
    }
    const list = data.manuscripts ?? data.projects
    const activeId = data.activeManuscriptId ?? data.activeProjectId
    if (list?.length) manuscripts.value = list
    if (activeId && manuscripts.value.some((m) => m.id === activeId)) {
      activeManuscriptId.value = activeId
    }
  } catch {
    /* ignore */
  }
}

function persistManuscripts() {
  localStorage.setItem(
    MANUSCRIPTS_STORAGE_KEY,
    JSON.stringify({
      manuscripts: manuscripts.value,
      activeManuscriptId: activeManuscriptId.value,
    }),
  )
}

function onManuscriptChange(id: string) {
  activeManuscriptId.value = id
  persistManuscripts()
}

async function onCreateManuscript() {
  try {
    const { value } = await ElMessageBox.prompt(
      '每篇论文独立存放选题、各阶段运行与稿件产出。',
      '新建论文',
      {
        confirmButtonText: '创建',
        cancelButtonText: '取消',
        inputPlaceholder: '如：稀疏注意力 · NeurIPS 2026',
      },
    )
    const title = value?.trim()
    if (!title) return
    const item: PaperManuscriptItem = {
      id: `ms-${Date.now()}`,
      title,
      venueHint: '',
      status: 'active',
    }
    manuscripts.value = [item, ...manuscripts.value]
    activeManuscriptId.value = item.id
    persistManuscripts()
    ElMessage.success('论文已创建（演示：接入后将同步 paper_manuscript）')
  } catch {
    /* cancelled */
  }
}

loadEnvFromStorage()
loadManuscriptsFromStorage()

const activeManuscripts = computed(() => manuscripts.value.filter((m) => m.status === 'active'))

const currentManuscript = computed(
  () =>
    manuscripts.value.find((m) => m.id === activeManuscriptId.value) ?? activeManuscripts.value[0],
)

const isEnvironmentModule = computed(() => activeModule.value === 'environment')
const isTopicDiscoveryModule = computed(() => activeModule.value === 'topic-discovery')

const SECONDARY_WORKFLOW_MODULES = [
  'literature-review',
  'experiment-planning',
  'auto-review',
  'paper-writing',
  'figure-generation',
  'manuscript-analysis',
] as const

const isSecondaryWorkflowModule = computed(() =>
  (SECONDARY_WORKFLOW_MODULES as readonly string[]).includes(activeModule.value),
)

const currentTopicRun = computed((): TopicRunDemo => {
  const id = activeManuscriptId.value
  if (!id) return createIdleTopicRun()
  if (!topicRunsByManuscript.value[id]) {
    topicRunsByManuscript.value[id] = createIdleTopicRun()
  }
  return topicRunsByManuscript.value[id]
})

const topicRunVisible = computed(
  () => isTopicDiscoveryModule.value && currentTopicRun.value.status !== 'idle',
)

const topicRunBusy = computed(
  () => currentTopicRun.value.status === 'running' || running.value,
)

const topicPrimaryDisabled = computed(() => {
  if (isEnvironmentModule.value) return running.value
  if (isTopicDiscoveryModule.value) {
    return (
      topicRunBusy.value ||
      currentTopicRun.value.status === 'checkpoint' ||
      currentTopicRun.value.status === 'running'
    )
  }
  return running.value
})

const primaryActionLabel = computed(() => {
  if (isEnvironmentModule.value) {
    return running.value ? '保存中…' : '保存配置'
  }
  if (isTopicDiscoveryModule.value) {
    if (currentTopicRun.value.status === 'checkpoint') return '等待确认'
    if (currentTopicRun.value.status === 'running' || running.value) return '运行中…'
    if (currentTopicRun.value.status === 'completed') return '再次运行'
    return '运行'
  }
  return running.value ? '运行中…' : '运行'
})

function persistTopicRun(msId: string, patch: TopicRunDemo) {
  topicRunsByManuscript.value = { ...topicRunsByManuscript.value, [msId]: patch }
}

function delay(ms: number, token: number) {
  return new Promise<void>((resolve, reject) => {
    window.setTimeout(() => {
      if (token !== topicRunToken.value) reject(new Error('aborted'))
      else resolve()
    }, ms)
  })
}

function checkpointView(key: TopicCheckpointKey): TopicCheckpointView {
  const copy = CHECKPOINT_COPY[key]
  return { key, title: copy.title, lines: [...copy.lines] }
}

async function executeTopicFlow(fromIndex: number, token: number) {
  const msId = activeManuscriptId.value
  if (!msId) return

  const humanOn = topicForm.humanCheckpoint
  let run = { ...currentTopicRun.value, steps: currentTopicRun.value.steps.map((s) => ({ ...s })) }

  for (let i = fromIndex; i < run.steps.length; i++) {
    if (token !== topicRunToken.value) return

    const step = run.steps[i]
    step.status = 'running'
    run.status = 'running'
    run.checkpoint = null
    persistTopicRun(msId, run)

    const ms = step.stageCode === 'retrieve' ? 1400 : step.stageCode === 'audit' ? 900 : 1100
    try {
      await delay(ms, token)
    } catch {
      return
    }

    step.status = 'completed'
    persistTopicRun(msId, { ...run })

    const cpKey = step.checkpointKey
    if (humanOn && cpKey) {
      run.status = 'checkpoint'
      run.checkpoint = checkpointView(cpKey)
      persistTopicRun(msId, { ...run })
      ElMessage.info('流程已暂停：请确认检查点内容后再继续')
      return
    }
  }

  run.status = 'completed'
  run.checkpoint = null
  persistTopicRun(msId, run)
  ElMessage.success('选题发现流程已完成（演示）')
}

function resetTopicRunForAction(msId: string) {
  topicRunToken.value += 1
  persistTopicRun(msId, createIdleTopicRun())
}

async function startTopicDiscoveryRun() {
  const msId = activeManuscriptId.value
  if (!msId) return

  resetTopicRunForAction(msId)
  const token = topicRunToken.value
  running.value = true

  let run = createIdleTopicRun()
  run.status = 'running'
  persistTopicRun(msId, run)

  try {
    await executeTopicFlow(0, token)
  } finally {
    if (token === topicRunToken.value) running.value = false
  }
}

function continueTopicAfterCheckpoint() {
  const msId = activeManuscriptId.value
  if (!msId || currentTopicRun.value.status !== 'checkpoint') return

  const nextIndex = currentTopicRun.value.steps.findIndex((s) => s.status === 'pending')
  if (nextIndex < 0) return

  const token = topicRunToken.value
  running.value = true
  persistTopicRun(msId, {
    ...currentTopicRun.value,
    status: 'running',
    checkpoint: null,
    steps: currentTopicRun.value.steps.map((s) => ({ ...s })),
  })

  executeTopicFlow(nextIndex, token).finally(() => {
    if (token === topicRunToken.value) running.value = false
  })
}

function rejectTopicCheckpoint() {
  const msId = activeManuscriptId.value
  if (!msId) return
  topicRunToken.value += 1
  const run = {
    ...currentTopicRun.value,
    status: 'failed' as const,
    checkpoint: null,
    steps: currentTopicRun.value.steps.map((s) =>
      s.status === 'running' ? { ...s, status: 'failed' as TopicFlowStepStatus } : { ...s },
    ),
  }
  persistTopicRun(msId, run)
  running.value = false
  ElMessage.warning('已驳回：可修改参数后再次运行')
}

function cancelTopicRun() {
  const msId = activeManuscriptId.value
  if (!msId) return
  resetTopicRunForAction(msId)
  running.value = false
  ElMessage.info('已取消运行')
}

function topicStepIcon(status: TopicFlowStepStatus) {
  if (status === 'completed') return '✓'
  if (status === 'running') return '…'
  if (status === 'waiting_human') return '⏸'
  if (status === 'failed') return '✕'
  if (status === 'skipped') return '–'
  return '○'
}

watch(activeManuscriptId, () => {
  running.value = false
})

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

function toggleLiteratureSource(code: string, checked: boolean) {
  const set = new Set(envPreference.literatureSourceCodes)
  if (checked) set.add(code)
  else set.delete(code)
  envPreference.literatureSourceCodes = [...set]
}

function isLiteratureSourceChecked(code: string) {
  return envPreference.literatureSourceCodes.includes(code)
}

async function onPrimaryAction() {
  if (activeModule.value === 'topic-discovery' && !topicForm.direction.trim()) {
    ElMessage.warning('请填写研究方向')
    return
  }
  if (activeModule.value === 'environment') {
    if (envPreference.literatureSourceCodes.length === 0) {
      ElMessage.warning('请至少选择一个文献来源')
      return
    }
  }
  if (activeModule.value === 'topic-discovery') {
    await startTopicDiscoveryRun()
    return
  }

  if (isSecondaryWorkflowModule.value) {
    running.value = true
    try {
      const ok = await workflowPanelsRef.value?.runModule(activeModule.value)
      if (ok) {
        ElMessage.success(
          `「${currentMeta.value.label}」已完成演示运行 · 论文「${currentManuscript.value?.title ?? '未命名'}」`,
        )
      }
    } finally {
      running.value = false
    }
    return
  }

  running.value = true
  try {
    await new Promise((r) => setTimeout(r, 500))
    if (activeModule.value === 'environment') {
      const toStore = { preference: { ...envPreference } }
      localStorage.setItem(ENV_STORAGE_KEY, JSON.stringify(toStore))
      ElMessage.success('环境配置已保存（本地演示）')
      return
    }
    ElMessage.success(
      `「${currentMeta.value.label}」已在论文「${currentManuscript.value?.title ?? '未命名'}」下启动（演示）`,
    )
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

      <div class="paper-manuscript-switcher">
        <label class="paper-manuscript-label" for="paper-manuscript-select">当前论文</label>
        <select
          id="paper-manuscript-select"
          class="paper-manuscript-select"
          :value="activeManuscriptId"
          @change="onManuscriptChange(($event.target as HTMLSelectElement).value)"
        >
          <option v-for="m in activeManuscripts" :key="m.id" :value="m.id">
            {{ m.title }}{{ m.venueHint ? ` · ${m.venueHint}` : '' }}
          </option>
        </select>
        <button type="button" class="paper-manuscript-new" @click="onCreateManuscript">+ 新建论文</button>
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
          <p v-if="currentManuscript" class="paper-main-manuscript">
            当前论文：<strong>{{ currentManuscript.title }}</strong>
            <span v-if="currentManuscript.venueHint"> · {{ currentManuscript.venueHint }}</span>
          </p>
          <p class="paper-main-desc">{{ currentMeta.description }}</p>
        </div>
        <button
          type="button"
          class="paper-run-btn"
          :disabled="topicPrimaryDisabled"
          @click="onPrimaryAction"
        >
          <span class="paper-run-icon" aria-hidden="true">{{ isEnvironmentModule ? '✓' : '▶' }}</span>
          {{ primaryActionLabel }}
        </button>
      </header>

      <div v-if="!isEnvironmentModule" class="paper-gate" role="status">
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
          <input
            v-model="topicForm.humanCheckpoint"
            type="checkbox"
            :disabled="topicRunBusy || currentTopicRun.status === 'checkpoint'"
          />
          <span>
            <strong>人工检查点</strong>
            <span class="paper-hint paper-hint--inline">
              开启后会在 <em>2 个节点</em> 暂停：① 选题与新颖性完成后 ② 实验计划生成后；关闭则一口气跑完 5 个阶段
            </span>
          </span>
        </label>
      </section>

      <section
        v-if="activeModule === 'topic-discovery' && topicRunVisible"
        class="paper-panel paper-panel--flow"
      >
        <div class="paper-flow-head">
          <h2 class="paper-panel-title paper-panel-title--tight">运行进度</h2>
          <span class="paper-flow-badge" :class="`paper-flow-badge--${currentTopicRun.status}`">
            {{
              currentTopicRun.status === 'checkpoint'
                ? '等待人工确认'
                : currentTopicRun.status === 'running'
                  ? '执行中'
                  : currentTopicRun.status === 'completed'
                    ? '已完成'
                    : currentTopicRun.status === 'failed'
                      ? '已中止'
                      : ''
            }}
          </span>
          <button
            v-if="currentTopicRun.status === 'running' || currentTopicRun.status === 'checkpoint'"
            type="button"
            class="paper-flow-cancel"
            @click="cancelTopicRun"
          >
            取消
          </button>
        </div>
        <p class="paper-section-lead">
          对应 <code>paper_run</code> → <code>paper_run_stage</code>；暂停时写入
          <code>paper_run_checkpoint</code>（演示交互，非真实 Agent）。
        </p>

        <ol class="paper-flow-steps">
          <li
            v-for="step in currentTopicRun.steps"
            :key="step.stageCode"
            class="paper-flow-step"
            :class="`paper-flow-step--${step.status}`"
          >
            <span class="paper-flow-step-icon" aria-hidden="true">{{ topicStepIcon(step.status) }}</span>
            <div class="paper-flow-step-body">
              <span class="paper-flow-step-label">{{ step.label }}</span>
              <span class="paper-flow-step-code">{{ step.stageCode }}</span>
              <span v-if="step.checkpointKey && topicForm.humanCheckpoint" class="paper-flow-step-tag">
                检查点 · {{ step.checkpointKey }}
              </span>
            </div>
          </li>
        </ol>

        <div v-if="currentTopicRun.status === 'checkpoint' && currentTopicRun.checkpoint" class="paper-checkpoint">
          <div class="paper-checkpoint-head">
            <span class="paper-checkpoint-pause" aria-hidden="true">⏸</span>
            <div>
              <h3 class="paper-checkpoint-title">人工检查点 · {{ currentTopicRun.checkpoint.key }}</h3>
              <p class="paper-checkpoint-sub">{{ currentTopicRun.checkpoint.title }}</p>
            </div>
          </div>
          <ul class="paper-checkpoint-list">
            <li v-for="(line, i) in currentTopicRun.checkpoint.lines" :key="i">{{ line }}</li>
          </ul>
          <p class="paper-checkpoint-note">
            确认前后续阶段不会继续；驳回后本次 run 标记为失败，可改参数后点「再次运行」。
          </p>
          <div class="paper-checkpoint-actions">
            <button type="button" class="paper-btn-primary" @click="continueTopicAfterCheckpoint">
              确认并继续
            </button>
            <button type="button" class="paper-btn-secondary" @click="rejectTopicCheckpoint">
              驳回并重跑
            </button>
          </div>
        </div>

        <div v-else-if="currentTopicRun.status === 'completed'" class="paper-flow-done">
          产出将写入当前论文（idea 列表、新颖性报告、实验计划 · 演示）。
        </div>
      </section>

      <!-- 环境配置 -->
      <section v-else-if="activeModule === 'environment'" class="paper-panel paper-panel--env">
        <h2 class="paper-panel-title">全局默认</h2>
        <p class="paper-section-lead">对应用户偏好；新建工作流时将预填以下选项。</p>

        <div class="paper-field-grid">
          <label class="paper-field">
            <span class="paper-label">默认学科</span>
            <select v-model="envPreference.disciplineCode" class="paper-select">
              <option v-for="d in DISCIPLINE_OPTIONS" :key="d.code" :value="d.code">
                {{ d.label }}
              </option>
            </select>
          </label>

          <label class="paper-field">
            <span class="paper-label">默认目标会议/期刊</span>
            <input v-model="envPreference.defaultVenueText" type="text" class="paper-input" />
          </label>

          <label class="paper-field">
            <span class="paper-label">默认执行强度</span>
            <select v-model="envPreference.intensity" class="paper-select">
              <option v-for="o in intensityOptions" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
          </label>

          <label class="paper-field">
            <span class="paper-label">默认审计等级</span>
            <select v-model="envPreference.auditLevel" class="paper-select">
              <option v-for="o in auditOptions" :key="o.value" :value="o.value">
                {{ o.label }}
              </option>
            </select>
          </label>
        </div>

        <div class="paper-field paper-field--block">
          <span class="paper-label">默认文献来源</span>
          <div class="paper-check-group">
            <label
              v-for="src in LITERATURE_SOURCE_OPTIONS"
              :key="src.code"
              class="paper-check paper-check--inline"
            >
              <input
                type="checkbox"
                :checked="isLiteratureSourceChecked(src.code)"
                @change="toggleLiteratureSource(src.code, ($event.target as HTMLInputElement).checked)"
              />
              <span>{{ src.label }}</span>
            </label>
          </div>
        </div>

        <label class="paper-check">
          <input v-model="envPreference.humanCheckpoint" type="checkbox" />
          <span>
            <strong>默认开启人工检查点</strong>
          </span>
        </label>

        <label class="paper-check">
          <input v-model="envPreference.referenceGateEnabled" type="checkbox" />
          <span>
            <strong>参考文献门禁</strong>
            <span class="paper-hint paper-hint--inline">未验证文献不得进入引用与正文</span>
          </span>
        </label>
      </section>

      <PaperWorkflowPanels
        v-else-if="isSecondaryWorkflowModule"
        ref="workflowPanelsRef"
        :module-id="activeModule"
      />
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

.paper-manuscript-switcher {
  margin-top: 20px;
  padding: 14px 12px;
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
}

.paper-manuscript-label {
  display: block;
  margin-bottom: 8px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.65);
}

.paper-manuscript-select {
  width: 100%;
  padding: 9px 10px;
  font-size: 13px;
  font-weight: 600;
  color: #1e1b4b;
  background: #fff;
  border: none;
  border-radius: 8px;
  cursor: pointer;
}

.paper-manuscript-new {
  width: 100%;
  margin-top: 8px;
  padding: 8px 10px;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  background: transparent;
  border: 1px dashed rgba(255, 255, 255, 0.35);
  border-radius: 8px;
  cursor: pointer;
}

.paper-manuscript-new:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.5);
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

.paper-main-manuscript {
  margin: 6px 0 0;
  font-size: 13px;
  color: var(--atm-text-muted, #64748b);
}

.paper-main-manuscript strong {
  color: var(--atm-text, #1e1b4b);
  font-weight: 700;
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

.paper-panel-title--tight {
  margin-bottom: 0;
}

.paper-panel--flow {
  margin-top: 18px;
}

.paper-flow-head {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 14px;
  align-items: center;
  margin-bottom: 8px;
}

.paper-flow-badge {
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 700;
  border-radius: 999px;
}

.paper-flow-badge--running {
  color: #5b21b6;
  background: #ede9fe;
}

.paper-flow-badge--checkpoint {
  color: #b45309;
  background: #fef3c7;
}

.paper-flow-badge--completed {
  color: #15803d;
  background: #dcfce7;
}

.paper-flow-badge--failed {
  color: #b91c1c;
  background: #fee2e2;
}

.paper-flow-cancel {
  margin-left: auto;
  padding: 6px 12px;
  font-size: 13px;
  color: var(--atm-text-muted, #64748b);
  background: transparent;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  cursor: pointer;
}

.paper-flow-cancel:hover {
  border-color: #cbd5e1;
  color: var(--atm-text, #1e1b4b);
}

.paper-flow-steps {
  margin: 20px 0 0;
  padding: 0;
  list-style: none;
}

.paper-flow-step {
  display: flex;
  gap: 14px;
  align-items: flex-start;
  padding: 12px 0;
  border-left: 2px solid #e2e8f0;
  margin-left: 11px;
  padding-left: 22px;
  position: relative;
}

.paper-flow-step:last-child {
  border-left-color: transparent;
}

.paper-flow-step-icon {
  position: absolute;
  left: -12px;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  font-size: 11px;
  font-weight: 800;
  color: #94a3b8;
  background: #fff;
  border: 2px solid #e2e8f0;
  border-radius: 50%;
}

.paper-flow-step--running .paper-flow-step-icon {
  color: #7c3aed;
  border-color: #7c3aed;
  animation: paper-flow-pulse 1s ease-in-out infinite;
}

.paper-flow-step--completed .paper-flow-step-icon {
  color: #fff;
  background: #7c3aed;
  border-color: #7c3aed;
}

.paper-flow-step--failed .paper-flow-step-icon {
  color: #fff;
  background: #dc2626;
  border-color: #dc2626;
}

.paper-flow-step-body {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 10px;
  align-items: center;
  min-height: 22px;
}

.paper-flow-step-label {
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-text, #1e1b4b);
}

.paper-flow-step-code {
  font-size: 11px;
  font-family: ui-monospace, monospace;
  color: #94a3b8;
}

.paper-flow-step-tag {
  font-size: 11px;
  font-weight: 600;
  color: #b45309;
  background: #fffbeb;
  padding: 2px 8px;
  border-radius: 6px;
}

.paper-checkpoint {
  margin-top: 20px;
  padding: 18px 20px;
  background: linear-gradient(180deg, #fffbeb 0%, #fff 40%);
  border: 1px solid #fde68a;
  border-radius: 14px;
}

.paper-checkpoint-head {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.paper-checkpoint-pause {
  font-size: 22px;
  line-height: 1;
}

.paper-checkpoint-title {
  margin: 0;
  font-size: 15px;
  font-weight: 800;
  color: #92400e;
}

.paper-checkpoint-sub {
  margin: 4px 0 0;
  font-size: 13px;
  color: #78716c;
}

.paper-checkpoint-list {
  margin: 14px 0 0;
  padding-left: 20px;
  font-size: 13px;
  line-height: 1.55;
  color: var(--atm-text, #1e1b4b);
}

.paper-checkpoint-note {
  margin: 12px 0 0;
  font-size: 12px;
  color: var(--atm-text-muted, #64748b);
}

.paper-checkpoint-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 16px;
}

.paper-btn-primary {
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  background: var(--atm-gradient, linear-gradient(135deg, #7c3aed, #6366f1));
  border: none;
  border-radius: 10px;
  cursor: pointer;
}

.paper-btn-secondary {
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-text, #1e1b4b);
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
}

.paper-flow-done {
  margin-top: 16px;
  padding: 12px 14px;
  font-size: 13px;
  color: #15803d;
  background: #f0fdf4;
  border-radius: 10px;
}

@keyframes paper-flow-pulse {
  0%,
  100% {
    box-shadow: 0 0 0 0 rgba(124, 58, 237, 0.35);
  }
  50% {
    box-shadow: 0 0 0 6px rgba(124, 58, 237, 0);
  }
}

.paper-panel--flow code {
  font-size: 11px;
  padding: 1px 5px;
  background: #f1f5f9;
  border-radius: 4px;
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

.paper-panel--env {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.paper-section-lead {
  margin: -8px 0 16px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--atm-text-muted, #64748b);
}

.paper-divider {
  margin: 22px 0;
  border: none;
  border-top: 1px solid #e8eaf0;
}

.paper-check-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 20px;
  margin-top: 8px;
}

.paper-check--inline {
  margin-top: 0;
}

.paper-field--span2 {
  grid-column: 1 / -1;
}

@media (min-width: 900px) {
  .paper-field--span2 {
    grid-column: span 2;
  }
}
</style>
