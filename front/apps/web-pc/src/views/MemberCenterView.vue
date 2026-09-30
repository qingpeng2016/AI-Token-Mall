<script setup lang="ts">
import { computed, nextTick, onBeforeMount, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  formatCny,
  formatNowBeijing,
  parseMoney,
  type OrderType,
  type UserWalletFlowItem,
} from '@ai-token-mall/shared'
import CatalogPickerModal from '@/components/catalog/CatalogPickerModal.vue'
import PurchaseModal from '@/components/checkout/PurchaseModal.vue'
import MemberSidebar from '@/components/member/MemberSidebar.vue'
import type { CatalogProduct } from '@/mocks/home'
import {
  reloadCatalogProducts,
  useCatalogProducts,
} from '@/composables/useCatalogProducts'
import {
  beginRouteNavigationLoading,
  endRouteNavigationLoading,
} from '@/composables/useRouteNavigationLoading'
import {
  clearSessionUser,
  getSessionUser,
  isLoggedIn,
  setSessionUser,
  userAccountLabel,
} from '@/composables/useSessionUser'
import { orderApi, subscriptionApi, userApi } from '@/api'
import type { UserOrderItem, UserSubscriptionItem } from '@ai-token-mall/shared'
import {
  formatTokens,
  memberNav,
  mockApiTeamMembers,
  mockInvitedUsers,
  mockInvoices,
  mockInviteRebatePolicy,
  mockInviteRebateRecords,
  mockPromoDomainBase,
  mockSubscriptions,
  mockTeamSubKeys,
  mockWithdrawalRecords,
  orderStatusLabel,
  withdrawalChannelLabel,
  withdrawalStatusLabel,
  type MemberTab,
  type MockSubAccount,
  type MockTeamSubKey,
  type MockWithdrawalRecord,
} from '@/mocks/member'

const router = useRouter()
const route = useRoute()
const user = ref(getSessionUser())
const { categories: catalogCategories, products: catalogProducts } = useCatalogProducts()
const catalogOpen = ref(false)
/** 非空时目录弹窗只展示该 products_category_id（升档） */
const catalogProductsCategoryId = ref<number | null>(null)
const catalogUpgradeBaselineLimitTokens = ref<number | null>(null)
const purchaseOpen = ref(false)
const purchaseProduct = ref<CatalogProduct | null>(null)
const purchaseOrderType = ref<OrderType>('purchase')
const purchaseUserSubscriptionId = ref(0)
const purchaseSubscriptionPeriodEnd = ref('')
const purchaseSubscriptionExpiresAt = ref('')
const memberSubscriptions = ref<UserSubscriptionItem[]>([])
const recentOrders = ref<UserOrderItem[]>([])
const memberOrders = ref<UserOrderItem[]>([])
const ordersPage = ref(1)
const ordersPageSize = 2
const ordersTotal = ref(0)
const walletFlows = ref<UserWalletFlowItem[]>([])
const walletBalance = ref(0)
const plansLoaded = ref(false)
const ordersLoaded = ref(false)
const recentOrdersLoaded = ref(false)
const walletFlowsLoaded = ref(false)
/** 各侧栏 Tab 仅首次进入时拉取接口，关闭弹窗等不重复请求 */
const tabLoadedOnce = ref<Partial<Record<MemberTab | 'profile' | 'recentOrders', boolean>>>({})

const subscriptionStatusLabel: Record<string, string> = {
  active: '使用中',
  expired: '已过期',
  suspended: '已暂停',
  cancelled: '已取消',
}
const teamInviteOpen = ref(false)
const apiKeyPanelTab = ref<'mine' | 'team' | 'group'>('mine')
const teamMembers = ref<MockSubAccount[]>([...mockApiTeamMembers])
const addTeamMemberOpen = ref(false)
const addTeamMemberInvitedId = ref<number | null>(null)
const inviteRebatePanelTab = ref<'details' | 'members' | 'rebates' | 'withdrawals'>('details')
const commissionAvailable = ref(0)
const withdrawalRecords = ref<MockWithdrawalRecord[]>([...mockWithdrawalRecords])
const payoutQr = reactive({ alipay: '', wechat: '' })
const payoutQrSetupOpen = ref(false)
const payoutQrSetupChannel = ref<'alipay' | 'wechat'>('alipay')
const payoutQrSetupValue = ref('')
const withdrawOpen = ref(false)
const withdrawForm = reactive({
  channel: 'alipay' as 'alipay' | 'wechat',
  amountYuan: '',
})
const teamSubKeys = ref<MockTeamSubKey[]>([...mockTeamSubKeys])
const assignSubKeyOpen = ref(false)
const assignSubKeyForm = reactive({
  subscriptionId: null as number | null,
  memberId: null as number | null,
  limitTokens: 0,
})

const editSubKeyLimitOpen = ref(false)
const editSubKeyLimitTargetId = ref<number | null>(null)
const editSubKeyLimitValue = ref(0)

const assignSubKeySubscription = computed(() =>
  mockSubscriptions.find((s) => s.id === assignSubKeyForm.subscriptionId),
)

const editSubKeyLimitTarget = computed(() =>
  editSubKeyLimitTargetId.value == null
    ? null
    : teamSubKeys.value.find((k) => k.id === editSubKeyLimitTargetId.value) ?? null,
)

const editSubKeyLimitMainTotal = computed(() => {
  const k = editSubKeyLimitTarget.value
  if (!k) return 0
  return mockSubscriptions.find((s) => s.id === k.subscriptionId)?.limitTokens ?? k.limitTokens
})

const activeTeamMembers = computed(() =>
  teamMembers.value.filter((m) => m.status === 'active'),
)

const apiTeamMemberIds = computed(() => new Set(teamMembers.value.map((m) => m.id)))

const addableInvitedUsers = computed(() =>
  mockInvitedUsers.filter((u) => !apiTeamMemberIds.value.has(u.id)),
)

function isInvitedUserInApiTeam(userId: number) {
  return apiTeamMemberIds.value.has(userId)
}

const activeTeamSubKeys = computed(() =>
  teamSubKeys.value.filter((k) => k.status === 'active'),
)

type MyApiKeyRow =
  | {
      keyType: 'main'
      rowKey: string
      planName: string
      apiKeyMasked: string
      apiKeyCopyValue: string
      active: boolean
    }
  | {
      keyType: 'team'
      rowKey: string
      planName: string
      memberNickname: string
      apiKeyMasked: string
      apiKeyCopyValue: string
      active: boolean
    }

const myApiKeyRows = computed<MyApiKeyRow[]>(() => {
  const mains: MyApiKeyRow[] = mockSubscriptions.map((sub) => ({
    keyType: 'main',
    rowKey: `main-${sub.id}`,
    planName: sub.productName,
    apiKeyMasked: sub.apiKeyMasked,
    apiKeyCopyValue: sub.apiKeyCopyValue,
    active: sub.status === 'active',
  }))
  const teams: MyApiKeyRow[] = activeTeamSubKeys.value.map((k) => ({
    keyType: 'team',
    rowKey: `team-${k.id}`,
    planName: k.subscriptionName,
    memberNickname: k.memberNickname,
    apiKeyMasked: k.apiKeyMasked,
    apiKeyCopyValue: k.apiKeyCopyValue,
    active: k.status === 'active',
  }))
  return [...mains, ...teams]
})

const validTabs = new Set(memberNav.map((n) => n.id))

const activeTab = computed<MemberTab>(() => {
  const q = route.query.tab
  const t = typeof q === 'string' ? q : 'overview'
  return validTabs.has(t as MemberTab) ? (t as MemberTab) : 'overview'
})

