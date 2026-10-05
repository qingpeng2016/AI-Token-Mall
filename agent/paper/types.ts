export type PaperModuleId =
  | 'topic-discovery'
  | 'literature-review'
  | 'experiment-planning'
  | 'auto-review'
  | 'paper-writing'
  | 'figure-generation'
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
    description: '多轮审稿式审查，识别逻辑与证据弱点',
  },
  {
    id: 'paper-writing',
    label: '论文写作',
    description: '按目标 venue 风格分节撰写与润色稿件',
  },
  {
    id: 'figure-generation',
    label: '图表生成',
    description: '根据实验数据或规划生成论文级图表',
  },
  {
    id: 'environment',
    label: '环境配置',
    description: '模型 Provider、Base URL、API Key 与算力环境',
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

export const DEFAULT_TOPIC_DISCOVERY: TopicDiscoveryForm = {
  direction: '',
  venue: 'NeurIPS/ICLR/ICML',
  literatureSources: 'arxiv, openalex, semantic-scholar',
  intensity: 'balanced',
  auditLevel: 'polished',
  humanCheckpoint: true,
}
