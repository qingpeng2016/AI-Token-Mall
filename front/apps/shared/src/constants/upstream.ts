import type { UpstreamName } from '../types/product'

export const UPSTREAM_LABEL: Record<UpstreamName, string> = {
  openai: 'ChatGPT / OpenAI',
  anthropic: 'Claude',
  xai: 'Grok',
  gemini: 'Gemini',
  perplexity: 'Perplexity',
}

export const UPSTREAM_FILTER_OPTIONS: { value: 'all' | UpstreamName; label: string }[] = [
  { value: 'all', label: '全部套餐' },
  { value: 'openai', label: 'ChatGPT' },
  { value: 'anthropic', label: 'Claude' },
  { value: 'xai', label: 'Grok' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'perplexity', label: 'Perplexity' },
]