watch(
  activeTab,
  (tab, prevTab) => {
    if (tab !== 'api-keys') apiKeyPanelTab.value = 'mine'
    if (tab !== 'sub-accounts') inviteRebatePanelTab.value = 'details'
    if (tab === prevTab) return
    if ((tab === 'plans' || tab === 'overview') && !tabLoadedOnce.value.plans) {
      tabLoadedOnce.value.plans = true
      void fetchPlansTabData()
    }
    if (tab === 'orders') {
      void fetchOrdersTabData()
    } else if (tab === 'overview' && !tabLoadedOnce.value.recentOrders) {
      tabLoadedOnce.value.recentOrders = true
      void fetchRecentOrders()
    }
    if ((tab === 'overview' || tab === 'settings') && !tabLoadedOnce.value.profile) {
      tabLoadedOnce.value.profile = true
      void fetchMemberProfile()
    }
    if (tab === 'account' && !tabLoadedOnce.value.account) {
      tabLoadedOnce.value.account = true
      void fetchWalletFlowsTabData()
    }
  },
  { immediate: true },
)

const pageTitle = computed(() => memberNav.find((n) => n.id === activeTab.value)?.label ?? '会员中心')

const displayName = computed(() =>
  user.value ? userAccountLabel(user.value) : '会员',
)

const overviewSubscriptions = computed(() =>
  memberSubscriptions.value.filter((s) => s.status === 'active'),
)

const teamInviteLink = computed(() => {
  const origin = typeof window !== 'undefined' ? window.location.origin : ''
  const uid = user.value?.id ?? 'guest'
  return `${origin}/register?team_invite=${uid}`
})

const exclusivePromoDomain = computed(() => {
  const uid = user.value?.id
  return uid != null ? `${mockPromoDomainBase}/${uid}` : `${mockPromoDomainBase}/guest`
})

onBeforeMount(() => {
  const profile = getSessionUser()
  if (!isLoggedIn() || !profile) {
    clearSessionUser()
    void router.replace({ path: '/login', query: { redirect: route.fullPath } })
    return
  }
  user.value = profile
})

function usagePercent(used: number, limit: number) {
  return limit > 0 ? Math.min(100, Math.round((used / limit) * 100)) : 0
}

function isSubscriptionActive(status: string) {
  return status === 'active'
}

function subscriptionTagClass(status: string) {
  if (status === 'active') return 'tag--active'
  if (status === 'expired') return 'tag--cancelled'
  return `tag--${status}`
}

async function ensureCatalogReady() {
  if (catalogProducts.value.length) return
  await reloadCatalogProducts({ soft: true })
}

async function withTopLoading<T>(task: () => Promise<T>): Promise<T> {
  beginRouteNavigationLoading()
  try {
    return await task()
  } finally {
    endRouteNavigationLoading()
  }
}

async function fetchPlansTabData() {
  void ensureCatalogReady().catch(() => {})
  await withTopLoading(async () => {
    try {
      memberSubscriptions.value = await subscriptionApi.list()
    } catch (e) {
      const msg = e instanceof Error ? e.message : ''
      if (/unauthorized|401/i.test(msg)) {
        ElMessage.error('登录已失效，请重新登录')
        clearSessionUser()
        void router.replace({ path: '/login', query: { redirect: route.fullPath } })
      } else {
        ElMessage.error(msg || '套餐列表加载失败')
      }
    } finally {
      plansLoaded.value = true
    }
  })
}

async function fetchRecentOrders() {
  await withTopLoading(async () => {
    try {
      const data = await orderApi.list({ page: 1, page_size: 3 })
      recentOrders.value = data.items
    } catch (e) {
      handleMemberAuthError(e, '最近订单加载失败')
    } finally {
      recentOrdersLoaded.value = true
    }
  })
}

async function fetchOrdersTabData() {
  await withTopLoading(async () => {
    try {
      const data = await orderApi.list({
        page: ordersPage.value,
        page_size: ordersPageSize,
      })
      memberOrders.value = data.items
      ordersTotal.value = data.total
      ordersPage.value = data.page
    } catch (e) {
      handleMemberAuthError(e, '订单列表加载失败')
    } finally {
      ordersLoaded.value = true
    }
  })
}

function onOrdersPageChange(page: number) {
  // v-model 已先更新 current-page，不可再与 ordersPage 比较后跳过请求
  ordersPage.value = page
  void fetchOrdersTabData()
}

function handleMemberAuthError(e: unknown, fallback: string) {
  const msg = e instanceof Error ? e.message : ''
  if (/unauthorized|401/i.test(msg)) {
    ElMessage.error('登录已失效，请重新登录')
    clearSessionUser()
    void router.replace({ path: '/login', query: { redirect: route.fullPath } })
    return
  }
  ElMessage.error(msg || fallback)
}

async function fetchMemberProfile() {
  await withTopLoading(async () => {
    try {
      const profile = await userApi.me()
      user.value = profile
      setSessionUser(profile)
      walletBalance.value = parseMoney(profile.wallet_balance)
      commissionAvailable.value = parseMoney(profile.commission_balance)
    } catch (e) {
      handleMemberAuthError(e, '用户信息加载失败')
    }
  })
}

async function fetchWalletFlowsTabData() {
  await withTopLoading(async () => {
    try {
      walletFlows.value = await userApi.walletFlows()
    } catch (e) {
      handleMemberAuthError(e, '资金流水加载失败')
    } finally {
      walletFlowsLoaded.value = true
    }
  })
}

function walletFlowTypeLabel(type: string) {
  if (type === 'recharge') return '充值'
  if (type === 'refund') return '退款'
  if (type === 'commission') return '佣金'
  if (type === 'withdraw') return '提现'
  return '消费'
}

function findCatalogProductById(productId: number) {
  return catalogProducts.value.find((item) => item.id === productId)
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

function goInviteRebateTab() {
  inviteRebatePanelTab.value = 'members'
  void router.push({ path: '/member', query: { tab: 'sub-accounts' } })
}

function goCommissionWithdrawFromOverview() {
  inviteRebatePanelTab.value = 'withdrawals'
  const open = () => {
    void nextTick(() => openWithdrawModal())
  }
  if (route.path === '/member' && route.query.tab === 'sub-accounts') {
    open()
    return
  }
  void router.push({ path: '/member', query: { tab: 'sub-accounts' } }).then(open)
}

function openCatalogPicker() {
  purchaseOrderType.value = 'purchase'
  purchaseUserSubscriptionId.value = 0
  purchaseSubscriptionPeriodEnd.value = ''
  purchaseSubscriptionExpiresAt.value = ''
  catalogProductsCategoryId.value = null
  catalogUpgradeBaselineLimitTokens.value = null
  catalogOpen.value = true
}

function resolveUpgradeCategoryId(sub: UserSubscriptionItem): number | null {
  const byProduct = catalogProducts.value.find((item) => item.id === sub.product_id)
  if (byProduct?.products_category_id != null && byProduct.products_category_id > 0) {
    return byProduct.products_category_id
  }
  for (const cat of catalogCategories.value) {
    if (cat.products.some((p) => p.id === sub.product_id)) {
      return cat.id
    }
  }
  const catName = sub.products_category_name?.trim()
  if (catName) {
    const cat = catalogCategories.value.find((c) =>
      c.products.some((p) => p.products_category_name === catName),
    )
    if (cat) return cat.id
  }
  return null
}

async function openUpgradeCatalog(sub: UserSubscriptionItem) {
  await ensureCatalogReady()
  const categoryId = resolveUpgradeCategoryId(sub)
  if (categoryId == null) {
    ElMessage.warning('未找到该套餐分类，请稍后重试或联系客服')
    return
  }
  purchaseOrderType.value = 'upgrade'
  purchaseUserSubscriptionId.value = sub.id
  purchaseSubscriptionPeriodEnd.value = sub.period_end
  purchaseSubscriptionExpiresAt.value = sub.expires_at
  catalogProductsCategoryId.value = categoryId
  catalogUpgradeBaselineLimitTokens.value = sub.limit_tokens
  catalogOpen.value = true
}

async function openQuotaAddon(sub: UserSubscriptionItem) {
  await ensureCatalogReady()
  const p = findCatalogProductById(sub.product_id)
  if (!p) {
    ElMessage.warning('未找到该套餐商品，请稍后重试或联系客服')
    return
  }
  purchaseOrderType.value = 'quota_addon'
  purchaseUserSubscriptionId.value = sub.id
  purchaseProduct.value = p
  purchaseOpen.value = true
}

async function openRenewSubscription(sub: UserSubscriptionItem) {
  await ensureCatalogReady()
  const p = findCatalogProductById(sub.product_id)
  if (!p) {
    ElMessage.warning('未找到该套餐商品，请稍后重试或联系客服')
    return
  }
  purchaseOrderType.value = 'renewal'
  purchaseUserSubscriptionId.value = sub.id
  purchaseProduct.value = p
  purchaseOpen.value = true
}

function onCatalogBuy(p: CatalogProduct) {
  if (purchaseOrderType.value !== 'upgrade') {
    purchaseOrderType.value = 'purchase'
    purchaseUserSubscriptionId.value = 0
  }
  purchaseProduct.value = p
  purchaseOpen.value = true
}

function onPurchasePaid() {
  catalogOpen.value = false
  if (activeTab.value === 'plans') {
    void fetchPlansTabData()
  }
}

function mockAction(msg: string) {
  ElMessage.info(msg)
}

async function copyToClipboard(text: string, successMessage: string) {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success(successMessage)
  } catch {
    ElMessage.warning('复制失败，请手动选择复制')
  }
}

