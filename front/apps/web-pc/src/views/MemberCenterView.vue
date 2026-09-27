<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { formatCnyFromCents } from '@ai-token-mall/shared'
import MemberSidebar from '@/components/member/MemberSidebar.vue'
import {
  clearSessionUser,
  getSessionUser,
  isLoggedIn,
  userAccountLabel,
} from '@/composables/useSessionUser'
import { userApi } from '@/api'
import {
  formatTokens,
  memberNav,
  mockApiKeys,
  mockInvoices,
  mockMemberOverview,
  mockOrders,
  mockSubAccounts,
  mockSubscriptions,
  mockWalletTx,
  orderStatusLabel,
  type MemberTab,
} from '@/mocks/member'

const router = useRouter()
const route = useRoute()
const user = ref(getSessionUser())

const validTabs = new Set(memberNav.map((n) => n.id))

const activeTab = computed<MemberTab>(() => {
  const q = route.query.tab
  const t = typeof q === 'string' ? q : 'overview'
  return validTabs.has(t as MemberTab) ? (t as MemberTab) : 'overview'
})

const pageTitle = computed(() => memberNav.find((n) => n.id === activeTab.value)?.label ?? '会员中心')

const displayName = computed(() =>
  user.value ? userAccountLabel(user.value) : '会员',
)

onMounted(() => {
  const profile = getSessionUser()
  if (!isLoggedIn() || !profile) {
    clearSessionUser()
    router.replace({ path: '/login', query: { redirect: '/member' } })
    return
  }
  user.value = profile
})

function usagePercent(used: number, limit: number) {
  return limit > 0 ? Math.min(100, Math.round((used / limit) * 100)) : 0
}

function logout() {
  clearSessionUser()
  void userApi.logout()
  ElMessage.success('已退出登录')
  router.push('/')
}

function mockRecharge() {
  ElMessage.info('余额充值接口对接中，可先使用支付宝 / 微信直接购套餐')
}

function mockAction(msg: string) {
  ElMessage.info(msg)
}
</script>

