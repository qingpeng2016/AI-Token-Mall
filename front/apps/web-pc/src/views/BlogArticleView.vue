<script setup lang="ts">
import { computed, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { getBlogPost } from '@/mocks/blog'

const route = useRoute()
const router = useRouter()

const slug = computed(() => String(route.params.slug ?? ''))
const post = computed(() => getBlogPost(slug.value))

watch(
  () => route.params.slug,
  () => {
    if (!post.value) router.replace('/blog')
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
  padding-bottom: 24px;
  border-bottom: 1px solid rgba(124, 58, 237, 0.12);
}

.article-meta {
  display: flex;
  align-items: center;
  gap: 14px;
  margin-bottom: 16px;
  font-size: 13px;
  color: #94a3b8;
}

.article-tag {
  padding: 4px 10px;
  font-weight: 600;
  color: var(--atm-primary-dark);
  background: var(--atm-primary-light);
  border-radius: 999px;
}

.article-title {
  margin: 0;
  font-size: clamp(1.5rem, 4vw, 2rem);
  font-weight: 800;
  line-height: 1.35;
  letter-spacing: -0.03em;
  color: var(--atm-text);
}

.article-body {
  font-size: 16px;
  line-height: 1.75;
  color: #334155;
}

.article-body p {
  margin: 0 0 1.15em;
}

.article-foot {
  margin-top: 40px;
  padding-top: 24px;
  border-top: 1px solid #f1f5f9;
}

.back-link {
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-primary);
  text-decoration: none;
}

.back-link:hover {
  color: var(--atm-primary-dark);
}
</style>