function copyPlanApiKey(fullKey: string) {
  void copyToClipboard(fullKey, '已复制 API Key')
}

function openTeamInviteModal() {
  teamInviteOpen.value = true
}

function copyTeamInviteLink() {
  void copyToClipboard(teamInviteLink.value, '已复制邀请链接')
}

function copyExclusivePromoDomain() {
  void copyToClipboard(exclusivePromoDomain.value, '已复制推广域名')
}

function openPayoutQrModal(channel: 'alipay' | 'wechat') {
  payoutQrSetupChannel.value = channel
  payoutQrSetupValue.value = channel === 'alipay' ? payoutQr.alipay : payoutQr.wechat
  payoutQrSetupOpen.value = true
}

function confirmPayoutQrSetup() {
  const url = payoutQrSetupValue.value.trim()
  if (!url) {
    ElMessage.warning('请填写收款码图片地址或上传后粘贴链接（演示）')
    return
  }
  if (payoutQrSetupChannel.value === 'alipay') payoutQr.alipay = url
  else payoutQr.wechat = url
  payoutQrSetupOpen.value = false
  ElMessage.success('收款码已保存')
}

function openWithdrawModal() {
  if (!payoutQr.alipay && !payoutQr.wechat) {
    ElMessage.warning('请先设置支付宝或微信收款码')
    return
  }
  withdrawForm.channel = payoutQr.alipay ? 'alipay' : 'wechat'
  withdrawForm.amountYuan = ''
  withdrawOpen.value = true
}

function confirmWithdraw() {
  const yuan = Number(withdrawForm.amountYuan)
  if (!Number.isFinite(yuan) || yuan <= 0) {
    ElMessage.warning('请输入有效的提现金额')
    return
  }
  if (yuan > commissionAvailable.value) {
    ElMessage.warning('提现金额不能超过可提现佣金')
    return
  }
  const channel = withdrawForm.channel
  if (channel === 'alipay' && !payoutQr.alipay) {
    ElMessage.warning('请先设置支付宝收款码')
    return
  }
  if (channel === 'wechat' && !payoutQr.wechat) {
    ElMessage.warning('请先设置微信收款码')
    return
  }
  withdrawalRecords.value.unshift({
    id: Date.now(),
    amount: yuan,
    channel,
    status: 'pending',
    createdAt: formatNowBeijing('datetime'),
  })
  commissionAvailable.value -= yuan
  withdrawOpen.value = false
  ElMessage.success('提现申请已提交')
}

const payoutQrSetupTitle = computed(() =>
  payoutQrSetupChannel.value === 'alipay' ? '设置支付宝收款码' : '设置微信收款码',
)

function syncAssignSubKeyLimitFromPlan() {
  const sub = mockSubscriptions.find((s) => s.id === assignSubKeyForm.subscriptionId)
  assignSubKeyForm.limitTokens = sub?.limitTokens ?? 0
}

function openAssignSubKeyModal() {
  if (!mockSubscriptions.length) {
    ElMessage.warning('请先开通套餐')
    return
  }
  if (!activeTeamMembers.value.length) {
    ElMessage.warning('请先在「我的团队」添加成员')
    return
  }
  assignSubKeyForm.subscriptionId = mockSubscriptions[0]?.id ?? null
  assignSubKeyForm.memberId = activeTeamMembers.value[0]?.id ?? null
  syncAssignSubKeyLimitFromPlan()
  assignSubKeyOpen.value = true
}

function openEditSubKeyLimitModal(k: MockTeamSubKey) {
  editSubKeyLimitTargetId.value = k.id
  editSubKeyLimitValue.value = k.limitTokens
  editSubKeyLimitOpen.value = true
}

function confirmEditSubKeyLimit() {
  const k = editSubKeyLimitTarget.value
  if (!k) return
  const max = editSubKeyLimitMainTotal.value
  const next = Math.floor(editSubKeyLimitValue.value)
  if (!Number.isFinite(next) || next <= 0) {
    ElMessage.warning('请输入有效的用量上限')
    return
  }
  if (next > max) {
    ElMessage.warning(`不能超过主 Key 套餐总量（${formatTokens(max)}）`)
    return
  }
  if (next < k.usedTokens) {
    ElMessage.warning(`上限不能低于已用量（${formatTokens(k.usedTokens)}）`)
    return
  }
  k.limitTokens = next
  editSubKeyLimitOpen.value = false
  ElMessage.success('子 Key 用量上限已更新')
}

function confirmAssignSubKey() {
  const subId = assignSubKeyForm.subscriptionId
  const memberId = assignSubKeyForm.memberId
  if (subId == null || memberId == null) {
    ElMessage.warning('请选择套餐与团队成员')
    return
  }
  const subscription = mockSubscriptions.find((s) => s.id === subId)
  const member = teamMembers.value.find((m) => m.id === memberId)
  if (!subscription || !member) return

  const duplicate = teamSubKeys.value.some(
    (k) => k.subscriptionId === subId && k.memberId === memberId && k.status === 'active',
  )
  if (duplicate) {
    ElMessage.warning('该成员在此套餐下已有子 Key')
    return
  }

  const suffix = Math.random().toString(36).slice(2, 8)
  const full = `ap_sub_${subId}m${memberId}_${suffix}`
  const masked = `${full.slice(0, 12)}••••••${full.slice(-4)}`

  const mainLimit = subscription.limitTokens
  let limitTokens = Math.floor(assignSubKeyForm.limitTokens)
  if (!Number.isFinite(limitTokens) || limitTokens <= 0) limitTokens = mainLimit
  if (limitTokens > mainLimit) {
    ElMessage.warning(`用量上限不能超过主 Key 套餐总量（${formatTokens(mainLimit)}）`)
    return
  }

  teamSubKeys.value.push({
    id: Date.now(),
    subscriptionId: subId,
    subscriptionName: subscription.productName,
    memberId,
    memberNickname: member.nickname,
    memberEmail: member.email,
    apiKeyMasked: masked,
    apiKeyCopyValue: full,
    usedTokens: 0,
    limitTokens,
    status: 'active',
    createdAt: formatNowBeijing('date'),
  })

  assignSubKeyOpen.value = false
  apiKeyPanelTab.value = 'team'
  ElMessage.success('子 Key 已分配')
}