<template>
  <div v-if="user" class="member-layout">
    <section class="member-hero">
      <div class="member-hero-bg" aria-hidden="true" />
      <div class="atm-container member-hero-inner">
        <div class="member-hero-copy">
          <span class="member-hero-badge">AI Plan · 会员中心</span>
          <h1 class="member-hero-title">你好，{{ displayName }}</h1>
        </div>
        <div class="member-hero-actions">
          <RouterLink to="/#catalog" class="hero-btn hero-btn--light">选购套餐</RouterLink>
          <button type="button" class="hero-btn hero-btn--ghost" @click="mockRecharge">
            余额充值
          </button>
        </div>
      </div>
    </section>

    <div class="atm-container member-body">
      <div class="member-grid">
        <div class="member-sidebar-wrap">
          <MemberSidebar :active="activeTab" :user="user" />
        </div>

        <main class="member-main">
          <section class="panel">
            <header class="panel-head">
              <h2 class="panel-head-title">{{ pageTitle }}</h2>
            </header>

            <!-- 概览 -->
            <div v-if="activeTab === 'overview'" class="panel-body">
            <div class="stat-row">
              <article class="stat-card stat-card--balance">
                <div class="stat-main">
                  <span class="stat-label">账户余额</span>
                  <strong class="stat-value">{{
                    formatCnyFromCents(mockMemberOverview.balanceCents)
                  }}</strong>
                </div>
                <button type="button" class="stat-link" @click="mockRecharge">去充值</button>
              </article>
              <article class="stat-card">
                <div class="stat-main">
                  <span class="stat-label">生效中套餐</span>
                  <strong class="stat-value">{{ mockMemberOverview.activePlans }}</strong>
                </div>
                <RouterLink :to="{ path: '/member', query: { tab: 'plans' } }" class="stat-link"
                  >查看</RouterLink
                >
              </article>
              <article class="stat-card">
                <div class="stat-main">
                  <span class="stat-label">待支付订单</span>
                  <strong class="stat-value">{{ mockMemberOverview.pendingOrders }}</strong>
                </div>
                <RouterLink :to="{ path: '/member', query: { tab: 'orders' } }" class="stat-link"
                  >查看</RouterLink
                >
              </article>
            </div>

            <h2 class="panel-subtitle">套餐用量</h2>
            <div class="plan-mini-list">
              <article
                v-for="sub in mockSubscriptions"
                :key="sub.id"
                class="plan-mini"
              >
                <div class="plan-mini-head">
                  <strong>{{ sub.productName }}</strong>
                  <span>至 {{ sub.expiresAt }}</span>
                </div>
                <div class="progress-track">
                  <div
                    class="progress-fill"
                    :style="{ width: `${usagePercent(sub.usedTokens, sub.limitTokens)}%` }"
                  />
                </div>
                <p class="progress-meta">
                  已用 {{ formatTokens(sub.usedTokens) }} / {{ formatTokens(sub.limitTokens) }}
                </p>
              </article>
            </div>

            <h2 class="panel-subtitle">最近订单</h2>
            <div class="table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>订单号</th>
                    <th>商品</th>
                    <th>金额</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="o in mockOrders.slice(0, 3)" :key="o.orderNo">
                    <td>{{ o.orderNo }}</td>
                    <td>{{ o.productName }}</td>
                    <td>{{ formatCnyFromCents(o.totalCents) }}</td>
                    <td>
                      <span class="tag" :class="`tag--${o.status}`">{{
                        orderStatusLabel[o.status]
                      }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            </div>

            <!-- 我的套餐 -->
            <div v-else-if="activeTab === 'plans'" class="panel-body">
            <p class="panel-desc">
              展示当前订阅的服务有效期与本周期 token 用量；续费或升档请前往
              <RouterLink to="/#catalog">套餐购买</RouterLink>。
            </p>
            <article v-for="sub in mockSubscriptions" :key="sub.id" class="plan-card">
              <div class="plan-card-head">
                <div>
                  <h2>{{ sub.productName }}</h2>
                  <span class="sku">{{ sub.skuLabel }}</span>
                </div>
                <span class="tag tag--active">使用中</span>
              </div>
              <div class="plan-metrics">
                <div>
                  <span class="metric-label">本周期额度</span>
                  <strong>{{ formatTokens(sub.limitTokens) }} tokens</strong>
                </div>
                <div>
                  <span class="metric-label">已使用</span>
                  <strong>{{ formatTokens(sub.usedTokens) }}</strong>
                </div>
                <div>
                  <span class="metric-label">周期截止</span>
                  <strong>{{ sub.periodEnd }}</strong>
                </div>
                <div>
                  <span class="metric-label">服务到期</span>
                  <strong>{{ sub.expiresAt }}</strong>
                </div>
              </div>
              <div class="progress-track progress-track--lg">
                <div
                  class="progress-fill"
                  :style="{ width: `${usagePercent(sub.usedTokens, sub.limitTokens)}%` }"
                />
              </div>
              <div class="plan-actions">
                <RouterLink to="/#catalog" class="atm-btn-ghost btn-xs">续费 / 升档</RouterLink>
                <button type="button" class="link-btn" @click="mockAction('加购 token 包即将上线')">
                  加购额度
                </button>
              </div>
            </article>
            <p v-if="!mockSubscriptions.length" class="empty">暂无生效套餐，去首页选购吧。</p>
            </div>

            <!-- 我的订单 -->
            <div v-else-if="activeTab === 'orders'" class="panel-body">
            <div class="table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>订单号</th>
                    <th>商品</th>
                    <th>数量</th>
                    <th>金额</th>
                    <th>开票</th>
                    <th>状态</th>
                    <th>时间</th>
                    <th />
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="o in mockOrders" :key="o.orderNo">
                    <td class="mono">{{ o.orderNo }}</td>
                    <td>{{ o.productName }}</td>
                    <td>{{ o.quantity }}</td>
                    <td>{{ formatCnyFromCents(o.totalCents) }}</td>
                    <td>{{ o.enterpriseInvoice ? '企业' : '—' }}</td>
                    <td>
                      <span class="tag" :class="`tag--${o.status}`">{{
                        orderStatusLabel[o.status]
                      }}</span>
                    </td>
                    <td class="muted">{{ o.createdAt }}</td>
                    <td>
                      <button
                        v-if="o.status === 'pending_payment'"
                        type="button"
                        class="link-btn"
                        @click="mockAction('继续支付对接中')"
                      >
                        去支付
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            </div>

            <!-- 账户余额 -->
            <div v-else-if="activeTab === 'account'" class="panel-body">
            <article class="balance-card">
              <div>
                <span class="stat-label">可用余额</span>
                <p class="balance-amount">
                  {{ formatCnyFromCents(mockMemberOverview.balanceCents) }}
                </p>
                <p class="panel-desc">
                  余额可用于快速下单部分套餐；也可直接使用支付宝 / 微信支付。
                </p>
              </div>
              <button type="button" class="hero-btn hero-btn--light" @click="mockRecharge">
                充值
              </button>
            </article>

            <h2 class="panel-subtitle">余额流水</h2>
            <div class="table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>类型</th>
                    <th>金额</th>
                    <th>说明</th>
                    <th>时间</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="tx in mockWalletTx" :key="tx.id">
                    <td>
                      {{
                        tx.type === 'recharge'
                          ? '充值'
                          : tx.type === 'refund'
                            ? '退款'
                            : '消费'
                      }}
                    </td>
                    <td :class="tx.amountCents > 0 ? 'amount-plus' : 'amount-minus'">
                      {{ tx.amountCents > 0 ? '+' : ''
                      }}{{ formatCnyFromCents(Math.abs(tx.amountCents)) }}
                    </td>
                    <td>{{ tx.remark }}</td>
                    <td class="muted">{{ tx.createdAt }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            </div>

            <!-- API 密钥 -->
            <div v-else-if="activeTab === 'api-keys'" class="panel-body">
            <p class="panel-desc">
              调用平台 API 时使用；密钥仅创建时完整展示一次，请妥善保存。
            </p>
            <button type="button" class="atm-btn-primary btn-xs" @click="mockAction('创建密钥 API 对接中')">
              创建新密钥
            </button>
            <div class="table-wrap table-wrap--spaced">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>名称</th>
                    <th>密钥</th>
                    <th>绑定套餐</th>
                    <th>最近使用</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="k in mockApiKeys" :key="k.id">
                    <td>{{ k.name }}</td>
                    <td class="mono">{{ k.prefix }}</td>
                    <td>{{ k.subscriptionName }}</td>
                    <td class="muted">{{ k.lastUsedAt ?? '—' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            </div>

            <!-- 发票 -->
            <div v-else-if="activeTab === 'invoices'" class="panel-body">
            <p class="panel-desc">
              下单时勾选「企业开票」的订单可在此申请与下载；也可在
              <RouterLink to="/enterprise#invoice">企业采购</RouterLink> 查看开票说明。
            </p>
            <div class="table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>订单号</th>
                    <th>抬头</th>
                    <th>金额</th>
                    <th>状态</th>
                    <th>申请时间</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="inv in mockInvoices" :key="inv.id">
                    <td class="mono">{{ inv.orderNo }}</td>
                    <td>{{ inv.title }}</td>
                    <td>{{ formatCnyFromCents(inv.amountCents) }}</td>
                    <td>
                      <span class="tag tag--completed">{{
                        inv.status === 'issued' ? '已开具' : '处理中'
                      }}</span>
                    </td>
                    <td class="muted">{{ inv.createdAt }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <button type="button" class="link-btn" @click="mockAction('发票抬头管理对接中')">
              管理发票抬头 →
            </button>
            </div>

            <!-- 子账号 -->
            <div v-else-if="activeTab === 'sub-accounts'" class="panel-body">
            <p class="panel-desc">
              主账号购买套餐后，可为团队成员开通子账号，共享额度并单独统计用量。
            </p>
            <button type="button" class="atm-btn-primary btn-xs" @click="mockAction('邀请子账号功能开发中')">
              邀请子账号
            </button>
            <div class="table-wrap table-wrap--spaced">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>昵称</th>
                    <th>邮箱</th>
                    <th>用量</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="s in mockSubAccounts" :key="s.id">
                    <td>{{ s.nickname }}</td>
                    <td>{{ s.email }}</td>
                    <td>
                      {{ formatTokens(s.usedTokens) }} /
                      {{ formatTokens(s.limitTokens) }}
                    </td>
                    <td>
                      <span class="tag tag--active">{{
                        s.status === 'active' ? '正常' : '停用'
                      }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            </div>

            <!-- 账户设置 -->
            <div v-else-if="activeTab === 'settings'" class="panel-body">
            <dl class="settings-list">
              <div class="settings-row">
                <dt>登录账号</dt>
                <dd>{{ userAccountLabel(user) }}</dd>
              </div>
              <div class="settings-row">
                <dt>昵称</dt>
                <dd>{{ user.nickname || '未设置' }}</dd>
              </div>
              <div class="settings-row">
                <dt>密码</dt>
                <dd>
                  <button type="button" class="link-btn" @click="mockAction('修改密码 API 对接中')">
                    修改密码
                  </button>
                </dd>
              </div>
            </dl>
            <div class="settings-actions">
              <RouterLink to="/#faq" class="atm-btn-ghost btn-xs">帮助与 FAQ</RouterLink>
              <button type="button" class="btn-logout" @click="logout">退出登录</button>
            </div>
            </div>
          </section>
        </main>
      </div>
    </div>
  </div>
</template>

<style scoped>
.member-layout {
  background: var(--atm-bg);
  min-height: calc(100vh - 56px);
}

.member-hero {
  position: relative;
  overflow: hidden;
  padding: 44px 0 88px;
  color: #fff;
}

.member-hero-bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(125deg, #1e1b4b 0%, #5b21b6 45%, #6366f1 100%);
}

.member-hero-bg::after {
  content: '';
  position: absolute;
  inset: 0;
  opacity: 0.15;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.12) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.12) 1px, transparent 1px);
  background-size: 40px 40px;
}

.member-hero-inner {
  position: relative;
  z-index: 1;
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px 32px;
}

.member-hero-badge {
  display: inline-block;
  margin-bottom: 14px;
  padding: 6px 14px;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: rgba(255, 255, 255, 0.92);
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: 999px;
}

.member-hero-title {
  margin: 0 0 10px;
  font-size: clamp(28px, 4vw, 36px);
  font-weight: 800;
  letter-spacing: -0.03em;
  line-height: 1.15;
}

.member-hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.hero-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 12px 26px;
  font-size: 15px;
  font-weight: 600;
  text-decoration: none;
  border-radius: 999px;
  cursor: pointer;
  border: none;
  transition:
    transform 0.15s,
    box-shadow 0.15s;
}

.hero-btn--light {
  color: var(--atm-primary);
  background: #fff;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.hero-btn--light:hover {
  transform: translateY(-1px);
}

.hero-btn--ghost {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.35);
}

.member-body {
  position: relative;
  z-index: 2;
  margin-top: -52px;
  padding: 0 0 64px;
}

.member-grid {
  display: flex;
  flex-direction: column;
  gap: 28px;
}

@media (min-width: 960px) {
  .member-grid {
    flex-direction: row;
    align-items: stretch;
  }

  .member-sidebar-wrap {
    flex-shrink: 0;
    width: 272px;
    align-self: stretch;
  }
}

.member-main {
  flex: 1;
  min-width: 0;
}

.panel-head {
  margin: 0 0 24px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f1f5f9;
}

.panel-head-title {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.panel-body {
  min-height: 200px;
}

.panel {
  padding: 32px 28px;
  background: #fff;
  border: 1px solid rgba(237, 233, 254, 0.9);
  border-radius: 20px;
  box-shadow:
    0 4px 24px rgba(124, 58, 237, 0.06),
    0 16px 48px rgba(30, 27, 75, 0.06);
}

.panel-desc {
  margin: 0 0 20px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--atm-text-muted);
}

.panel-desc a {
  color: var(--atm-primary);
  font-weight: 500;
}

.panel-subtitle {
  margin: 28px 0 14px;
  font-size: 17px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.panel-subtitle:first-child {
  margin-top: 0;
}

.stat-row {
  display: grid;
  gap: 16px;
  padding: 4px 0;
}

@media (min-width: 640px) {
  .stat-row {
    grid-template-columns: repeat(3, 1fr);
    gap: 20px;
  }
}

.stat-card {
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  gap: 16px;
  min-height: 112px;
  padding: 18px 22px 16px;
  background: #fff;
  border: 1px solid rgba(124, 58, 237, 0.1);
  border-radius: 16px;
  box-shadow: 0 6px 24px rgba(124, 58, 237, 0.06);
  transition:
    background 0.15s ease,
    box-shadow 0.15s ease;
}

.stat-card--balance {
  background: linear-gradient(160deg, #faf5ff 0%, #fff 70%);
}

.stat-card:hover {
  background: #fdfcff;
  box-shadow: 0 10px 28px rgba(124, 58, 237, 0.1);
}

.stat-card--balance:hover {
  background: linear-gradient(160deg, #f3ebff 0%, #fff 70%);
}

.stat-main {
  min-width: 0;
}

.stat-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--atm-text-muted);
}

.stat-value {
  display: block;
  margin-top: 6px;
  font-size: 28px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.03em;
  line-height: 1.15;
  color: var(--atm-text);
}

.stat-card--balance .stat-value {
  color: var(--atm-primary-dark);
}

.stat-link {
  align-self: flex-start;
  padding: 6px 14px;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.2;
  color: var(--atm-primary-dark);
  background: var(--atm-primary-light);
  border: none;
  border-radius: 999px;
  cursor: pointer;
  text-decoration: none;
  transition:
    background 0.15s ease,
    color 0.15s ease;
}

.stat-link:hover {
  color: #fff;
  background: var(--atm-primary);
}

.plan-mini-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.plan-mini {
  padding: 18px 20px;
  background: #f8fafc;
  border: 1px solid #f1f5f9;
  border-radius: 14px;
}

.plan-mini-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
  font-size: 14px;
}

.plan-mini-head strong {
  color: var(--atm-text);
}

.plan-mini-head span {
  font-size: 12px;
  color: var(--atm-text-muted);
}

.progress-track {
  height: 8px;
  overflow: hidden;
  background: #e2e8f0;
  border-radius: 999px;
}

.progress-track--lg {
  height: 10px;
  margin-top: 16px;
}

.progress-fill {
  height: 100%;
  background: var(--atm-gradient);
  border-radius: 999px;
}

.progress-meta {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--atm-text-muted);
}

