<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  blogCategories,
  blogPageMeta,
  blogPosts,
  type BlogCategoryId,
} from '@/mocks/blog'

const PAGE_SIZE = 12

const route = useRoute()
const router = useRouter()

function parseCategory(q: unknown): BlogCategoryId {
  const id = typeof q === 'string' ? q : 'all'
  return blogCategories.some((c) => c.id === id) ? (id as BlogCategoryId) : 'all'
}

const activeCategory = ref<BlogCategoryId>(parseCategory(route.query.cat))
const visibleCount = ref(PAGE_SIZE)

watch(
  () => route.query.cat,
  (cat) => {
    activeCategory.value = parseCategory(cat)
    visibleCount.value = PAGE_SIZE
  },
)

const filteredPosts = computed(() => {
  if (activeCategory.value === 'all') return blogPosts
  return blogPosts.filter((p) => p.category === activeCategory.value)
})

const visiblePosts = computed(() => filteredPosts.value.slice(0, visibleCount.value))

const remaining = computed(() =>
  Math.max(0, filteredPosts.value.length - visibleCount.value),
)

function setCategory(id: BlogCategoryId) {
  activeCategory.value = id
  visibleCount.value = PAGE_SIZE
  const query = id === 'all' ? {} : { cat: id }
  router.replace({ path: '/blog', query })
}

function loadMore() {
  visibleCount.value += PAGE_SIZE
}
</script>

<template>
  <div class="blog-page">
    <div class="atm-container blog-container">
      <nav class="blog-breadcrumb" aria-label="面包屑">
        <RouterLink to="/">首页</RouterLink>
        <span class="sep">/</span>
        <span class="current">教程</span>
      </nav>

      <header class="blog-hero">
        <h1 class="blog-title">{{ blogPageMeta.title }}</h1>
        <p class="blog-subtitle">{{ blogPageMeta.subtitle }}</p>
      </header>

      <div class="blog-filters" role="tablist" aria-label="文章分类">
        <button
          v-for="cat in blogCategories"
          :key="cat.id"
          type="button"
          class="filter-pill"
          :class="{ 'filter-pill--active': activeCategory === cat.id }"
          role="tab"
          :aria-selected="activeCategory === cat.id"
          @click="setCategory(cat.id)"
        >
          {{ cat.label }}
        </button>
      </div>

      <div v-if="visiblePosts.length" class="blog-grid">
        <RouterLink
          v-for="post in visiblePosts"
          :key="post.slug"
          :to="{ name: 'blog-article', params: { slug: post.slug } }"
          class="blog-card"
        >
          <div class="blog-card-meta">
            <span class="blog-card-tag">{{ post.categoryLabel }}</span>
            <time class="blog-card-date" :datetime="post.date">{{ post.date }}</time>
          </div>
          <h2 class="blog-card-title">{{ post.title }}</h2>
          <p class="blog-card-excerpt">{{ post.excerpt }}</p>
        </RouterLink>
      </div>

      <p v-else class="blog-empty">该分类暂无文章，试试「全部」。</p>

      <div v-if="remaining > 0" class="blog-load-wrap">
        <button type="button" class="blog-load-more" @click="loadMore">
          加载更多（剩余 {{ remaining }} 篇）
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.blog-page {
  padding: 28px 0 72px;
}

.blog-container {
  max-width: 960px;
}

.blog-breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 28px;
  font-size: 14px;
  color: var(--atm-text-muted);
}

.blog-breadcrumb a {
  color: var(--atm-text-muted);
  text-decoration: none;
  transition: color 0.15s ease;
}

.blog-breadcrumb a:hover {
  color: var(--atm-primary);
}

.blog-breadcrumb .current {
  color: var(--atm-text);
}

.blog-breadcrumb .sep {
  opacity: 0.45;
}

.blog-hero {
  text-align: center;
  margin-bottom: 32px;
}

.blog-title {
  margin: 0 0 14px;
  font-size: clamp(1.75rem, 4vw, 2.25rem);
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1.2;
  color: var(--atm-primary-dark);
}

.blog-subtitle {
  margin: 0 auto;
  max-width: 36em;
  font-size: 15px;
  line-height: 1.65;
  color: var(--atm-text-muted);
}

.blog-filters {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 10px;
  margin-bottom: 36px;
}

.filter-pill {
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.2;
  color: var(--atm-text);
  background: #fff;
  border: 1px solid rgba(124, 58, 237, 0.14);
  border-radius: 999px;
  cursor: pointer;
  transition:
    background 0.15s ease,
    color 0.15s ease,
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.filter-pill:hover {
  border-color: rgba(124, 58, 237, 0.28);
  box-shadow: 0 4px 14px rgba(124, 58, 237, 0.08);
}

.filter-pill--active {
  color: #fff;
  background: var(--atm-gradient);
  border-color: transparent;
  box-shadow: 0 6px 20px rgba(124, 58, 237, 0.25);
}

.blog-grid {
  display: grid;
  gap: 20px;
}

@media (min-width: 768px) {
  .blog-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 22px 24px;
  }
}

.blog-card {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-height: 100%;
  padding: 22px 22px 20px;
  text-decoration: none;
  color: inherit;
  background: #fff;
  border: 1px solid rgba(124, 58, 237, 0.1);
  border-radius: 16px;
  box-shadow: 0 6px 24px rgba(124, 58, 237, 0.06);
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease,
    transform 0.15s ease;
}

.blog-card:hover {
  border-color: rgba(124, 58, 237, 0.22);
  box-shadow: 0 12px 32px rgba(124, 58, 237, 0.1);
  transform: translateY(-2px);
}

.blog-card-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.blog-card-tag {
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-primary-dark);
  background: var(--atm-primary-light);
  border-radius: 999px;
}

.blog-card-date {
  flex-shrink: 0;
  font-size: 12px;
  color: #94a3b8;
}

.blog-card-title {
  margin: 0;
  font-size: 17px;
  font-weight: 800;
  line-height: 1.45;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.blog-card-excerpt {
  margin: 0;
  font-size: 14px;
  line-height: 1.65;
  color: var(--atm-text-muted);
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  overflow: hidden;
}

.blog-empty {
  text-align: center;
  padding: 48px 16px;
  color: var(--atm-text-muted);
}

.blog-load-wrap {
  display: flex;
  justify-content: center;
  margin-top: 40px;
}

.blog-load-more {
  min-width: min(100%, 320px);
  padding: 14px 28px;
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-text);
  background: #fff;
  border: 1px solid #d1d5db;
  border-radius: 12px;
  cursor: pointer;
  transition:
    border-color 0.15s ease,
    box-shadow 0.15s ease;
}

.blog-load-more:hover {
  border-color: rgba(124, 58, 237, 0.35);
  box-shadow: 0 6px 20px rgba(124, 58, 237, 0.08);
}
</style>