function openAddTeamMemberModal() {
  if (!addableInvitedUsers.value.length) {
    ElMessage.warning('暂无已邀请且未加入团队的成员，请先在「邀请返利」分享邀请链接')
    return
  }
  addTeamMemberInvitedId.value = addableInvitedUsers.value[0]?.id ?? null
  addTeamMemberOpen.value = true
}

function confirmAddTeamMember() {
  const invitedId = addTeamMemberInvitedId.value
  if (invitedId == null) {
    ElMessage.warning('请选择成员')
    return
  }
  const invited = mockInvitedUsers.find((u) => u.id === invitedId)
  if (!invited) return
  if (apiTeamMemberIds.value.has(invited.id)) {
    ElMessage.warning('该成员已在团队中')
    return
  }
  teamMembers.value.push({
    id: invited.id,
    nickname: invited.nickname,
    email: invited.email,
    usedTokens: 0,
    limitTokens: 0,
    status: 'active',
  })
  addTeamMemberOpen.value = false
  ElMessage.success('已加入团队')
}
</script>

<template>
  <div v-if="user" class="member-layout">
    <section class="member-hero">
      <div class="member-hero-bg" aria-hidden="true" />
      <div class="atm-container member-hero-inner">
        <div class="member-hero-copy">
          <span class="member-hero-badge">AI Plan · 会员中心</span>
          <h1 class="member-hero-title">
            <span class="member-hero-greeting">你好</span>
            <span class="member-hero-account">{{ displayName }}</span>
          </h1>
        </div>
        <div class="member-hero-actions">
          <button type="button" class="hero-btn hero-btn--light" @click="openCatalogPicker">
            选购套餐
          </button>
          <button type="button" class="hero-btn hero-btn--ghost" @click="goInviteRebateTab">
            邀请返利
          </button>
        </div>
      </div>
    </section>

    <div class="atm-container member-body">
      <div class="member-grid">
        <div class="member-sidebar-wrap">
          <MemberSidebar :active="activeTab" :user="user" @pick-plan="openCatalogPicker" />
        </div>

        <main class="member-main">
          <section class="panel">
            <header
              class="panel-head"
              :class="{
                'panel-head--segmented':
                  activeTab === 'api-keys' || activeTab === 'sub-accounts',
              }"
            >
              <div class="panel-head-row">
                <h2 class="panel-head-title">{{ pageTitle }}</h2>
                <button
                  v-if="activeTab === 'invoices'"
                  type="button"
                  class="link-btn panel-head-action"
                  @click="mockAction('发票抬头管理对接中')"
                >
                  管理发票抬头 →
                </button>
              </div>
              <div
                v-if="activeTab === 'sub-accounts'"
                class="api-key-tabs"
                role="tablist"
                aria-label="邀请返利"
              >
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': inviteRebatePanelTab === 'details' }"
                  role="tab"
                  :aria-selected="inviteRebatePanelTab === 'details'"
                  @click="inviteRebatePanelTab = 'details'"
                >
                  返佣详情
                </button>
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': inviteRebatePanelTab === 'withdrawals' }"
                  role="tab"
                  :aria-selected="inviteRebatePanelTab === 'withdrawals'"
                  @click="inviteRebatePanelTab = 'withdrawals'"
                >
                  我的佣金
                </button>
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': inviteRebatePanelTab === 'members' }"
                  role="tab"
                  :aria-selected="inviteRebatePanelTab === 'members'"
                  @click="inviteRebatePanelTab = 'members'"
                >
                  邀请成员
                </button>
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': inviteRebatePanelTab === 'rebates' }"
                  role="tab"
                  :aria-selected="inviteRebatePanelTab === 'rebates'"
                  @click="inviteRebatePanelTab = 'rebates'"
                >
                  返利记录
                </button>
              </div>
              <div
                v-if="activeTab === 'api-keys'"
                class="api-key-tabs"
                role="tablist"
                aria-label="API Key 类型"
              >
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': apiKeyPanelTab === 'mine' }"
                  role="tab"
                  :aria-selected="apiKeyPanelTab === 'mine'"
                  @click="apiKeyPanelTab = 'mine'"
                >
                  我的 Key
                </button>
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': apiKeyPanelTab === 'team' }"
                  role="tab"
                  :aria-selected="apiKeyPanelTab === 'team'"
                  @click="apiKeyPanelTab = 'team'"
                >
                  团队子 Key
                </button>
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': apiKeyPanelTab === 'group' }"
                  role="tab"
                  :aria-selected="apiKeyPanelTab === 'group'"
                  @click="apiKeyPanelTab = 'group'"
                >
                  我的团队
                </button>
              </div>
            </header>

            <!-- 概览 -->
            <div v-if="activeTab === 'overview'" class="panel-body">
            <div class="stat-row">
              <article class="stat-card stat-card--balance">
                <div class="stat-main">
                  <span class="stat-label">余额</span>
                  <strong class="stat-value">{{ formatCny(walletBalance) }}</strong>
                </div>
                <button type="button" class="stat-link" @click="mockRecharge">充值</button>
              </article>
              <article class="stat-card stat-card--commission">
                <div class="stat-main">
                  <span class="stat-label">佣金</span>
                  <strong class="stat-value">{{
                    formatCny(commissionAvailable)
                  }}</strong>
                </div>
                <button type="button" class="stat-link" @click="goCommissionWithdrawFromOverview">
                  提现
                </button>
              </article>
            </div>

            <h2 class="panel-subtitle">套餐用量</h2>
            <div v-if="overviewSubscriptions.length" class="plan-mini-list">
              <article
                v-for="sub in overviewSubscriptions"
                :key="sub.id"
                class="plan-mini"
              >
                <div class="plan-mini-head">
                  <strong>{{ sub.product_name }}</strong>
                  <span>至 {{ sub.expires_at }}</span>
                </div>
                <div class="progress-track">
                  <div
                    class="progress-fill"
                    :style="{ width: `${usagePercent(sub.used_tokens, sub.limit_tokens)}%` }"
                  />
                </div>
                <p class="progress-meta">
                  已用 {{ formatTokens(sub.used_tokens) }} / {{ formatTokens(sub.limit_tokens) }}
                </p>
              </article>
            </div>
            <p v-else-if="plansLoaded" class="empty">暂无使用中的套餐</p>

            <h2 class="panel-subtitle">最近订单</h2>
            <div v-if="recentOrders.length" class="table-wrap">
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
                  <tr v-for="o in recentOrders" :key="o.order_no">
                    <td>{{ o.order_no }}</td>
                    <td>{{ o.product_name }}</td>
                    <td>{{ formatCny(o.total_amount) }}</td>
                    <td>
                      <span class="tag" :class="`tag--${o.status}`">{{
                        orderStatusLabel[o.status]
                      }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else-if="recentOrdersLoaded" class="empty">暂无订单</p>
            <p v-if="recentOrders.length" class="overview-orders-more muted">
              <RouterLink :to="{ path: '/member', query: { tab: 'orders' } }">查看全部订单 →</RouterLink>
            </p>
            </div>

            <!-- 我的套餐 -->
            <div v-else-if="activeTab === 'plans'" class="panel-body">
              <article v-for="sub in memberSubscriptions" :key="sub.id" class="plan-card">
                <div class="plan-card-head">
                  <div>
                    <h2>{{ sub.product_name }}</h2>
                  </div>
                  <div class="plan-card-tags">
                    <span v-if="isSubscriptionActive(sub.status)" class="tag tag--active">
                      余额自动续费中
                    </span>
                    <span
                      v-else
                      class="tag"
                      :class="subscriptionTagClass(sub.status)"
                    >
                      {{ subscriptionStatusLabel[sub.status] ?? sub.status }}
                    </span>
                  </div>
                </div>
                <div class="plan-metrics">
                  <div>
                    <span class="metric-label">本周期额度</span>
                    <strong>{{ formatTokens(sub.limit_tokens) }} tokens</strong>
                  </div>
                  <div>
                    <span class="metric-label">已使用</span>
                    <strong>{{ formatTokens(sub.used_tokens) }}</strong>
                  </div>
                  <div>
                    <span class="metric-label">周期截止</span>
                    <strong>{{ sub.period_end }}</strong>
                  </div>
                  <div>
                    <span class="metric-label">服务到期</span>
                    <strong>{{ sub.expires_at }}</strong>
                  </div>
                </div>
                <div
                  class="progress-track progress-track--lg"
                  :class="{ 'progress-track--inactive': !isSubscriptionActive(sub.status) }"
                >
                  <div
                    v-if="isSubscriptionActive(sub.status)"
                    class="progress-fill"
                    :style="{ width: `${usagePercent(sub.used_tokens, sub.limit_tokens)}%` }"
                  />
                </div>
                <div v-if="isSubscriptionActive(sub.status)" class="plan-actions">
                  <button
                    type="button"
                    class="atm-btn-primary btn-xs"
                    @click="openRenewSubscription(sub)"
                  >
                    续费
                  </button>
                  <button type="button" class="atm-btn-primary btn-xs" @click="openUpgradeCatalog(sub)">
                    升档
                  </button>
                  <button
                    type="button"
                    class="atm-btn-primary btn-xs"
                    @click="openQuotaAddon(sub)"
                  >
                    加本期额度
                  </button>
                </div>
              </article>
              <p v-if="plansLoaded && !memberSubscriptions.length" class="empty">
                暂无套餐记录，去首页选购吧。
              </p>
            </div>

            <!-- API Key -->
            <div v-else-if="activeTab === 'api-keys'" class="panel-body panel-body--segmented">
            <div v-if="apiKeyPanelTab === 'mine'" role="tabpanel">
              <div v-if="myApiKeyRows.length" class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>类型</th>
                      <th>关联套餐</th>
                      <th>API Key</th>
                      <th>状态</th>
                      <th />
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="row in myApiKeyRows" :key="row.rowKey">
                      <td>
                        <span
                          class="tag"
                          :class="row.keyType === 'main' ? 'tag--key-main' : 'tag--key-team'"
                        >
                          {{ row.keyType === 'main' ? '主 Key' : '团队子 Key' }}
                        </span>
                      </td>
                      <td>
                        {{ row.planName }}
                        <span v-if="row.keyType === 'team'" class="cell-sub muted">
                          {{ row.memberNickname }}
                        </span>
                      </td>
                      <td class="mono">{{ row.apiKeyMasked }}</td>
                      <td>
                        <span class="tag" :class="row.active ? 'tag--active' : ''">{{
                          row.active ? '使用中' : '不可用'
                        }}</span>
                      </td>
                      <td>
                        <button
                          type="button"
                          class="link-btn"
                          @click="copyPlanApiKey(row.apiKeyCopyValue)"
                        >
                          复制
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else class="empty">暂无 Key，开通套餐或分配团队子 Key 后将在此展示。</p>
            </div>

            <div v-else-if="apiKeyPanelTab === 'team'" role="tabpanel">
              <div class="panel-tab-toolbar">
                <button type="button" class="atm-btn-primary btn-xs" @click="openAssignSubKeyModal">
                  分配子 Key
                </button>
              </div>
              <div v-if="activeTeamSubKeys.length" class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>成员</th>
                      <th>关联套餐</th>
                      <th>使用量 / 总量</th>
                      <th />
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="k in activeTeamSubKeys" :key="k.id">
                      <td>
                        <strong>{{ k.memberNickname }}</strong>
                        <span class="cell-sub muted">{{ k.memberEmail }}</span>
                      </td>
                      <td>{{ k.subscriptionName }}</td>
                      <td class="subkey-usage">
                        <span class="subkey-usage-text">
                          {{ formatTokens(k.usedTokens) }} / {{ formatTokens(k.limitTokens) }}
                        </span>
                        <div class="progress-track progress-track--sm">
                          <div
                            class="progress-fill"
                            :style="{
                              width: `${usagePercent(k.usedTokens, k.limitTokens)}%`,
                            }"
                          />
                        </div>
                      </td>
                      <td class="subkey-row-actions">
                        <button
                          type="button"
                          class="link-btn"
                          @click="copyPlanApiKey(k.apiKeyCopyValue)"
                        >
                          复制
                        </button>
                        <button
                          type="button"
                          class="link-btn"
                          @click="openEditSubKeyLimitModal(k)"
                        >
                          设置
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else class="empty">
                暂无团队子 Key。点击上方「分配子 Key」，为成员指定套餐下的独立 Key。
              </p>
            </div>

            <div v-else role="tabpanel">
              <div class="panel-tab-toolbar">
                <button type="button" class="atm-btn-primary btn-xs" @click="openAddTeamMemberModal">
                  添加成员
                </button>
              </div>
              <p class="rebate-lead muted">请从已通过邀请链接注册的用户中添加。</p>
              <div v-if="activeTeamMembers.length" class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>昵称</th>
                      <th>邮箱</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="m in teamMembers" :key="m.id">
                      <td>{{ m.nickname }}</td>
                      <td>{{ m.email }}</td>
                      <td>
                        <span class="tag tag--active">{{
                          m.status === 'active' ? '正常' : '停用'
                        }}</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else class="empty">
                团队暂无成员，点击上方「添加成员」从已邀请用户中选择。
              </p>
            </div>
            </div>

            <!-- 我的订单 -->
            <div v-else-if="activeTab === 'orders'" class="panel-body">
            <div v-if="memberOrders.length" class="table-wrap">
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
                  <tr v-for="o in memberOrders" :key="o.order_no">
                    <td class="mono">{{ o.order_no }}</td>
                    <td>{{ o.product_name }}</td>
                    <td>{{ o.quantity }}</td>
                    <td>{{ formatCny(o.total_amount) }}</td>
                    <td>{{ o.enterprise_invoice ? '企业' : '—' }}</td>
                    <td>
                      <span class="tag" :class="`tag--${o.status}`">{{
                        orderStatusLabel[o.status]
                      }}</span>
                    </td>
                    <td class="muted">{{ o.created_at }}</td>
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
            <p v-else-if="ordersLoaded" class="empty">暂无订单</p>
            <div v-if="ordersTotal > ordersPageSize" class="orders-pagination">
              <el-pagination
                v-model:current-page="ordersPage"
                :page-size="ordersPageSize"
                :total="ordersTotal"
                layout="total, prev, pager, next"
                background
                @current-change="onOrdersPageChange"
              />
            </div>
            </div>

            <!-- 资金流水 -->
            <div v-else-if="activeTab === 'account'" class="panel-body">
            <div v-if="walletFlows.length" class="table-wrap">
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
                  <tr v-for="tx in walletFlows" :key="tx.id">
                    <td>{{ walletFlowTypeLabel(tx.type) }}</td>
                    <td
                      :class="
                        parseMoney(tx.amount) > 0 ? 'amount-plus' : 'amount-minus'
                      "
                    >
                      {{ parseMoney(tx.amount) > 0 ? '+' : ''
                      }}{{ formatCny(Math.abs(parseMoney(tx.amount))) }}
                    </td>
                    <td>{{ tx.remark }}</td>
                    <td class="muted">{{ tx.created_at }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else-if="walletFlowsLoaded" class="empty">暂无流水</p>
            </div>

            <!-- 发票 -->
            <div v-else-if="activeTab === 'invoices'" class="panel-body">
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
                    <td>{{ formatCny(inv.amount) }}</td>
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
            </div>

            <!-- 邀请返利 -->
            <div v-else-if="activeTab === 'sub-accounts'" class="panel-body panel-body--segmented">
            <div v-if="inviteRebatePanelTab === 'details'" role="tabpanel">
              <article class="rebate-details-summary">
                <div class="rebate-details-section">
                  <span class="metric-label">专属推广域名</span>
                  <div class="promo-domain-row">
                    <code class="promo-domain-value">{{ exclusivePromoDomain }}</code>
                    <button
                      type="button"
                      class="atm-btn-primary btn-xs"
                      @click="copyExclusivePromoDomain"
                    >
                      复制
                    </button>
                  </div>
                </div>
                <div class="rebate-details-section">
                  <span class="metric-label">您当前的返佣比例</span>
                  <p class="rebate-rate-value">
                    {{ mockInviteRebatePolicy.currentRatePercent }}<span class="rebate-rate-unit">%</span>
                  </p>
                </div>
              </article>

              <h3 class="panel-subtitle">等级与比例</h3>
              <div class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>等级</th>
                      <th>有效邀请（≥）</th>
                      <th>返佣比例</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="tier in mockInviteRebatePolicy.tiers"
                      :key="tier.levelLabel"
                      :class="{
                        'rebate-tier-row--current':
                          tier.levelLabel === mockInviteRebatePolicy.currentLevelLabel,
                      }"
                    >
                      <td>
                        <strong>{{ tier.levelLabel }}</strong>
                        <span
                          v-if="tier.levelLabel === mockInviteRebatePolicy.currentLevelLabel"
                          class="tag tag--key-main rebate-tier-badge"
                        >
                          当前
                        </span>
                      </td>
                      <td>{{ tier.minInvites }} 人</td>
                      <td>{{ tier.ratePercent }}%</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h3 class="panel-subtitle">说明</h3>
              <ul class="rebate-notes">
                <li v-for="(note, i) in mockInviteRebatePolicy.notes" :key="i">{{ note }}</li>
              </ul>
            </div>

            <div v-else-if="inviteRebatePanelTab === 'members'" role="tabpanel">
              <div class="panel-tab-toolbar">
                <button type="button" class="atm-btn-primary btn-xs" @click="openTeamInviteModal">
                  邀请成员
                </button>
              </div>
              <div class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>昵称</th>
                      <th>邮箱</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="u in mockInvitedUsers" :key="u.id">
                      <td>{{ u.nickname }}</td>
                      <td>{{ u.email }}</td>
                      <td>
                        <span
                          class="tag"
                          :class="isInvitedUserInApiTeam(u.id) ? 'tag--active' : 'tag--key-team'"
                        >
                          {{ isInvitedUserInApiTeam(u.id) ? '已加入团队' : '已注册待添加' }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-if="!mockInvitedUsers.length" class="empty">
                暂无邀请成员，点击上方「邀请成员」分享链接。
              </p>
            </div>

            <div v-else-if="inviteRebatePanelTab === 'rebates'" role="tabpanel">
              <p class="rebate-lead muted">
                受邀用户完成支付后，返利将自动加到「佣金」。
              </p>
              <div v-if="mockInviteRebateRecords.length" class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>邀请用户</th>
                      <th>订单号</th>
                      <th>商品</th>
                      <th>订单金额</th>
                      <th>返利金额</th>
                      <th>时间</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="r in mockInviteRebateRecords" :key="r.id">
                      <td>
                        <strong>{{ r.inviteeNickname }}</strong>
                        <span class="cell-sub muted">{{ r.inviteeEmail }}</span>
                      </td>
                      <td class="mono">{{ r.orderNo }}</td>
                      <td>{{ r.productName }}</td>
                      <td>{{ formatCny(r.orderAmount) }}</td>
                      <td class="amount-plus">+{{ formatCny(r.rebateAmount) }}</td>
                      <td class="muted">{{ r.createdAt }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else class="empty">暂无返利记录。</p>
            </div>

            <div v-else-if="inviteRebatePanelTab === 'withdrawals'" role="tabpanel">
              <section class="withdraw-section-box">
                <h3 class="withdraw-section-title">佣金提现</h3>
                <div class="withdraw-toolbar">
                  <span class="withdraw-balance">
                    可提现佣金
                    <strong>{{ formatCny(commissionAvailable) }}</strong>
                  </span>
                  <button type="button" class="atm-btn-primary btn-xs" @click="openWithdrawModal">
                    提现
                  </button>
                </div>

                <h3 class="withdraw-section-title withdraw-section-title--sub">收款方式</h3>
                <div class="withdraw-payout-list">
                  <article class="withdraw-payout-row">
                    <strong class="withdraw-payout-name">支付宝</strong>
                    <p class="withdraw-payout-status muted">
                      {{ payoutQr.alipay ? '收款码已配置' : '未设置收款码' }}
                    </p>
                    <button
                      type="button"
                      class="atm-btn-primary btn-xs withdraw-payout-btn"
                      @click="openPayoutQrModal('alipay')"
                    >
                      {{ payoutQr.alipay ? '更换收款码' : '设置收款码' }}
                    </button>
                  </article>
                  <article class="withdraw-payout-row">
                    <strong class="withdraw-payout-name">微信</strong>
                    <p class="withdraw-payout-status muted">
                      {{ payoutQr.wechat ? '收款码已配置' : '未设置收款码' }}
                    </p>
                    <button
                      type="button"
                      class="atm-btn-primary btn-xs withdraw-payout-btn"
                      @click="openPayoutQrModal('wechat')"
                    >
                      {{ payoutQr.wechat ? '更换收款码' : '设置收款码' }}
                    </button>
                  </article>
                </div>
              </section>

              <h3 class="panel-subtitle">提现记录</h3>
              <div v-if="withdrawalRecords.length" class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>金额</th>
                      <th>到账方式</th>
                      <th>状态</th>
                      <th>时间</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="w in withdrawalRecords" :key="w.id">
                      <td>{{ formatCny(w.amount) }}</td>
                      <td>{{ withdrawalChannelLabel[w.channel] }}</td>
                      <td>
                        <span
                          class="tag"
                          :class="{
                            'tag--active': w.status === 'completed',
                            'tag--pending_payment': w.status === 'pending',
                            'tag--failed': w.status === 'failed',
                          }"
                        >
                          {{ withdrawalStatusLabel[w.status] }}
                        </span>
                      </td>
                      <td class="muted">{{ w.createdAt }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else class="empty">暂无提现记录。</p>
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
              <button type="button" class="atm-btn-primary btn-xs" @click="logout">退出登录</button>
            </div>
            </div>
          </section>
        </main>
      </div>
    </div>

    <CatalogPickerModal
      v-model:open="catalogOpen"
      :products-category-id="catalogProductsCategoryId"
      :upgrade-baseline-limit-tokens="catalogUpgradeBaselineLimitTokens"
      @buy="onCatalogBuy"
    />
    <PurchaseModal
      v-model:open="purchaseOpen"
      :product="purchaseProduct"
      :user="user"
      :order-type="purchaseOrderType"
      :user-subscription-id="purchaseUserSubscriptionId"
      :subscription-period-end="purchaseSubscriptionPeriodEnd"
      :subscription-expires-at="purchaseSubscriptionExpiresAt"
      @paid="onPurchasePaid"
    />

    <Teleport to="body">
      <div
        v-if="teamInviteOpen"
        class="team-invite-backdrop"
        @click.self="teamInviteOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="team-invite-title">
          <header class="team-invite-head">
            <h3 id="team-invite-title">邀请成员</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="teamInviteOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">将下方链接发给同事，对方打开并完成注册即可加入你的团队。</p>
          <label class="team-invite-field">
            <span class="metric-label">邀请链接</span>
            <div class="team-invite-row">
              <input class="team-invite-input" type="text" readonly :value="teamInviteLink" />
              <button type="button" class="atm-btn-primary btn-xs" @click="copyTeamInviteLink">
                复制链接
              </button>
            </div>
          </label>
        </div>
      </div>

      <div
        v-if="addTeamMemberOpen"
        class="team-invite-backdrop"
        @click.self="addTeamMemberOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="add-team-member-title">
          <header class="team-invite-head">
            <h3 id="add-team-member-title">添加成员</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="addTeamMemberOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            选择已通过您的邀请链接注册、尚未加入 API 团队的用户。
          </p>
          <label class="team-invite-field">
            <span class="metric-label">已邀请用户</span>
            <select v-model="addTeamMemberInvitedId" class="member-select">
              <option v-for="u in addableInvitedUsers" :key="u.id" :value="u.id">
                {{ u.nickname }}（{{ u.email }}）
              </option>
            </select>
          </label>
          <div class="assign-subkey-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="addTeamMemberOpen = false">
              取消
            </button>
            <button type="button" class="atm-btn-primary btn-xs" @click="confirmAddTeamMember">
              确认添加
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="payoutQrSetupOpen"
        class="team-invite-backdrop"
        @click.self="payoutQrSetupOpen = false"
      >
        <div class="team-invite-panel" role="dialog" :aria-labelledby="'payout-qr-title'">
          <header class="team-invite-head">
            <h3 :id="'payout-qr-title'">{{ payoutQrSetupTitle }}</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="payoutQrSetupOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            上传收款码后，将图片链接粘贴到下方（演示环境；正式版支持本地上传）。
          </p>
          <label class="team-invite-field">
            <span class="metric-label">收款码图片 URL</span>
            <input v-model="payoutQrSetupValue" type="url" class="team-invite-input" placeholder="https://..." />
          </label>
          <div class="assign-subkey-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="payoutQrSetupOpen = false">
              取消
            </button>
            <button type="button" class="atm-btn-primary btn-xs" @click="confirmPayoutQrSetup">
              保存
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="withdrawOpen"
        class="team-invite-backdrop"
        @click.self="withdrawOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="withdraw-title">
          <header class="team-invite-head">
            <h3 id="withdraw-title">提现</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="withdrawOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            可提现 {{ formatCny(commissionAvailable) }}，提现将打款至所选渠道的收款码。
          </p>
          <label class="team-invite-field">
            <span class="metric-label">到账方式</span>
            <select v-model="withdrawForm.channel" class="member-select">
              <option v-if="payoutQr.alipay" value="alipay">支付宝</option>
              <option v-if="payoutQr.wechat" value="wechat">微信</option>
            </select>
          </label>
          <label class="team-invite-field">
            <span class="metric-label">提现金额（元）</span>
            <input
              v-model="withdrawForm.amountYuan"
              type="number"
              min="0.01"
              step="0.01"
              class="team-invite-input"
              placeholder="例如 20.00"
            />
          </label>
          <div class="assign-subkey-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="withdrawOpen = false">
              取消
            </button>
            <button type="button" class="atm-btn-primary btn-xs" @click="confirmWithdraw">
              确认提现
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="assignSubKeyOpen"
        class="team-invite-backdrop"
        @click.self="assignSubKeyOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="assign-subkey-title">
          <header class="team-invite-head">
            <h3 id="assign-subkey-title">分配子 Key</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="assignSubKeyOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            从您的套餐额度中为团队成员生成独立子 Key，调用计入该套餐，用量单独归属成员。
          </p>
          <label class="team-invite-field">
            <span class="metric-label">关联套餐</span>
            <select
              v-model="assignSubKeyForm.subscriptionId"
              class="member-select"
              @change="syncAssignSubKeyLimitFromPlan"
            >
              <option v-for="s in mockSubscriptions" :key="s.id" :value="s.id">
                {{ s.productName }}
              </option>
            </select>
          </label>
          <label class="team-invite-field">
            <span class="metric-label">用量上限（tokens）</span>
            <input
              v-model.number="assignSubKeyForm.limitTokens"
              type="number"
              class="team-invite-input"
              min="1"
              :max="assignSubKeySubscription?.limitTokens"
            />
            <p v-if="assignSubKeySubscription" class="field-hint">
              默认与主 Key 一致：{{ formatTokens(assignSubKeySubscription.limitTokens) }}（本周期套餐总量）
            </p>
          </label>
          <label class="team-invite-field">
            <span class="metric-label">团队成员</span>
            <select v-model="assignSubKeyForm.memberId" class="member-select">
              <option v-for="m in activeTeamMembers" :key="m.id" :value="m.id">
                {{ m.nickname }}（{{ m.email }}）
              </option>
            </select>
          </label>
          <div class="assign-subkey-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="assignSubKeyOpen = false">
              取消
            </button>
            <button type="button" class="atm-btn-primary btn-xs" @click="confirmAssignSubKey">
              确认分配
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="editSubKeyLimitOpen && editSubKeyLimitTarget"
        class="team-invite-backdrop"
        @click.self="editSubKeyLimitOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="edit-subkey-limit-title">
          <header class="team-invite-head">
            <h3 id="edit-subkey-limit-title">设置子 Key 用量上限</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="editSubKeyLimitOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            {{ editSubKeyLimitTarget.memberNickname }} · {{ editSubKeyLimitTarget.subscriptionName }}
          </p>
          <label class="team-invite-field">
            <span class="metric-label">用量上限（tokens）</span>
            <input
              v-model.number="editSubKeyLimitValue"
              type="number"
              class="team-invite-input"
              min="1"
              :max="editSubKeyLimitMainTotal"
            />
            <p class="field-hint">
              主 Key 套餐总量 {{ formatTokens(editSubKeyLimitMainTotal) }}；已用
              {{ formatTokens(editSubKeyLimitTarget.usedTokens) }}
            </p>
          </label>
          <div class="assign-subkey-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="editSubKeyLimitOpen = false">
              取消
            </button>
            <button type="button" class="atm-btn-primary btn-xs" @click="confirmEditSubKeyLimit">
              保存
            </button>
          </div>
        </div>
      </div>
    </Teleport>
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
  font-weight: 700;
  line-height: 1.4;
}

.member-hero-greeting {
  display: block;
  font-size: clamp(18px, 2.2vw, 22px);
  letter-spacing: -0.02em;
}

.member-hero-account {
  display: block;
  margin-top: 4px;
  max-width: min(100%, 520px);
  font-size: clamp(14px, 1.6vw, 16px);
  font-weight: 600;
  letter-spacing: 0;
  opacity: 0.92;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 16px;
  margin: 0 0 24px;
}

.panel-head--segmented {
  margin-bottom: 0;
  gap: 12px;
}

.panel-body--segmented {
  margin-top: 12px;
}

.panel-head-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  width: 100%;
}

