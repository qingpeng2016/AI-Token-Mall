export type PaperModuleId =
  | 'topic-discovery'
  | 'literature-review'
  | 'experiment-planning'
  | 'auto-review'
  | 'paper-writing'
  | 'figure-generation'
  | 'manuscript-analysis'
  | 'environment'

export type PaperModuleMeta = {
  id: PaperModuleId
  label: string
  description: string
}

export const PAPER_MODULES: PaperModuleMeta[] = [
  {
    id: 'topic-discovery',
    label: '选题发现',
    description: '从研究方向生成候选 idea、novelty 检查和实验计划',
  },
  {
    id: 'literature-review',
    label: '文献综述',
    description: '基于真实数据源检索并生成可溯源的结构化综述',
  },
  {
    id: 'experiment-planning',
    label: '实验规划',
    description: '设计假设、基线、指标与可执行实验步骤',
  },
  {
    id: 'auto-review',
    label: '自动审查',
    description: '写作前：审查实验方案与上传数据，模拟审稿人挑 plan / 证据硬伤',
  },
  {
    id: 'paper-writing',
    label: '论文写作',
    description: '按目标 venue 风格分节撰写与润色稿件',
  },
  {
    id: 'figure-generation',
    label: '图表生成',
    description: '基于实验数据生成可复现的论文统计图（曲线、柱状、消融等）',
  },
  {
    id: 'manuscript-analysis',
    label: '成稿分析',
    description: '写作完成后：对全文做投稿前全面分析（贡献、实验、引用、venue  fit、kill argument）',
  },
  {
    id: 'environment',
    label: '环境配置',
    description: '新建运行时的默认科研偏好与文献策略（产出由平台按项目自动存储）',
  },
]

/** 工作台「当前论文」（对应 paper_manuscript） */
export type PaperManuscriptItem = {
  id: string
  title: string
  venueHint: string
  status: 'active' | 'archived'
}

export const DEMO_PAPER_MANUSCRIPTS: PaperManuscriptItem[] = [
  {
    id: 'ms-demo-1',
    title: '稀疏注意力机制',
    venueHint: 'NeurIPS 2026',
    status: 'active',
  },
  {
    id: 'ms-demo-2',
    title: '多模态 RAG 审计',
    venueHint: 'ACL 2026',
    status: 'active',
  },
]

export type ExecutionIntensity = 'fast' | 'balanced' | 'deep'
export type AuditLevel = 'standard' | 'polished' | 'strict'

export type TopicDiscoveryForm = {
  direction: string
  venue: string
  literatureSources: string
  intensity: ExecutionIntensity
  auditLevel: AuditLevel
  humanCheckpoint: boolean
}

/** 选题发现 run 内阶段（对齐 paper_run_stage.stage_code） */
export type TopicCheckpointKey = 'ideas_ready' | 'plan_ready'

export type TopicFlowStepDef = {
  stageCode: string
  label: string
  checkpointKey?: TopicCheckpointKey
}

export const TOPIC_DISCOVERY_FLOW_STEPS: TopicFlowStepDef[] = [
  { stageCode: 'retrieve', label: '文献检索与归纳' },
  { stageCode: 'generate_ideas', label: '生成候选选题' },
  { stageCode: 'novelty', label: '新颖性检查', checkpointKey: 'ideas_ready' },
  { stageCode: 'experiment_plan', label: '生成实验计划', checkpointKey: 'plan_ready' },
  { stageCode: 'audit', label: '引用与断言初 audit' },
]

export type TopicFlowStepStatus =
  | 'pending'
  | 'running'
  | 'completed'
  | 'waiting_human'
  | 'skipped'
  | 'failed'

export type LiteratureReviewForm = {
  theme: string
  yearFrom: string
  maxPapers: number
  literatureSources: string
  intensity: ExecutionIntensity
  auditLevel: AuditLevel
}

export type ExperimentPlanningForm = {
  ideaSummary: string
  venue: string
  baselinesText: string
  resources: string
  intensity: ExecutionIntensity
}

export type AutoReviewForm = {
  reviewFocus: 'plan' | 'data' | 'both'
  dataFileLabel: string
  strictKill: boolean
  auditLevel: AuditLevel
}

export type PaperWritingForm = {
  venue: string
  sections: string[]
  useReviewArtifact: boolean
  useLitReview: boolean
  tone: 'concise' | 'standard'
}

