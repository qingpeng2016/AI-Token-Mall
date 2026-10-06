<script setup lang="ts">
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  DEMO_AUTO_REVIEW,
  DEMO_EXPERIMENT_PLAN,
  DEMO_FIGURES,
  DEMO_LIT_REVIEW,
  DEMO_MANUSCRIPT,
  DEMO_MANUSCRIPT_ANALYSIS,
} from './demoModuleOutputs'
import {
  DEFAULT_AUTO_REVIEW,
  DEFAULT_EXPERIMENT_PLANNING,
  DEFAULT_FIGURE_GENERATION,
  DEFAULT_LITERATURE_REVIEW,
  DEFAULT_MANUSCRIPT_ANALYSIS,
  DEFAULT_PAPER_WRITING,
  FIGURE_CHART_OPTIONS,
  PAPER_WRITING_SECTION_OPTIONS,
  type AutoReviewForm,
  type ExperimentPlanningForm,
  type FigureGenerationForm,
  type LiteratureReviewForm,
  type ManuscriptAnalysisForm,
  type PaperModuleId,
  type PaperWritingForm,
} from './types'

const { moduleId } = defineProps<{
  moduleId: PaperModuleId
}>()

const litForm = reactive<LiteratureReviewForm>({ ...DEFAULT_LITERATURE_REVIEW })
const planForm = reactive<ExperimentPlanningForm>({ ...DEFAULT_EXPERIMENT_PLANNING })
const reviewForm = reactive<AutoReviewForm>({ ...DEFAULT_AUTO_REVIEW })
const writeForm = reactive<PaperWritingForm>({
  ...DEFAULT_PAPER_WRITING,
  sections: [...DEFAULT_PAPER_WRITING.sections],
})
const figureForm = reactive<FigureGenerationForm>({
  ...DEFAULT_FIGURE_GENERATION,
  chartTypes: [...DEFAULT_FIGURE_GENERATION.chartTypes],
})
const analysisForm = reactive<ManuscriptAnalysisForm>({ ...DEFAULT_MANUSCRIPT_ANALYSIS })

const resultVisible = ref<Partial<Record<PaperModuleId, boolean>>>({})

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

function toggleWritingSection(value: string, checked: boolean) {
  const set = new Set(writeForm.sections)
  if (checked) set.add(value)
  else set.delete(value)
  writeForm.sections = [...set]
}

function isWritingSectionChecked(value: string) {
  return writeForm.sections.includes(value)
}

function toggleChartType(value: string, checked: boolean) {
  const set = new Set(figureForm.chartTypes)
  if (checked) set.add(value)
  else set.delete(value)
  figureForm.chartTypes = [...set]
}

function isChartChecked(value: string) {
  return figureForm.chartTypes.includes(value)
}

async function runModule(id: PaperModuleId): Promise<boolean> {
  if (id === 'literature-review' && !litForm.theme.trim()) {
    ElMessage.warning('请填写综述主题')
    return false
  }
  if (id === 'experiment-planning' && !planForm.ideaSummary.trim()) {
    ElMessage.warning('请填写核心 idea / 假设')
    return false
  }
  if (id === 'paper-writing' && writeForm.sections.length === 0) {
    ElMessage.warning('请至少选择一个章节')
    return false
  }
  if (id === 'figure-generation' && figureForm.chartTypes.length === 0) {
    ElMessage.warning('请至少选择一种图表类型')
    return false
  }

  await new Promise((r) => setTimeout(r, 750))
  resultVisible.value = { ...resultVisible.value, [id]: true }
  return true
}

defineExpose({ runModule })
</script>

