<script setup lang="ts">
import { nextTick, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { isNavDropdown, navBrandDropdowns, navMegaMenu } from '@/mocks/nav'

const megaOpen = ref(false)
const openDrop = ref<string | null>(null)
const megaTriggerRef = ref<HTMLElement | null>(null)
const megaArrowLeft = ref(46)

function openMega() {
  megaOpen.value = true
  openDrop.value = null
  nextTick(() => {
    if (megaTriggerRef.value) {
      megaArrowLeft.value = megaTriggerRef.value.offsetWidth / 2
    }
  })
}

function closeAll() {
  megaOpen.value = false
  openDrop.value = null
}

function openBrandDrop(id: string) {
  openDrop.value = id
  megaOpen.value = false
}
</script>

<template>
  <header class="site-header">
    <div class="atm-container header-row">
      <RouterLink to="/" class="logo">
        <span class="logo-mark">AI</span>
        <span class="logo-text">Token Mall</span>
      </RouterLink>

      <div class="nav-shell" @mouseleave="closeAll">
        <nav class="main-nav" aria-label="主导航">
          <div class="nav-item nav-item--mega" @mouseenter="openMega">
            <a
              ref="megaTriggerRef"
              href="#catalog"
              class="nav-link nav-link--mega"
              :class="{ 'nav-link--active': megaOpen }"
            >
              套餐购买
              <span class="nav-chevron" aria-hidden="true">▾</span>
            </a>

            <div
              v-show="megaOpen"
              class="mega-popover"
              @mouseenter="openMega"
            >
              <div
                class="mega-popover-arrow"
                aria-hidden="true"
                :style="{ left: `${megaArrowLeft}px` }"
              />
              <div class="mega-grid">
                <div v-for="col in navMegaMenu" :key="col.title" class="mega-col">
                  <h3 class="mega-col-title">{{ col.title }}</h3>
                  <ul class="mega-list">
                    <li v-for="item in col.items" :key="item.label">
                      <a
                        :href="item.href"
                        class="mega-link"
                        :class="{ 'mega-link--featured': item.featured }"
                      >
                        <span class="mega-link-label">
                          {{ item.label }}
                          <span v-if="item.featured" class="mega-hot">热门</span>
                        </span>
                        <span class="mega-link-price">{{ item.price }}</span>
                      </a>
                    </li>
                  </ul>
                </div>
              </div>
              <a href="#catalog" class="mega-footer">全部套餐</a>
            </div>
          </div>

          <template v-for="brand in navBrandDropdowns" :key="brand.id">
            <RouterLink
              v-if="!isNavDropdown(brand)"
              :to="brand.to"
              class="nav-link"
            >
              {{ brand.label }}
            </RouterLink>
            <div
              v-else
              class="nav-item"
              @mouseenter="openBrandDrop(brand.id)"
            >
              <button
                type="button"
                class="nav-link nav-link-btn"
                :class="{ 'nav-link--active': openDrop === brand.id }"
              >
                {{ brand.label }}
                <span v-if="brand.hotSale" class="nav-hot-tag">热销</span>
                <span class="nav-chevron" aria-hidden="true">▾</span>
              </button>
              <div v-show="openDrop === brand.id" class="nav-drop-panel">
                <a
                  v-for="item in brand.items"
                  :key="item.label"
                  :href="item.href"
                  class="nav-drop-link"
                  :class="{ 'nav-drop-link--featured': item.featured }"
                >
                  <span class="nav-drop-label">
                    {{ item.label }}
                    <span v-if="item.featured" class="nav-drop-hot">热门</span>
                  </span>
                  <span v-if="item.price" class="nav-drop-price">{{ item.price }}</span>
                </a>
              </div>
            </div>
          </template>

          <a href="#faq" class="nav-link nav-link-plain">教程 FAQ</a>
        </nav>
      </div>

      <div class="actions">
        <RouterLink to="/login" class="link-muted">登录</RouterLink>
        <RouterLink to="/register" class="atm-btn-primary btn-sm">注册</RouterLink>
      </div>
    </div>
  </header>
</template>

<style scoped>
.site-header {
  position: sticky;
  top: 0;
  z-index: 200;
  background: var(--atm-bg);
  border-bottom: 1px solid #ede9fe;
}

.header-row {
  display: flex;
  align-items: center;
  gap: 16px;
  height: 56px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 8px;
  text-decoration: none;
  flex-shrink: 0;
}

.logo-mark {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 800;
  color: #fff;
  background: var(--atm-gradient);
  border-radius: 8px;
}

.logo-text {
  font-size: 17px;
  font-weight: 700;
  color: var(--atm-text);
}

.nav-shell {
  position: relative;
  flex: 1;
  display: none;
  justify-content: flex-start;
  min-width: 0;
}

@media (min-width: 1024px) {
  .nav-shell {
    display: flex;
  }
}

.main-nav {
  display: flex;
  align-items: center;
  gap: 2px;
}

.nav-item {
  position: relative;
}

.nav-item--mega {
  position: relative;
}

.nav-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  font-size: 14px;
  font-weight: 500;
  color: var(--atm-text);
  text-decoration: none;
  border-radius: 999px;
  white-space: nowrap;
  transition:
    background 0.15s,
    color 0.15s;
}

.nav-link-btn {
  border: none;
  background: transparent;
  cursor: pointer;
  font-family: inherit;
}