export type FigureGenerationForm = {
  datasetLabel: string
  chartTypes: string[]
  includeErrorBars: boolean
  captionLang: 'zh' | 'en'
}

export type ManuscriptAnalysisForm = {
  depth: AuditLevel
  includeFigures: boolean
  includeBib: boolean
}

export const DEFAULT_LITERATURE_REVIEW: LiteratureReviewForm = {
  theme: '稀疏注意力与长上下文 KV 压缩',
  yearFrom: '2022',
  maxPapers: 80,
  literatureSources: 'arxiv, openalex, semantic-scholar',
  intensity: 'balanced',
  auditLevel: 'polished',
}

export const DEFAULT_EXPERIMENT_PLANNING: ExperimentPlanningForm = {
  ideaSummary: '训练-free 动态稀疏路由，固定 KV 预算 B',
  venue: 'NeurIPS/ICLR/ICML',
  baselinesText: 'Full Attention, StreamingLLM, H2O, SparseAttn',
  resources: '8×A100 80GB · 约 4 周',
  intensity: 'balanced',
}

export const DEFAULT_AUTO_REVIEW: AutoReviewForm = {
  reviewFocus: 'both',
  dataFileLabel: 'results_main.csv（已上传 · 演示）',
  strictKill: true,
  auditLevel: 'strict',
}

export const DEFAULT_PAPER_WRITING: PaperWritingForm = {
  venue: 'NeurIPS/ICLR/ICML',
  sections: ['abstract', 'intro', 'related', 'method', 'experiments'],
  useReviewArtifact: true,
  useLitReview: true,
  tone: 'standard',
}

export const DEFAULT_FIGURE_GENERATION: FigureGenerationForm = {
  datasetLabel: 'results_main.csv',
  chartTypes: ['bar', 'line'],
  includeErrorBars: true,
  captionLang: 'en',
}

export const DEFAULT_MANUSCRIPT_ANALYSIS: ManuscriptAnalysisForm = {
  depth: 'strict',
  includeFigures: true,
  includeBib: true,
}

export const PAPER_WRITING_SECTION_OPTIONS = [
  { value: 'abstract', label: 'Abstract' },
  { value: 'intro', label: 'Introduction' },
  { value: 'related', label: 'Related Work' },
  { value: 'method', label: 'Method' },
  { value: 'experiments', label: 'Experiments' },
  { value: 'conclusion', label: 'Conclusion' },
] as const

export const FIGURE_CHART_OPTIONS = [
  { value: 'bar', label: '柱状图（主结果）' },
  { value: 'line', label: '折线（训练/缩放）' },
  { value: 'scatter', label: '散点（吞吐-准确率）' },
  { value: 'ablation', label: '消融分组图' },
] as const

export const DEFAULT_TOPIC_DISCOVERY: TopicDiscoveryForm = {
  direction: '',
  venue: 'NeurIPS/ICLR/ICML',
  literatureSources: 'arxiv, openalex, semantic-scholar',
  intensity: 'balanced',
  auditLevel: 'polished',
  humanCheckpoint: true,
}

export type EnvironmentPreferenceForm = {
  disciplineCode: string
  defaultVenueText: string
  literatureSourceCodes: string[]
  intensity: ExecutionIntensity
  auditLevel: AuditLevel
  humanCheckpoint: boolean
  referenceGateEnabled: boolean
}

export const LITERATURE_SOURCE_OPTIONS = [
  { code: 'arxiv', label: 'arXiv' },
  { code: 'openalex', label: 'OpenAlex' },
  { code: 'semantic_scholar', label: 'Semantic Scholar' },
  { code: 'crossref', label: 'Crossref' },
  { code: 'pubmed', label: 'PubMed' },
] as const

export const DISCIPLINE_OPTIONS = [
  { code: 'cs_ai', label: '计算机 / 人工智能' },
  { code: 'general', label: '跨学科通用' },
] as const

export const DEFAULT_ENV_PREFERENCE: EnvironmentPreferenceForm = {
  disciplineCode: 'cs_ai',
  defaultVenueText: 'NeurIPS/ICLR/ICML',
  literatureSourceCodes: ['arxiv', 'openalex', 'semantic_scholar'],
  intensity: 'balanced',
  auditLevel: 'polished',
  humanCheckpoint: true,
  referenceGateEnabled: true,
}