<template>
  <div class="wf-root">
  <!-- 文献综述 -->
  <div v-if="moduleId === 'literature-review'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <label class="wf-field wf-field--block">
        <span class="wf-label">综述主题 <em class="req">*</em></span>
        <textarea v-model="litForm.theme" class="wf-textarea" rows="3" />
      </label>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">起始年份</span>
          <input v-model="litForm.yearFrom" type="text" class="wf-input" />
        </label>
        <label class="wf-field">
          <span class="wf-label">最大文献数</span>
          <input v-model.number="litForm.maxPapers" type="number" class="wf-input" min="10" />
        </label>
        <label class="wf-field wf-field--span2">
          <span class="wf-label">文献来源</span>
          <input v-model="litForm.literatureSources" type="text" class="wf-input" />
        </label>
        <label class="wf-field">
          <span class="wf-label">执行强度</span>
          <select v-model="litForm.intensity" class="wf-select">
            <option v-for="o in intensityOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
        </label>
        <label class="wf-field">
          <span class="wf-label">审计等级</span>
          <select v-model="litForm.auditLevel" class="wf-select">
            <option v-for="o in auditOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
        </label>
      </div>
    </section>
    <section v-if="resultVisible['literature-review']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">综述产出（演示）</h2>
      <p class="wf-meta">检索 {{ DEMO_LIT_REVIEW.retrieved }} 篇 · 验证通过 {{ DEMO_LIT_REVIEW.verified }} 篇</p>
      <div v-for="sec in DEMO_LIT_REVIEW.sections" :key="sec.title" class="wf-block">
        <h3 class="wf-subtitle">{{ sec.title }}</h3>
        <ul class="wf-list">
          <li v-for="(p, i) in sec.papers" :key="i">{{ p }}</li>
        </ul>
      </div>
      <p class="wf-callout"><strong>Research gap：</strong>{{ DEMO_LIT_REVIEW.gap }}</p>
      <pre class="wf-pre">{{ DEMO_LIT_REVIEW.excerpt }}</pre>
    </section>
  </div>

  <!-- 实验规划 -->
  <div v-else-if="moduleId === 'experiment-planning'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <label class="wf-field wf-field--block">
        <span class="wf-label">核心 idea / 假设 <em class="req">*</em></span>
        <textarea v-model="planForm.ideaSummary" class="wf-textarea" rows="2" />
      </label>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">目标 venue</span>
          <input v-model="planForm.venue" type="text" class="wf-input" />
        </label>
        <label class="wf-field">
          <span class="wf-label">执行强度</span>
          <select v-model="planForm.intensity" class="wf-select">
            <option v-for="o in intensityOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
        </label>
        <label class="wf-field wf-field--span2">
          <span class="wf-label">基线（逗号或换行）</span>
          <textarea v-model="planForm.baselinesText" class="wf-textarea" rows="2" />
        </label>
        <label class="wf-field wf-field--span2">
          <span class="wf-label">资源与时间</span>
          <input v-model="planForm.resources" type="text" class="wf-input" />
        </label>
      </div>
    </section>
    <section v-if="resultVisible['experiment-planning']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">实验计划（演示）</h2>
      <p class="wf-kv"><span>假设</span>{{ DEMO_EXPERIMENT_PLAN.hypothesis }}</p>
      <p class="wf-kv"><span>基线</span>{{ DEMO_EXPERIMENT_PLAN.baselines.join(' · ') }}</p>
      <p class="wf-kv"><span>指标</span>{{ DEMO_EXPERIMENT_PLAN.metrics.join(' · ') }}</p>
      <p class="wf-kv"><span>消融</span>{{ DEMO_EXPERIMENT_PLAN.ablations.join(' · ') }}</p>
      <ol class="wf-list wf-list--ordered">
        <li v-for="(s, i) in DEMO_EXPERIMENT_PLAN.steps" :key="i">{{ s }}</li>
      </ol>
    </section>
  </div>

  <!-- 自动审查（写前） -->
  <div v-else-if="moduleId === 'auto-review'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <p class="wf-lead">写作前审查：实验方案与上传数据（非完整稿）。</p>
      <div class="wf-field wf-field--block">
        <span class="wf-label">审查对象</span>
        <div class="wf-radios">
          <label><input v-model="reviewForm.reviewFocus" type="radio" value="plan" /> 仅实验方案</label>
          <label><input v-model="reviewForm.reviewFocus" type="radio" value="data" /> 仅上传数据</label>
          <label><input v-model="reviewForm.reviewFocus" type="radio" value="both" /> 方案 + 数据</label>
        </div>
      </div>
      <label class="wf-field wf-field--block">
        <span class="wf-label">数据文件</span>
        <input v-model="reviewForm.dataFileLabel" type="text" class="wf-input" readonly />
      </label>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">审计等级</span>
          <select v-model="reviewForm.auditLevel" class="wf-select">
            <option v-for="o in auditOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
        </label>
        <label class="wf-check wf-check--solo">
          <input v-model="reviewForm.strictKill" type="checkbox" />
          <span>启用 kill argument（严苛反驳）</span>
        </label>
      </div>
    </section>
    <section v-if="resultVisible['auto-review']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">审查意见（演示）</h2>
      <p class="wf-meta">{{ DEMO_AUTO_REVIEW.target }}</p>
      <div class="wf-scores">
        <span>Rigor {{ DEMO_AUTO_REVIEW.scores.rigor }}</span>
        <span>Complete {{ DEMO_AUTO_REVIEW.scores.completeness }}</span>
        <span>Claims {{ DEMO_AUTO_REVIEW.scores.claimSupport }}</span>
      </div>
      <ul class="wf-findings">
        <li v-for="(f, i) in DEMO_AUTO_REVIEW.findings" :key="i" :class="`wf-finding--${f.level}`">
          <strong>{{ f.level === 'major' ? 'Major' : 'Minor' }}</strong> {{ f.text }}
        </li>
      </ul>
      <p class="wf-callout wf-callout--warn"><strong>Kill：</strong>{{ DEMO_AUTO_REVIEW.kill }}</p>
    </section>
  </div>

  <!-- 论文写作 -->
  <div v-else-if="moduleId === 'paper-writing'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">目标 venue</span>
          <input v-model="writeForm.venue" type="text" class="wf-input" />
        </label>
        <label class="wf-field">
          <span class="wf-label">文风</span>
          <select v-model="writeForm.tone" class="wf-select">
            <option value="concise">简洁</option>
            <option value="standard">标准 conference</option>
          </select>
        </label>
      </div>
      <div class="wf-field wf-field--block">
        <span class="wf-label">生成章节</span>
        <div class="wf-check-group">
          <label v-for="opt in PAPER_WRITING_SECTION_OPTIONS" :key="opt.value" class="wf-check wf-check--inline">
            <input
              type="checkbox"
              :checked="isWritingSectionChecked(opt.value)"
              @change="toggleWritingSection(opt.value, ($event.target as HTMLInputElement).checked)"
            />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </div>
      <label class="wf-check">
        <input v-model="writeForm.useLitReview" type="checkbox" />
        <span>引用已有文献综述 artifact</span>
      </label>
      <label class="wf-check">
        <input v-model="writeForm.useReviewArtifact" type="checkbox" />
        <span>参考写前审查意见修订 Experiments 表述</span>
      </label>
    </section>
    <section v-if="resultVisible['paper-writing']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">稿件（演示）</h2>
      <p class="wf-kv"><span>标题</span>{{ DEMO_MANUSCRIPT.title }}</p>
      <table class="wf-table">
        <thead>
          <tr>
            <th>章节</th>
            <th>字数</th>
            <th>状态</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in DEMO_MANUSCRIPT.sections" :key="s.name">
            <td>{{ s.name }}</td>
            <td>{{ s.words }}</td>
            <td>
              <span class="wf-tag" :class="s.status === 'done' ? 'wf-tag--ok' : 'wf-tag--draft'">{{
                s.status
              }}</span>
            </td>
          </tr>
        </tbody>
      </table>
      <p class="wf-meta">引用 {{ DEMO_MANUSCRIPT.cites }} · 已验证 {{ DEMO_MANUSCRIPT.verifiedCites }}</p>
      <pre class="wf-pre">{{ DEMO_MANUSCRIPT.excerpt }}</pre>
    </section>
  </div>

  <!-- 图表生成 -->
  <div v-else-if="moduleId === 'figure-generation'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <p class="wf-lead">基于实验数据生成可复现统计图（非通用美工工具）。</p>
      <label class="wf-field wf-field--block">
        <span class="wf-label">数据文件</span>
        <select v-model="figureForm.datasetLabel" class="wf-select">
          <option value="results_main.csv">results_main.csv（演示）</option>
          <option value="ablation_B.csv">ablation_B.csv（演示）</option>
          <option value="throughput.csv">throughput.csv（演示）</option>
        </select>
      </label>
      <div class="wf-field wf-field--block">
        <span class="wf-label">图表类型</span>
        <div class="wf-check-group">
          <label v-for="opt in FIGURE_CHART_OPTIONS" :key="opt.value" class="wf-check wf-check--inline">
            <input
              type="checkbox"
              :checked="isChartChecked(opt.value)"
              @change="toggleChartType(opt.value, ($event.target as HTMLInputElement).checked)"
            />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </div>
      <label class="wf-check">
        <input v-model="figureForm.includeErrorBars" type="checkbox" />
        <span>误差条 / 多 seed 汇总</span>
      </label>
      <label class="wf-field">
        <span class="wf-label">Caption 语言</span>
        <select v-model="figureForm.captionLang" class="wf-select">
          <option value="en">English</option>
          <option value="zh">中文</option>
        </select>
      </label>
    </section>
    <section v-if="resultVisible['figure-generation']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">Figure 产出（演示）</h2>
      <div class="wf-figure-grid">
        <div v-for="fig in DEMO_FIGURES.files" :key="fig.id" class="wf-figure-card">
          <div class="wf-figure-preview" :class="`wf-figure-preview--${fig.type}`">
            <div v-for="n in 5" :key="n" class="wf-bar" :style="{ height: `${30 + n * 12}%` }" />
          </div>
          <p class="wf-figure-cap">{{ fig.title }}</p>
          <code class="wf-figure-file">{{ fig.note }}</code>
        </div>
      </div>
      <table class="wf-table wf-table--compact">
        <thead>
          <tr>
            <th>Method</th>
            <th>Score</th>
            <th>±</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in DEMO_FIGURES.dataPreview" :key="row.method">
            <td>{{ row.method }}</td>
            <td>{{ row.score }}</td>
            <td>{{ row.std }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>

  <!-- 成稿分析 -->
  <div v-else-if="moduleId === 'manuscript-analysis'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <p class="wf-lead">写作完成后对全文做投稿前全面分析。</p>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">分析深度</span>
          <select v-model="analysisForm.depth" class="wf-select">
            <option v-for="o in auditOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
        </label>
      </div>
      <label class="wf-check">
        <input v-model="analysisForm.includeFigures" type="checkbox" />
        <span>纳入图表与表（{{ DEMO_FIGURES.files.length }} 个 demo figure）</span>
      </label>
      <label class="wf-check">
        <input v-model="analysisForm.includeBib" type="checkbox" />
        <span>纳入参考文献门禁结果</span>
      </label>
    </section>
    <section v-if="resultVisible['manuscript-analysis']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">成稿分析报告（演示）</h2>
      <p class="wf-score-big">总评 {{ DEMO_MANUSCRIPT_ANALYSIS.overall }} / 10</p>
      <p class="wf-meta">{{ DEMO_MANUSCRIPT_ANALYSIS.recommendation }}</p>
      <div class="wf-dim-grid">
        <div v-for="d in DEMO_MANUSCRIPT_ANALYSIS.dimensions" :key="d.name" class="wf-dim">
          <span>{{ d.name }}</span>
          <strong>{{ d.score }}</strong>
        </div>
      </div>
      <h3 class="wf-subtitle">Must fix</h3>
      <ul class="wf-list">
        <li v-for="(m, i) in DEMO_MANUSCRIPT_ANALYSIS.mustFix" :key="i">{{ m }}</li>
      </ul>
      <p class="wf-callout wf-callout--warn"><strong>Kill argument：</strong>{{ DEMO_MANUSCRIPT_ANALYSIS.kill }}</p>
    </section>
  </div>
  </div>
</template>

<style scoped>
.wf-root {
  display: contents;
}

.wf-stack {
  display: contents;
}

.wf-panel {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.wf-panel--result {
  margin-top: 18px;
  border-color: #ddd6fe;
  background: linear-gradient(180deg, #faf5ff 0%, #fff 120px);
}

.wf-title {
  margin: 0 0 18px;
  font-size: 16px;
  font-weight: 700;
  color: #1e1b4b;
}

.wf-subtitle {
  margin: 16px 0 8px;
  font-size: 14px;
  font-weight: 700;
  color: #1e1b4b;
}

.wf-lead {
  margin: -8px 0 16px;
  font-size: 13px;
  color: #64748b;
}

.wf-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px 20px;
}

@media (max-width: 900px) {
  .wf-grid {
    grid-template-columns: 1fr;
  }
}

.wf-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wf-field--block {
  margin-bottom: 4px;
}

.wf-field--span2 {
  grid-column: 1 / -1;
}

.wf-label {
  font-size: 13px;
  font-weight: 600;
  color: #1e1b4b;
}

.wf-label .req {
  color: #e53935;
  font-style: normal;
}

.wf-input,
.wf-select,
.wf-textarea {
  width: 100%;
  padding: 10px 12px;
  font-size: 14px;
  color: #1e1b4b;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.wf-textarea {
  resize: vertical;
  font-family: inherit;
  line-height: 1.5;
}

.wf-check {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  margin-top: 14px;
  font-size: 14px;
  cursor: pointer;
}

.wf-check--inline {
  margin-top: 0;
}

.wf-check--solo {
  justify-content: flex-end;
  align-self: end;
}

.wf-check-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
  margin-top: 8px;
}

.wf-radios {
  display: flex;
  flex-wrap: wrap;
  gap: 14px 20px;
  margin-top: 8px;
  font-size: 14px;
}

.wf-meta {
  margin: 0 0 12px;
  font-size: 13px;
  color: #64748b;
}

.wf-kv {
  margin: 0 0 10px;
  font-size: 14px;
  line-height: 1.5;
}

.wf-kv span {
  display: inline-block;
  min-width: 4em;
  margin-right: 8px;
  font-weight: 700;
  color: #64748b;
}

.wf-list {
  margin: 0;
  padding-left: 20px;
  font-size: 13px;
  line-height: 1.55;
}

.wf-list--ordered {
  margin-top: 12px;
}

.wf-pre {
  margin: 12px 0 0;
  padding: 12px;
  font-size: 12px;
  line-height: 1.5;
  color: #334155;
  white-space: pre-wrap;
  background: #f1f5f9;
  border-radius: 10px;
}

.wf-callout {
  margin: 16px 0 0;
  padding: 12px 14px;
  font-size: 13px;
  line-height: 1.5;
  background: #f0fdf4;
  border-radius: 10px;
}

.wf-callout--warn {
  background: #fffbeb;
}

.wf-block {
  margin-top: 8px;
}

.wf-scores {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 13px;
  font-weight: 600;
}

.wf-scores span {
  padding: 6px 10px;
  background: #ede9fe;
  border-radius: 8px;
}

.wf-findings {
  margin: 0;
  padding: 0;
  list-style: none;
}

.wf-finding--major,
.wf-finding--minor {
  margin-bottom: 8px;
  padding: 10px 12px;
  font-size: 13px;
  border-radius: 10px;
}

.wf-finding--major {
  background: #fee2e2;
}

.wf-finding--minor {
  background: #f1f5f9;
}

.wf-table {
  width: 100%;
  margin: 12px 0;
  font-size: 13px;
  border-collapse: collapse;
}

.wf-table th,
.wf-table td {
  padding: 8px 10px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.wf-table th {
  font-weight: 700;
  color: #64748b;
}

.wf-tag {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
}

.wf-tag--ok {
  color: #15803d;
}

.wf-tag--draft {
  color: #b45309;
}

.wf-figure-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

@media (max-width: 900px) {
  .wf-figure-grid {
    grid-template-columns: 1fr;
  }
}

.wf-figure-card {
  padding: 12px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.wf-figure-preview {
  display: flex;
  gap: 6px;
  align-items: flex-end;
  justify-content: center;
  height: 100px;
  padding: 8px;
  background: #f8fafc;
  border-radius: 8px;
}

.wf-bar {
  flex: 1;
  max-width: 24px;
  background: linear-gradient(180deg, #7c3aed, #6366f1);
  border-radius: 4px 4px 0 0;
}

.wf-figure-preview--line .wf-bar {
  background: #6366f1;
  opacity: 0.85;
}

.wf-figure-cap {
  margin: 10px 0 4px;
  font-size: 13px;
  font-weight: 600;
}

.wf-figure-file {
  font-size: 11px;
  color: #64748b;
}

.wf-score-big {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  color: #5b21b6;
}

.wf-dim-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  margin: 16px 0;
}

@media (max-width: 700px) {
  .wf-dim-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.wf-dim {
  padding: 12px;
  text-align: center;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.wf-dim span {
  display: block;
  font-size: 11px;
  color: #64748b;
}

.wf-dim strong {
  font-size: 18px;
  color: #1e1b4b;
}
</style>