.plan-card {
  padding: 28px 26px;
  margin-bottom: 20px;
  border: 1px solid #ddd6fe;
  border-radius: 18px;
  background: linear-gradient(165deg, #f5f3ff 0%, #fff 42%);
  box-shadow: 0 8px 28px rgba(124, 58, 237, 0.08);
}

.plan-card:last-child {
  margin-bottom: 0;
}

.plan-card-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.plan-card-head h2 {
  margin: 0 0 4px;
  font-size: 18px;
  font-weight: 700;
}

.sku {
  font-size: 12px;
  color: var(--atm-text-muted);
}

.plan-metrics {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px 20px;
}

@media (min-width: 640px) {
  .plan-metrics {
    grid-template-columns: repeat(4, 1fr);
  }
}

.metric-label {
  display: block;
  font-size: 12px;
  color: var(--atm-text-muted);
  margin-bottom: 4px;
}

.plan-metrics strong {
  font-size: 14px;
  color: var(--atm-text);
}

.plan-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: center;
  margin-top: 16px;
}

.table-wrap {
  overflow-x: auto;
  margin: 0 -4px;
}

.table-wrap--spaced {
  margin-top: 16px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.data-table th {
  padding: 14px 14px;
  text-align: left;
  font-weight: 600;
  color: var(--atm-text-muted);
  border-bottom: 1px solid #e2e8f0;
  white-space: nowrap;
}

.data-table td {
  padding: 14px;
  border-bottom: 1px solid #f1f5f9;
  color: var(--atm-text);
}

.mono {
  font-family: ui-monospace, monospace;
  font-size: 12px;
}

.muted {
  color: var(--atm-text-muted);
  white-space: nowrap;
}

.tag {
  display: inline-block;
  padding: 3px 8px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 999px;
  background: #f1f5f9;
  color: #64748b;
}

.tag--active,
.tag--completed {
  background: #dcfce7;
  color: #15803d;
}

.tag--pending_payment {
  background: #fef3c7;
  color: #b45309;
}

.tag--failed,
.tag--cancelled {
  background: #fee2e2;
  color: #b91c1c;
}

.balance-card {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 32px 28px;
  margin-bottom: 12px;
  background: linear-gradient(125deg, #4c1d95 0%, #7c3aed 55%, #6366f1 100%);
  border-radius: 18px;
  color: #fff;
  box-shadow: 0 16px 40px rgba(124, 58, 237, 0.35);
}

.balance-card .stat-label,
.balance-card .panel-desc {
  color: rgba(255, 255, 255, 0.78);
}

.balance-amount {
  margin: 10px 0 12px;
  font-size: 40px;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: #fff;
}

.amount-plus {
  color: #15803d;
  font-weight: 600;
}

.amount-minus {
  color: var(--atm-text);
  font-weight: 600;
}

.link-btn {
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-primary);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.link-btn:hover {
  text-decoration: underline;
}

.btn-xs {
  padding: 8px 16px;
  font-size: 13px;
}

.settings-list {
  margin: 0 0 28px;
}

.settings-row {
  display: grid;
  grid-template-columns: 100px 1fr;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid #f1f5f9;
  font-size: 14px;
}

.settings-row dt {
  color: var(--atm-text-muted);
}

.settings-row dd {
  margin: 0;
  color: var(--atm-text);
}

.settings-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.btn-logout {
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  color: #b91c1c;
  cursor: pointer;
  background: #fff;
  border: 1px solid #fecaca;
  border-radius: 999px;
}

.empty {
  margin: 0;
  font-size: 14px;
  color: var(--atm-text-muted);
}
</style>
