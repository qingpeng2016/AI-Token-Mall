import { computed, ref, type ComputedRef, type Ref } from 'vue'
import type { TutorialArticle, TutorialCategory } from '@ai-token-mall/shared'
import { readSessionCache, writeSessionCache } from '@ai-token-mall/shared'
import { tutorialApi } from '@/api'
import {
  blogCategories,
  blogPageMeta,
  blogPosts,
  type BlogPost,
} from '@/mocks/blog'

const SESSION_KEY = 'atm:tutorial-list:v3'

export type TutorialFilterCategory = {
  id: string
  label: string
}

type TutorialCache = {
  title: string
  categories: TutorialFilterCategory[]
  articles: BlogPost[]
  fromApi?: boolean
}

const pageTitle = ref('')
const categories = ref<TutorialFilterCategory[]>([])
const articles = ref<BlogPost[]>([])
const revalidating = ref(false)

let inflight: Promise<void> | null = null

function fromMock(): TutorialCache {
  return {
    title: blogPageMeta.title,
    categories: blogCategories.map((c) => ({ id: c.id, label: c.label })),
    articles: blogPosts,
    fromApi: false,
  }
}

function mapApiArticle(a: TutorialArticle): BlogPost {
  return {
    slug: a.slug,
    category: a.category_code as BlogPost['category'],
    categoryLabel: a.category_name,
    date: a.date,
    title: a.title,
    excerpt: a.excerpt,
    body: a.body ?? [],
  }
}

function applyCache(cache: TutorialCache) {
  pageTitle.value = cache.title
  categories.value = cache.categories
  articles.value = cache.articles
}

function applyApiList(data: {
  title?: string
  categories?: TutorialCategory[]
  articles?: TutorialArticle[]
}) {
  const cache: TutorialCache = {
    title: data.title || blogPageMeta.title,
    categories: [
      { id: 'all', label: '全部' },
      ...(data.categories ?? [])
        .slice()
        .sort((a, b) => a.sort - b.sort)
        .map((c) => ({ id: c.code, label: c.name })),
    ],
    articles: (data.articles ?? []).map(mapApiArticle),
    fromApi: true,
  }
  applyCache(cache)
  writeSessionCache(SESSION_KEY, cache)
}

function seedDisplayList() {
  const session = readSessionCache<TutorialCache>(SESSION_KEY)
  if (session?.fromApi) {
    applyCache(session)
    return
  }
  applyCache(fromMock())
}

async function revalidateList(): Promise<void> {
  try {
    const data = await tutorialApi.list()
    applyApiList(data)
  } catch {
    /* 保留当前展示（mock / session / 上次 API） */
  }
}

function revalidateInBackground(): Promise<void> {
  if (inflight) return inflight
  revalidating.value = true
  inflight = revalidateList().finally(() => {
    revalidating.value = false
    inflight = null
  })
  return inflight
}

export function useTutorialBlog(): {
  pageTitle: Ref<string>
  categories: Ref<TutorialFilterCategory[]>
  articles: Ref<BlogPost[]>
  revalidating: ComputedRef<boolean>
  reload: () => Promise<void>
} {
  if (!pageTitle.value && !articles.value.length) {
    seedDisplayList()
  }
  void revalidateInBackground()

  async function reload() {
    revalidating.value = true
    try {
      await revalidateList()
    } finally {
      revalidating.value = false
    }
  }

  return {
    pageTitle,
    categories,
    articles,
    revalidating: computed(() => revalidating.value),
    reload,
  }
}
