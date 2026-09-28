<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { useTutorialBlog } from '@/composables/useTutorialBlog'

const route = useRoute()
const router = useRouter()
const { pageTitle, categories, articles } = useTutorialBlog()

function parseCategory(q: unknown): string {
  const id = typeof q === 'string' ? q : 'all'
  return categories.value.some((c) => c.id === id) ? id : 'all'
}

const activeCategory = ref('all')

watch(
  categories,
  () => {
    activeCategory.value = parseCategory(route.query.cat)
  },
  { immediate: true },
)

watch(
  () => route.query.cat,
  (cat) => {
    activeCategory.value = parseCategory(cat)
  },
)

const filteredPosts = computed(() => {
  if (activeCategory.value === 'all') return articles.value
  return articles.value.filter((p) => p.category === activeCategory.value)
})

function setCategory(id: string) {
  activeCategory.value = id
  const query = id === 'all' ? {} : { cat: id }
  router.replace({ path: '/blog', query })
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
        <h1 class="blog-title">{{ pageTitle }}</h1>
      </header>

      <div class="blog-filters" role="tablist" aria-label="文章分类">
        <button
          v-for="cat in categories"
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

      <div v-if="filteredPosts.length" class="blog-grid">
        <RouterLink
          v-for="post in filteredPosts"
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
    </div>
  </div>
</template>

<style scoped>
.blog-page {
  padding: 28px 0 72px;
  background: var(--atm-bg-soft);
  min-height: 60vh;
}

.blog-container {
  max-width: 1080px;
}

.blog-breadcrumb {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
  margin-bottom: 20px;
  font-size: 14px;
  color: var(--atm-text-muted);
}

.blog-breadcrumb a {
  color: var(--atm-text-muted);
  text-decoration: none;
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
  margin-bottom: 28px;
}

.blog-title {
  margin: 0;
  font-size: clamp(26px, 4vw, 34px);
  font-weight: 800;
  color: var(--atm-text);
  letter-spacing: -0.02em;
}

.blog-filters {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 28px;
}

.filter-pill {
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-text-muted);
  background: #fff;
  border: 1px solid #e8ecf4;
  border-radius: 999px;
  cursor: pointer;
  transition:
    color 0.15s,
    background 0.15s,
    border-color 0.15s;
}

.filter-pill:hover {
  color: var(--atm-primary);
  border-color: #ddd6fe;
}

.filter-pill--active {
  color: #fff;
  background: var(--atm-gradient);
  border-color: transparent;
}

.blog-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 16px;
}

@media (min-width: 720px) {
  .blog-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
  }
}

.blog-card {
  display: block;
  padding: 22px 24px;
  text-decoration: none;
  background: #fff;
  border: 1px solid #eef2ff;
  border-radius: 16px;
  box-shadow: 0 4px 18px rgba(30, 27, 75, 0.05);
  transition:
    transform 0.15s,
    box-shadow 0.15s,
    border-color 0.15s;
}

.blog-card:hover {
  transform: translateY(-2px);
  border-color: #ddd6fe;
  box-shadow: 0 12px 28px rgba(124, 58, 237, 0.1);
}

.blog-card-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.blog-card-tag {
  padding: 3px 10px;
  font-size: 11px;
  font-weight: 700;
  color: var(--atm-primary);
  background: #f5f3ff;
  border-radius: 999px;
}

.blog-card-date {
  font-size: 12px;
  color: var(--atm-text-muted);
}

.blog-card-title {
  margin: 0 0 10px;
  font-size: 17px;
  font-weight: 700;
  line-height: 1.4;
  color: var(--atm-text);
}

.blog-card-excerpt {
  margin: 0;
  font-size: 14px;
  line-height: 1.55;
  color: var(--atm-text-muted);
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.blog-empty {
  padding: 48px 0;
  text-align: center;
  color: var(--atm-text-muted);
}
</style>