.panel-head-title {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.panel-head-action {
  flex-shrink: 0;
  font-size: 14px;
  font-weight: 600;
}

.panel-head-action--hidden {
  visibility: hidden;
  pointer-events: none;
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

.panel-subtitle--tight {
  margin-top: 0;
}

.withdraw-section-box {
  margin-bottom: 20px;
  padding: 20px 22px;
  background: linear-gradient(135deg, #f5f3ff 0%, #faf5ff 55%, #fff 100%);
  border: 1px solid rgba(124, 58, 237, 0.12);
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.06);
}

.withdraw-section-title {
  margin: 0 0 14px;
  font-size: 15px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.withdraw-section-title--sub {
  margin-top: 22px;
  padding-top: 20px;
  border-top: 1px solid rgba(124, 58, 237, 0.1);
}

.withdraw-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  width: 100%;
}

.withdraw-balance {
  font-size: 14px;
  color: var(--atm-text-muted);
}

.withdraw-balance strong {
  margin-left: 6px;
  font-size: 18px;
  font-weight: 800;
  color: var(--atm-text);
}

.withdraw-payout-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.withdraw-payout-row {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px 16px;
}

.withdraw-payout-name {
  flex-shrink: 0;
  min-width: 56px;
  font-size: 15px;
  font-weight: 700;
  color: var(--atm-text);
}

.withdraw-payout-status {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 13px;
  line-height: 1.5;
}

.withdraw-payout-btn {
  flex-shrink: 0;
  margin-left: auto;
}

.withdraw-qr-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.withdraw-qr-preview img {
  width: 120px;
  height: 120px;
  object-fit: contain;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}


.stat-row {
  display: grid;
  gap: 16px;
  padding: 4px 0;
}

@media (min-width: 640px) {
  .stat-row {
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
  }
}

.stat-card {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 0;
  padding: 16px 18px;
  background: #fff;
  border: 1px solid rgba(124, 58, 237, 0.1);
  border-radius: 16px;
  box-shadow: 0 6px 24px rgba(124, 58, 237, 0.06);
  transition:
    background 0.15s ease,
    box-shadow 0.15s ease;
}

.stat-card--balance,
.stat-card--commission {
  background: linear-gradient(160deg, #faf5ff 0%, #fff 70%);
}

.panel-tab-toolbar {
  display: flex;
  justify-content: flex-start;
  margin-bottom: 16px;
}

.rebate-lead {
  margin: 0 0 16px;
  font-size: 13px;
  line-height: 1.5;
}

.rebate-rate-hero {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px 24px;
  margin-bottom: 8px;
  padding: 20px 22px;
  background: linear-gradient(135deg, #f5f3ff 0%, #faf5ff 55%, #fff 100%);
  border: 1px solid rgba(124, 58, 237, 0.12);
  border-radius: 16px;
}

.rebate-details-summary {
  margin-bottom: 8px;
  padding: 20px 22px;
  background: linear-gradient(135deg, #f5f3ff 0%, #faf5ff 55%, #fff 100%);
  border: 1px solid rgba(124, 58, 237, 0.12);
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.06);
}

.rebate-details-section + .rebate-details-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid rgba(124, 58, 237, 0.1);
}

.promo-domain-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px 16px;
  margin-top: 8px;
}

.promo-domain-value {
  flex: 1;
  min-width: 0;
  margin: 0;
  padding: 0;
  font-size: 15px;
  font-weight: 600;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: var(--atm-primary-dark);
  word-break: break-all;
  background: transparent;
  border: none;
}

.rebate-rate-value {
  margin: 6px 0 0;
  font-size: 40px;
  font-weight: 800;
  line-height: 1;
  letter-spacing: -0.03em;
  color: var(--atm-primary-dark);
}

.rebate-rate-unit {
  margin-left: 2px;
  font-size: 22px;
  font-weight: 700;
}

.rebate-rate-meta {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--atm-text-muted);
}

.rebate-rate-next {
  max-width: 280px;
  padding: 12px 14px;
  background: #fff;
  border: 1px solid rgba(124, 58, 237, 0.1);
  border-radius: 12px;
}

.rebate-rate-next strong {
  display: block;
  margin-top: 4px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--atm-text);
}

