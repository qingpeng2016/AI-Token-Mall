<script setup lang="ts">
import { RouterLink } from 'vue-router'

const chatgptLinks = [
  { label: 'GPT Go', href: '#catalog' },
  { label: 'Plus 月卡', href: '#catalog' },
  { label: 'Pro 5X', href: '#catalog' },
  { label: 'Pro 20X', href: '#catalog' },
  { label: 'Codex', href: '#catalog' },
  { label: 'Images 2.5', href: '#catalog' },
]

const moreAiLinks = [
  { label: 'Claude Pro', href: '#catalog' },
  { label: 'Max 5X', href: '#catalog' },
  { label: 'Max 20X', href: '#catalog' },
  { label: 'Claude Code', href: '#catalog' },
  { label: 'Grok 套餐', href: '#catalog' },
  { label: 'Gemini 套餐', href: '#catalog' },
  { label: 'Cursor 配置', href: '#catalog' },
  { label: 'Perplexity Pro', href: '#catalog' },
  { label: '全部套餐', href: '#catalog' },
]

const toolsLinks = [
  { label: '工具中心', href: '#' },
  { label: 'Codex 使用说明', href: '#' },
  { label: 'Skills 库', href: '#' },
  { label: '价格中心', href: '#catalog' },
  { label: '价格说明', href: '#faq' },
  { label: '教程资讯', href: '#' },
]

const aboutLinks = [
  { label: '关于我们', href: '#' },
  { label: '联系我们', href: '#' },
  { label: '企业采购', href: '#faq' },
  { label: '分销合作', href: '#' },
  { label: '服务条款与隐私', href: '#' },
]
</script>

<template>
  <footer class="site-footer">
    <div class="atm-container footer-grid">
      <div class="footer-brand">
        <RouterLink to="/" class="footer-logo">AI Token Mall</RouterLink>
        <p class="footer-tagline">
          低价 AI 套餐 · ChatGPT / Claude / Cursor 等；支付宝 / 微信自助购买，团队对公走
          <a href="#faq">企业采购</a>。
        </p>
        <button type="button" class="footer-support">在线客服</button>
      </div>

      <div v-for="(col, idx) in [
        { title: 'CHATGPT', links: chatgptLinks },
        { title: '更多 AI', links: moreAiLinks },
        { title: '工具 · 帮助', links: toolsLinks },
        { title: '关于', links: aboutLinks },
      ]" :key="idx" class="footer-col">
        <h3 class="footer-col-title">{{ col.title }}</h3>
        <ul class="footer-col-list">
          <li v-for="link in col.links" :key="link.label">
            <a :href="link.href">{{ link.label }}</a>
          </li>
        </ul>
      </div>
    </div>

    <div class="footer-legal">
      <div class="atm-container footer-legal-inner">
        <p>
          © 2026 AI Token Mall · 独立第三方 AI 服务平台，与 OpenAI 等商标持有人无隶属或授权关系
        </p>
        <p class="footer-legal-muted">
          ChatGPT、OpenAI 为 OpenAI, Inc. 商标；名称仅作描述性（nominative）说明用途。
        </p>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.site-footer {
  background: #0c1222;
  color: #94a3b8;
  font-size: 13px;
  line-height: 1.5;
}

.footer-grid {
  display: grid;
  grid-template-columns: minmax(200px, 1.4fr) repeat(4, minmax(0, 1fr));
  gap: 32px 24px;
  padding: 48px 24px 40px;
}

@media (max-width: 1024px) {
  .footer-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
  .footer-brand {
    grid-column: 1 / -1;
    max-width: 420px;
  }
}

@media (max-width: 640px) {
  .footer-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 28px 20px;
    padding: 36px 16px 32px;
  }
}

.footer-logo {
  display: inline-block;
  margin-bottom: 12px;
  font-size: 20px;
  font-weight: 700;
  color: #f8fafc;
  text-decoration: none;
  letter-spacing: -0.02em;
}

.footer-tagline {
  margin: 0 0 16px;
  color: #94a3b8;
  line-height: 1.65;
}

.footer-tagline a {
  color: #cbd5e1;
  text-decoration: underline;
  text-underline-offset: 2px;
}

.footer-tagline a:hover {
  color: #fff;
}

.footer-support {
  padding: 10px 20px;
  font-size: 13px;
  font-weight: 500;
  color: #e2e8f0;
  cursor: pointer;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 999px;
  transition: background 0.15s, border-color 0.15s;
}

.footer-support:hover {
  background: #334155;
  border-color: #475569;
}

.footer-col-title {
  margin: 0 0 14px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: #64748b;
}

.footer-col-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.footer-col-list li + li {
  margin-top: 10px;
}

.footer-col-list a {
  color: #cbd5e1;
  text-decoration: none;
  transition: color 0.15s;
}

.footer-col-list a:hover {
  color: #fff;
}

.footer-legal {
  border-top: 1px solid #1e293b;
  background: #080d18;
}

.footer-legal-inner {
  padding: 20px 24px 28px;
  text-align: center;
}

.footer-legal p {
  margin: 0;
  font-size: 12px;
  color: #64748b;
  line-height: 1.6;
}

.footer-legal-muted {
  margin-top: 8px !important;
  font-size: 11px !important;
  color: #475569 !important;
}
</style>
