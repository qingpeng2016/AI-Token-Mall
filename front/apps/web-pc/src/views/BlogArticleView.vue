<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import type { TutorialArticle } from '@ai-token-mall/shared'
import { readSessionCache, writeSessionCache } from '@ai-token-mall/shared'
import { tutorialApi } from '@/api'
import { getBlogPost, type BlogPost } from '@/mocks/blog'

const route = useRoute()
const router = useRouter()

const slug = computed(() => String(route.params.slug ?? ''))
const revalidating = ref(false)

function cacheKey(s: string) {
  return `atm:tutorial-article:${s}`
}

function mapArticle(a: TutorialArticle): BlogPost {
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

function hydratePost(currentSlug: string): BlogPost | null {
  return (
    readSessionCache<BlogPost>(cacheKey(currentSlug)) ??
    getBlogPost(currentSlug) ??
    null
  )
}

const post = ref<BlogPost | null>(null)

async function loadPost(currentSlug: string) {
  post.value = hydratePost(currentSlug)
  revalidating.value = true
  try {
    const data = await tutorialApi.detail(currentSlug)
    if (data?.slug) {
      const mapped = mapArticle(data)
      post.value = mapped
      writeSessionCache(cacheKey(currentSlug), mapped)
      return
    }
  } catch {
    /* 保留 hydrate */
  } finally {
    revalidating.value = false
  }
  if (!post.value) {
    router.replace('/blog')
  }
}

watch(
  slug,
  (s) => {
    if (s) void loadPost(s)
  },
  { immediate: true },
)
</script>

<template>
  <article v-if="post" class="article-page">
    <div class="atm-container article-container">
      <nav class="article-breadcrumb" aria-label="面包屑">
        <RouterLink to="/">首页</RouterLink>
        <span class="sep">/</span>
        <RouterLink to="/blog">教程</RouterLink>
        <span class="sep">/</span>
        <span class="current">正文</span>
      </nav>

      <header class="article-head">
        <div class="article-meta">
          <span class="article-tag">{{ post.categoryLabel }}</span>
          <time :datetime="post.date">{{ post.date }}</time>
        </div>
        <h1 class="article-title">{{ post.title }}</h1>
      </header>

      <div class="article-body">
        <p v-for="(para, i) in post.body" :key="i">{{ para }}</p>
      </div>

      <footer class="article-foot">
        <RouterLink to="/blog" class="back-link">← 返回教程列表</RouterLink>
      </footer>
    </div>
  </article>
</template>

<style scoped>
.article-page {
  padding: 28px 0 72px;
}

.article-container {
  max-width: 720px;
}

.article-breadcrumb {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 28px;
  font-size: 14px;
  color: var(--atm-text-muted);
}

.article-breadcrumb a {
  color: var(--atm-text-muted);
  text-decoration: none;
}

.article-breadcrumb a:hover {
  color: var(--atm-primary);
}

.article-breadcrumb .current {
  color: var(--atm-text);
}

.article-breadcrumb .sep {
  opacity: 0.45;
}

.article-head {
  margin-bottom: 28px;
}

.article-meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
  margin-bottom: 14px;
  font-size: 13px;
  color: var(--atm-text-muted);
}

.article-tag {
  padding: 4px 10px;
  font-weight: 700;
  color: var(--atm-primary);
  background: #f5f3ff;
  border-radius: 999px;
}

.article-title {
  margin: 0;
  font-size: clamp(24px, 4vw, 32px);
  font-weight: 800;
  line-height: 1.35;
  color: var(--atm-text);
}

.article-body p {
  margin: 0 0 1.1em;
  font-size: 16px;
  line-height: 1.75;
  color: var(--atm-text);
}

.article-foot {
  margin-top: 40px;
  padding-top: 24px;
  border-top: 1px solid #eef2ff;
}

.back-link {
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-primary);
  text-decoration: none;
}

.back-link:hover {
  text-decoration: underline;
}
</style>