.rebate-tier-row--current {
  background: rgba(237, 233, 254, 0.35);
}

.rebate-tier-badge {
  margin-left: 8px;
  vertical-align: middle;
}

.rebate-notes {
  margin: 0;
  padding-left: 1.25rem;
  font-size: 14px;
  line-height: 1.65;
  color: var(--atm-text-muted);
}

.rebate-notes li + li {
  margin-top: 8px;
}

.stat-card:hover {
  background: #fdfcff;
  box-shadow: 0 10px 28px rgba(124, 58, 237, 0.1);
}

.stat-card--balance:hover,
.stat-card--commission:hover {
  background: linear-gradient(160deg, #f3ebff 0%, #fff 70%);
}

.stat-main {
  flex: 1;
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

.stat-card--balance .stat-value,
.stat-card--commission .stat-value {
  color: var(--atm-primary-dark);
}

.stat-link {
  flex-shrink: 0;
  padding: 7px 15px;
  font-size: 14px;
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

.progress-track--inactive {
  background: #cbd5e1;
}

.progress-track--sm {
  height: 6px;
  margin-top: 6px;
}

.subkey-usage {
  min-width: 140px;
}

.subkey-usage-text {
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-text);
  white-space: nowrap;
}

.subkey-row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  white-space: nowrap;
}

.field-hint {
  margin: 8px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--atm-text-muted);
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
  color: var(--atm-text);
}