.nav-link-plain {
  color: var(--atm-text-muted);
}

.nav-link:hover,
.nav-link--active {
  color: var(--atm-primary);
  background: #ede9fe;
}

.nav-hot-tag {
  padding: 1px 6px;
  font-size: 10px;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(90deg, #f97316, #fb923c);
  border-radius: 999px;
  line-height: 1.4;
}

.nav-chevron {
  font-size: 10px;
  color: var(--atm-text-muted);
}

.nav-drop-panel {
  position: absolute;
  top: calc(100% + 8px);
  left: 50%;
  transform: translateX(-50%);
  min-width: 240px;
  padding: 6px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 14px;
  box-shadow:
    0 4px 6px rgba(30, 27, 75, 0.04),
    0 20px 48px rgba(30, 27, 75, 0.14);
  z-index: 210;
}

.nav-drop-panel::before {
  content: '';
  position: absolute;
  top: -6px;
  left: 50%;
  transform: translateX(-50%) rotate(45deg);
  width: 10px;
  height: 10px;
  background: #fff;
  border-left: 1px solid #e8eaf0;
  border-top: 1px solid #e8eaf0;
}

.nav-drop-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 9px 12px;
  font-size: 13px;
  color: var(--atm-text);
  text-decoration: none;
  border-radius: 8px;
}

.nav-drop-link:hover {
  background: #f5f3ff;
  color: var(--atm-primary);
}

.nav-drop-link--featured {
  color: #fff;
  background: var(--atm-gradient);
}

.nav-drop-link--featured:hover {
  color: #fff;
  filter: brightness(1.03);
}

.nav-drop-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-weight: 500;
}

.nav-drop-hot {
  padding: 1px 5px;
  font-size: 10px;
  font-weight: 700;
  background: rgba(255, 255, 255, 0.22);
  border-radius: 999px;
}

.nav-drop-price {
  flex-shrink: 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-primary);
  white-space: nowrap;
}

.nav-drop-link--featured .nav-drop-price {
  color: rgba(255, 255, 255, 0.95);
}

/* 浮层弹窗：锚定在「套餐购买」下，箭头对准触发按钮 */
.mega-popover {
  position: absolute;
  top: calc(100% + 10px);
  left: 0;
  width: min(1064px, calc(100vw - 40px));
  padding: 22px 26px 18px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow:
    0 4px 8px rgba(30, 27, 75, 0.04),
    0 24px 64px rgba(30, 27, 75, 0.16);
  z-index: 210;
}

.mega-popover::before {
  content: '';
  position: absolute;
  top: -18px;
  left: 0;
  right: 0;
  height: 18px;
}

.mega-popover-arrow {
  position: absolute;
  top: -6px;
  width: 12px;
  height: 12px;
  background: #fff;
  border-left: 1px solid #e8eaf0;
  border-top: 1px solid #e8eaf0;
  transform: translateX(-50%) rotate(45deg);
}

.mega-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(158px, 1fr));
  gap: 18px 28px;
}

@media (min-width: 1100px) {
  .mega-grid {
    grid-template-columns: repeat(5, minmax(152px, 1fr));
    gap: 16px 22px;
  }
}

.mega-col-title {
  margin: 0 0 8px;
  padding: 0 4px;
  font-size: 13px;
  font-weight: 700;
  color: var(--atm-text);
}

.mega-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.mega-col {
  min-width: 0;
}

.mega-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  padding: 7px 8px;
  font-size: 13px;
  line-height: 1.4;
  color: #64748b;
  text-decoration: none;
  border-radius: 10px;
  transition:
    background 0.12s,
    color 0.12s;
}

.mega-link:hover {
  background: #f8fafc;
  color: var(--atm-text);
}

.mega-link--featured {
  margin-bottom: 4px;
  padding: 8px 10px;
  color: #fff;
  background: var(--atm-gradient);
  box-shadow: 0 4px 14px rgba(124, 58, 237, 0.35);
}

.mega-link--featured:hover {
  color: #fff;
  filter: brightness(1.03);
}

.mega-link-label {
  display: inline-flex;
  align-items: center;
  flex: 1;
  flex-wrap: wrap;
  gap: 4px;
  min-width: 0;
  font-weight: 500;
  color: var(--atm-text);
}

.mega-link--featured .mega-link-label {
  font-weight: 600;
  color: #fff;
}

.mega-link-price {
  flex-shrink: 0;
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-primary);
  white-space: nowrap;
  padding-left: 4px;
}

.mega-link--featured .mega-link-price {
  color: rgba(255, 255, 255, 0.95);
}

.mega-hot {
  padding: 1px 5px;
  font-size: 10px;
  font-weight: 700;
  line-height: 1.3;
  color: #fff;
  background: rgba(255, 255, 255, 0.22);
  border-radius: 999px;
}

.mega-footer {
  display: inline-block;
  margin-top: 14px;
  padding-left: 4px;
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-primary);
  text-decoration: none;
}

.mega-footer:hover {
  text-decoration: underline;
}

.actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-left: auto;
  flex-shrink: 0;
}

.link-muted {
  font-size: 14px;
  color: var(--atm-text-muted);
  text-decoration: none;
}

.link-muted:hover {
  color: var(--atm-primary);
}

.btn-sm {
  padding: 8px 16px;
  font-size: 14px;
}

</style>
