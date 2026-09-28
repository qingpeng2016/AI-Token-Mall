import { ref, type Ref } from 'vue'
import type { TutorialArticle, TutorialCategory } from '@ai-token-mall/shared'
import { tutorialApi } from '@/api'
import {
  blogCategories,
  blogPageMeta,
  blogPosts,
  getBlogPost,
  type BlogPost,
} from '@/mocks/blog'
import type { ReloadOptions } from '@/composables/reloadOptions'

export type TutorialFilterCategory = {
  id: string
  label: string
}

const pageTitle = ref('')
const categories = ref<TutorialFilterCategory[]>([])
const articles = ref<BlogPost[]>([])
const loading = ref(false)

let loadSeq = 0

function fromMock() {
  pageTitle.value = blogPageMeta.title
  categories.value = blogCategories.map((c) => ({ id: c.id, label: c.label }))
  articles.value = blogPosts
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

function applyApiList(data: {
  title?: string
  categories?: TutorialCategory[]
  articles?: TutorialArticle[]
}) {
  pageTitle.value = data.title || blogPageMeta.title
  categories.value = [
    { id: 'all', label: '全部' },
    ...(data.categories ?? [])
      .slice()
      .sort((a, b) => a.sort - b.sort)
      .map((c) => ({ id: c.code, label: c.name })),
  ]
  articles.value = (data.articles ?? []).map(mapApiArticle)
}

export async function reloadTutorialBlog(
  options?: ReloadOptions,
): Promise<void> {
  const seq = ++loadSeq
  const soft = options?.soft === true
  loading.value = !soft || articles.value.length === 0
  if (!soft) {
    pageTitle.value = ''
    categories.value = []
    articles.value = []
  }
  try {
    const data = await tutorialApi.list()
    if (seq !== loadSeq) return
    applyApiList(data)
  } catch {
    if (seq !== loadSeq) return
    fromMock()
  } finally {
    if (seq === loadSeq) loading.value = false
  }
}

export function findCachedBlogPost(slug: string): BlogPost | undefined {
  return articles.value.find((a) => a.slug === slug)
}

let prefetchedSlug: string | null = null
let prefetchedPost: BlogPost | null = null

export async function prefetchBlogArticle(slug: string): Promise<void> {
  try {
    const data = await tutorialApi.detail(slug)
    if (data?.slug) {
      prefetchedPost = mapApiArticle(data)
      prefetchedSlug = slug
      return
    }
  } catch {
    /* fallback in view */
  }
  const fromList = findCachedBlogPost(slug)
  if (fromList?.body?.length) {
    prefetchedPost = fromList
    prefetchedSlug = slug
    return
  }
  prefetchedPost = getBlogPost(slug) ?? null
  prefetchedSlug = prefetchedPost ? slug : null
}

/** 路由守卫预取成功后，详情页直接消费，避免再闪 skeleton */
export function takePrefetchedBlogPost(slug: string): BlogPost | null {
  if (prefetchedSlug === slug && prefetchedPost) {
    const post = prefetchedPost
    prefetchedSlug = null
    prefetchedPost = null
    return post
  }
  return null
}

export function useTutorialBlog(): {
  pageTitle: Ref<string>
  categories: Ref<TutorialFilterCategory[]>
  articles: Ref<BlogPost[]>
  loading: Ref<boolean>
  reload: () => Promise<void>
} {
  return {
    pageTitle,
    categories,
    articles,
    loading,
    reload: reloadTutorialBlog,
  }
}