.plan-card-tags {
  display: flex;
  flex-shrink: 0;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  justify-content: flex-end;
}

.plan-card-head .tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
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

.orders-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.overview-orders-more {
  margin-top: 12px;
  font-size: 13px;
}

.overview-orders-more a {
  color: var(--atm-primary, #7c3aed);
  text-decoration: none;
}

.overview-orders-more a:hover {
  text-decoration: underline;
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

.tag--key-main {
  background: #ede9fe;
  color: #5b21b6;
}

.tag--key-team {
  background: #e0f2fe;
  color: #0369a1;
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

.settings-list .link-btn {
  text-decoration: underline;
  text-underline-offset: 3px;
}

.settings-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.empty {
  margin: 0;
  font-size: 14px;
  color: var(--atm-text-muted);
}

.api-key-tabs {
  display: inline-flex;
  gap: 4px;
  width: fit-content;
  padding: 4px;
  background: #f1f5f9;
  border-radius: 12px;
}

.api-key-tab {
  padding: 8px 18px;
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  color: var(--atm-text-muted);
  background: transparent;
  border: none;
  border-radius: 9px;
  cursor: pointer;
  transition:
    background 0.15s ease,
    color 0.15s ease,
    box-shadow 0.15s ease;
}

.api-key-tab--active {
  color: var(--atm-primary-dark);
  background: #fff;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.12);
}

.cell-sub {
  display: block;
  margin-top: 2px;
  font-size: 12px;
  font-weight: 400;
}

.member-select {
  display: block;
  width: 100%;
  margin-top: 8px;
  padding: 10px 12px;
  font-size: 14px;
  font-family: inherit;
  color: var(--atm-text);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.assign-subkey-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}

.team-invite-field + .team-invite-field {
  margin-top: 14px;
}

.team-invite-backdrop {
  position: fixed;
  inset: 0;
  z-index: 2050;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px 16px;
  background: rgba(15, 10, 40, 0.52);
  backdrop-filter: blur(4px);
}

.team-invite-panel {
  width: 100%;
  max-width: 520px;
  padding: 22px 24px 24px;
  background: #fff;
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(30, 27, 75, 0.22);
}

.team-invite-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.team-invite-head h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 800;
  color: var(--atm-text);
}

.team-invite-close {
  width: 36px;
  height: 36px;
  font-size: 22px;
  line-height: 1;
  color: var(--atm-text-muted);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
}

.team-invite-lead {
  margin: 0 0 18px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--atm-text-muted);
}

.team-invite-field {
  display: block;
}

.team-invite-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 8px;
}

.team-invite-input {
  flex: 1;
  min-width: 200px;
  padding: 10px 12px;
  font-size: 13px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: var(--atm-text);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}
</style>
