/** 会员中心 UI mock（待接 API） */

export type MemberTab =
  | 'overview'
  | 'plans'
  | 'orders'
  | 'account'
  | 'api-keys'
  | 'invoices'
  | 'sub-accounts'
  | 'settings'

export const memberNav: { id: MemberTab; label: string; desc?: string }[] = [
  { id: 'overview', label: '概览' },
  { id: 'plans', label: '我的套餐', desc: '额度与有效期' },
  { id: 'orders', label: '我的订单' },
  { id: 'account', label: '账户余额', desc: '充值与流水' },
  { id: 'api-keys', label: 'API 密钥' },
  { id: 'invoices', label: '发票管理' },
  { id: 'sub-accounts', label: '子账号' },
  { id: 'settings', label: '账户设置' },
]

export type MockSubscription = {
  id: number
  productName: string
  skuLabel: string
  status: 'active' | 'expired' | 'suspended'
  limitTokens: number
  usedTokens: number
  expiresAt: string
  periodEnd: string
}

export type MockOrder = {
  orderNo: string
  productName: string
  quantity: number
  totalCents: number
  status: 'pending_payment' | 'completed' | 'failed' | 'cancelled'
  createdAt: string
  enterpriseInvoice: boolean
}

export type MockWalletTx = {
  id: string
  type: 'recharge' | 'pay' | 'refund'
  amountCents: number
  remark: string
  createdAt: string
}

export type MockApiKey = {
  id: number
  name: string
  prefix: string
  subscriptionName: string
  createdAt: string
  lastUsedAt: string | null
}

export type MockInvoice = {
  id: number
  orderNo: string
  title: string
  amountCents: number
  status: 'pending' | 'issued' | 'failed'
  createdAt: string
}

export type MockSubAccount = {
  id: number
  nickname: string
  email: string
  usedTokens: number
  limitTokens: number
  status: 'active' | 'disabled'
}

export const mockMemberOverview = {
  balanceCents: 12800,
  activePlans: 2,
  pendingOrders: 1,
}

export const mockSubscriptions: MockSubscription[] = [
  {
    id: 1,
    productName: 'ChatGPT Plus 月卡',
    skuLabel: 'OAI-PLUS-M',
    status: 'active',
    limitTokens: 8_000_000,
    usedTokens: 3_240_000,
    expiresAt: '2026-04-26',
    periodEnd: '2026-04-01',
  },
  {
    id: 2,
    productName: 'Claude Pro 月卡',
    skuLabel: 'ANT-PRO-M',
    status: 'active',
    limitTokens: 8_000_000,
    usedTokens: 1_100_000,
    expiresAt: '2026-04-10',
    periodEnd: '2026-04-01',
  },
]

export const mockOrders: MockOrder[] = [
  {
    orderNo: 'AP202603270001',
    productName: 'ChatGPT Plus 月卡',
    quantity: 1,
    totalCents: 17800,
    status: 'completed',
    createdAt: '2026-03-27 14:20',
    enterpriseInvoice: false,
  },
  {
    orderNo: 'AP202603150032',
    productName: 'Claude Pro 月卡',
    quantity: 1,
    totalCents: 21900,
    status: 'completed',
    createdAt: '2026-03-15 09:08',
    enterpriseInvoice: true,
  },
  {
    orderNo: 'AP202603280008',
    productName: 'GPT Go 月卡',
    quantity: 1,
    totalCents: 8900,
    status: 'pending_payment',
    createdAt: '2026-03-28 10:00',
    enterpriseInvoice: false,
  },
]

export const mockWalletTx: MockWalletTx[] = [
  {
    id: 'tx1',
    type: 'recharge',
    amountCents: 50000,
    remark: '支付宝充值',
    createdAt: '2026-03-10 11:30',
  },
  {
    id: 'tx2',
    type: 'pay',
    amountCents: -17800,
    remark: '订单 AP202603270001',
    createdAt: '2026-03-27 14:21',
  },
  {
    id: 'tx3',
    type: 'pay',
    amountCents: -21900,
    remark: '订单 AP202603150032',
    createdAt: '2026-03-15 09:09',
  },
]

export const mockApiKeys: MockApiKey[] = [
  {
    id: 1,
    name: '默认密钥',
    prefix: 'ap_live_8f2a…',
    subscriptionName: 'ChatGPT Plus 月卡',
    createdAt: '2026-03-27',
    lastUsedAt: '2026-03-28 09:12',
  },
]

export const mockInvoices: MockInvoice[] = [
  {
    id: 1,
    orderNo: 'AP202603150032',
    title: '某某科技有限公司',
    amountCents: 23214,
    status: 'issued',
    createdAt: '2026-03-16',
  },
]

export const mockSubAccounts: MockSubAccount[] = [
  {
    id: 101,
    nickname: '同事 A',
    email: 'teama@example.com',
    usedTokens: 420_000,
    limitTokens: 2_000_000,
    status: 'active',
  },
]

export const orderStatusLabel: Record<MockOrder['status'], string> = {
  pending_payment: '待支付',
  completed: '已完成',
  failed: '失败',
  cancelled: '已取消',
}

export function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(0)}K`
  return String(n)
}
