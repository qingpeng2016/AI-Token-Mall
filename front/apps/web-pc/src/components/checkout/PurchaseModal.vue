<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  countUpgradeBillingCycles,
  formatCny,
  parseMoney,
  formatSubscriptionDate,
  parseSubscriptionDate,
  formatTokenCount,
} from '@ai-token-mall/shared'
import type { OrderType, UserProfile } from '@ai-token-mall/shared'
import { orderApi } from '@/api'
import type { CatalogProduct } from '@/mocks/home'

const props = withDefaults(
  defineProps<{
    open: boolean
    product: CatalogProduct | null
    user: UserProfile | null
    /** 新购默认 purchase；会员中心续费/升档 */
    orderType?: OrderType
    /** 续费/升档关联 user_subscriptions.id，新购为 0 */
    userSubscriptionId?: number
    /** 升档计价：当前订阅 period_end / expires_at */
    subscriptionPeriodEnd?: string
    subscriptionExpiresAt?: string
  }>(),
  {
    orderType: 'purchase',
    userSubscriptionId: 0,
    subscriptionPeriodEnd: '',
    subscriptionExpiresAt: '',
  },
)

const emit = defineEmits<{
  'update:open': [value: boolean]
  paid: []
}>()

const quantity = ref(1)
const qtyPresets = [1, 5, 10, 20]
const couponInput = ref('')
const couponApplied = ref<{ code: string; percentOff: number } | null>(null)
const paying = ref(false)

/** 仅新购可选份数；升档周期由订阅 period_end / expires_at 自动计算 */
const showQuantityPicker = computed(() => props.orderType === 'purchase')

const showTotalSub = computed(() =>
  ['purchase', 'renewal', 'upgrade', 'quota_addon'].includes(props.orderType),
)

const totalSubUnit = computed(() => {
  switch (props.orderType) {
    case 'upgrade':
      return '个周期'
    case 'renewal':
    case 'quota_addon':
      return '份'
    default:
      return '份'
  }
})

const billingPeriodUnit = computed(() => {
  const p = props.product?.billing_period
  if (p === 'year') return '年'
  if (p === 'once') return '次'
  return '月'
})

const upgradeBillingCycles = computed(() => {
  if (props.orderType !== 'upgrade' || !props.product) return 1
  return countUpgradeBillingCycles(
    props.subscriptionPeriodEnd ?? '',
    props.subscriptionExpiresAt ?? '',
    props.product.billing_period ?? 'month',
  )
})

const checkoutQuantity = computed(() =>
  props.orderType === 'upgrade' ? upgradeBillingCycles.value : quantity.value,
)

const upgradeExpiresLabel = computed(() => {
  if (props.orderType !== 'upgrade') return ''
  const exp = parseSubscriptionDate(props.subscriptionExpiresAt ?? '')
  return exp ? formatSubscriptionDate(exp) : props.subscriptionExpiresAt
})

watch(
  () => props.open,
  (visible) => {
    if (visible) {
      quantity.value = 1
      couponInput.value = ''
      couponApplied.value = null
    }
  },
)

watch(
  () => props.orderType,
  () => {
    if (!showQuantityPicker.value) {
      quantity.value = 1
    }
  },
)

const dialogTitle = computed(() => {
  const title = props.product?.card_title ?? '套餐'
  switch (props.orderType) {
    case 'renewal':
      return `续购 - ${title}`
    case 'quota_addon':
      return `加本期额度 - ${title}`
    case 'upgrade':
      return `升档 - ${title}`
    default:
      return title
  }
})

const planLabel = computed(() => {
  const p = props.product
  if (!p) return '套餐'
  if (p.billing_period === 'month') return '月卡'
  if (p.billing_period === 'year') return '年卡'
  return '单次'
})

const unitPrice = computed(() => parseMoney(props.product?.price))

const subtotalAmount = computed(() => unitPrice.value * checkoutQuantity.value)

const couponOffAmount = computed(() => {
  if (!couponApplied.value) return 0
  return Math.round(subtotalAmount.value * (couponApplied.value.percentOff / 100) * 100) / 100
})

const totalAmount = computed(() =>
  Math.max(0, Math.round((subtotalAmount.value - couponOffAmount.value) * 100) / 100),
)

const quotaAddonGrantTokens = computed(() => {
  if (props.orderType !== 'quota_addon' || !props.product) return 0
  return props.product.limit_tokens * quantity.value
})

function close() {
  emit('update:open', false)
}

function clampQty(n: number) {
  quantity.value = Math.min(99, Math.max(1, n))
}

function applyCoupon() {
  const code = couponInput.value.trim()
  if (!code) {
    couponApplied.value = null
    return
  }
  if (code.toUpperCase() === 'AIPlan10') {
    couponApplied.value = { code, percentOff: 10 }
    ElMessage.success('优惠券已应用：9 折')
    return
  }
  ElMessage.warning('优惠券无效')
  couponApplied.value = null
}

async function submitPay(channel: 'alipay' | 'paypal') {
  if (!props.product) {
    ElMessage.warning('请选择套餐')
    return
  }
  if (!props.user) {
    ElMessage.warning('请先登录后再支付')
    return
  }
  paying.value = true
  try {
    const body: Parameters<typeof orderApi.create>[0] = {
      product_id: props.product.id,
      order_type: props.orderType,
      quantity: checkoutQuantity.value,
      channel,
      enterprise_invoice: false,
    }
    if (props.userSubscriptionId > 0) {
      body.user_subscription_id = props.userSubscriptionId
    }
    const created = await orderApi.checkout(body)
    if (created.status !== 'completed') {
      ElMessage.warning(`订单 ${created.order_no} 已创建，但支付未完成，请稍后重试或联系客服`)
      return
    }
    ElMessage.success(
      `${channel === 'alipay' ? '支付宝' : 'PayPal'} 支付成功 · 订单 ${created.order_no}`,
    )
    emit('paid')
    close()
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '下单失败')
  } finally {
    paying.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="purchase-fade">
      <div v-if="open && product" class="purchase-overlay" @click.self="close">
        <div class="purchase-dialog" role="dialog" aria-modal="true" :aria-label="dialogTitle">
          <header class="purchase-head">
            <h2 class="purchase-title">{{ dialogTitle }}</h2>
            <button type="button" class="purchase-close" aria-label="关闭" @click="close">×</button>
          </header>

          <div class="purchase-plan">
            <label class="plan-row">
              <span class="plan-radio" aria-hidden="true" />
              <span class="plan-name">{{ planLabel }}</span>
              <span class="plan-stock">库存充足</span>
              <span class="plan-price">{{ formatCny(product.price) }}</span>
            </label>
          </div>

          <div v-if="showQuantityPicker" class="purchase-qty">
            <span class="field-label">购买数量</span>
            <div class="qty-row">
              <div class="qty-stepper">
                <button type="button" aria-label="减少" @click="clampQty(quantity - 1)">−</button>
                <input
                  v-model.number="quantity"
                  type="number"
                  min="1"
                  max="99"
                  @change="clampQty(quantity)"
                />
                <button type="button" aria-label="增加" @click="clampQty(quantity + 1)">+</button>
              </div>
              <div class="qty-presets">
                <button
                  v-for="n in qtyPresets"
                  :key="n"
                  type="button"
                  class="qty-pill"
                  :class="{ active: quantity === n }"
                  @click="clampQty(n)"
                >
                  {{ n }}
                </button>
              </div>
            </div>
          </div>

          <div class="purchase-coupon">
            <input
              v-model="couponInput"
              type="text"
              placeholder="优惠券代码（可选）"
              @keyup.enter="applyCoupon"
            />
            <button type="button" class="coupon-apply" @click="applyCoupon">应用</button>
          </div>

          <div class="purchase-total">
            <p v-if="orderType === 'quota_addon' && quotaAddonGrantTokens > 0" class="quota-grant-line">
              新增额度 {{ formatTokenCount(quotaAddonGrantTokens) }} tokens
            </p>
            <p v-if="orderType === 'upgrade'" class="upgrade-expires-line">
              含 {{ upgradeBillingCycles }} 个{{ billingPeriodUnit }}计费周期（含当前周期）· 升档后服务到期
              {{ upgradeExpiresLabel }}
            </p>
            <p class="total-line">
              合计
              <strong>{{ formatCny(totalAmount) }}</strong>
            </p>
            <p v-if="showTotalSub" class="total-sub">
              {{ checkoutQuantity }} {{ totalSubUnit }} ×
              {{ formatCny(unitPrice) }}
              <template v-if="couponApplied"> · 优惠券 −{{ couponApplied.percentOff }}%</template>
            </p>
          </div>

          <div class="purchase-pay">
            <button
              type="button"
              class="pay-alipay"
              :disabled="paying"
              @click="submitPay('alipay')"
            >
              支付宝扫码支付
            </button>
            <button
              type="button"
              class="pay-paypal"
              :disabled="paying"
              @click="submitPay('paypal')"
            >
              PayPal
            </button>
          </div>

          <p class="purchase-hint">
            付款后约 1 分钟自动开通；订单与额度可在
            <RouterLink to="/login">会员中心</RouterLink>
            查看。
          </p>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.purchase-overlay {
  position: fixed;
  inset: 0;
  z-index: 2100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(15, 23, 42, 0.45);
  backdrop-filter: blur(4px);
}

.purchase-dialog {
  width: 100%;
  max-width: 520px;
  max-height: min(92vh, 720px);
  overflow-y: auto;
  padding: 28px 28px 24px;
  background: #fff;
  border-radius: 20px;
  box-shadow: 0 24px 64px rgba(30, 27, 75, 0.18);
}

.purchase-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 20px;
}

.purchase-title {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.purchase-close {
  flex-shrink: 0;
  width: 36px;
  height: 36px;
  font-size: 24px;
  line-height: 1;
  color: #94a3b8;
  cursor: pointer;
  background: #f8fafc;
  border: none;
  border-radius: 10px;
}

.purchase-close:hover {
  color: var(--atm-text);
  background: #f1f5f9;
}

.purchase-plan {
  margin-bottom: 22px;
}

.plan-row {
  display: grid;
  grid-template-columns: auto 1fr auto auto;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  cursor: default;
  border: 2px solid var(--atm-primary);
  border-radius: 14px;
  background: #faf5ff;
}

.plan-radio {
  width: 18px;
  height: 18px;
  border: 5px solid var(--atm-primary);
  border-radius: 50%;
  box-sizing: border-box;
}

.plan-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--atm-text);
}

.plan-stock {
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 600;
  color: #15803d;
  background: #dcfce7;
  border-radius: 999px;
}

.plan-price {
  font-size: 18px;
  font-weight: 800;
  color: var(--atm-primary);
}

.field-label {
  display: block;
  margin-bottom: 10px;
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-text);
}

.purchase-qty {
  margin-bottom: 20px;
}

.qty-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}

.qty-stepper {
  display: flex;
  align-items: center;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.qty-stepper button {
  width: 40px;
  height: 40px;
  font-size: 18px;
  color: var(--atm-text);
  cursor: pointer;
  background: #f8fafc;
  border: none;
}

.qty-stepper input {
  width: 48px;
  height: 40px;
  font-size: 15px;
  font-weight: 600;
  text-align: center;
  border: none;
  border-left: 1px solid #e2e8f0;
  border-right: 1px solid #e2e8f0;
  -moz-appearance: textfield;
}

.qty-stepper input::-webkit-outer-spin-button,
.qty-stepper input::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.qty-presets {
  display: flex;
  gap: 8px;
}

.qty-pill {
  min-width: 40px;
  padding: 8px 14px;
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-text-muted);
  cursor: pointer;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 999px;
}

.qty-pill.active {
  color: #fff;
  background: var(--atm-gradient);
  border-color: transparent;
}

.purchase-coupon {
  display: flex;
  gap: 0;
  margin-bottom: 20px;
  overflow: hidden;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.purchase-coupon input {
  flex: 1;
  min-width: 0;
  padding: 12px 14px;
  font-size: 14px;
  border: none;
  outline: none;
}

.coupon-apply {
  padding: 0 18px;
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-primary);
  cursor: pointer;
  background: #f5f3ff;
  border: none;
  border-left: 1px solid #e2e8f0;
}

.purchase-total {
  margin-bottom: 20px;
}

.quota-grant-line,
.upgrade-expires-line {
  margin: 0 0 10px;
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-text);
}

.total-line {
  margin: 0 0 4px;
  font-size: 15px;
  color: var(--atm-text-muted);
}

.total-line strong {
  margin-left: 8px;
  font-size: 26px;
  font-weight: 800;
  color: var(--atm-text);
}

.total-sub {
  margin: 0;
  font-size: 13px;
  color: #94a3b8;
}

.purchase-pay {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 16px;
}

.pay-alipay {
  padding: 14px 12px;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  background: var(--atm-gradient);
  border: none;
  border-radius: 999px;
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.35);
}

.pay-alipay:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.pay-paypal {
  padding: 14px 12px;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  cursor: pointer;
  background: #0f172a;
  border: none;
  border-radius: 999px;
}

.pay-paypal:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.purchase-hint {
  margin: 0;
  font-size: 12px;
  line-height: 1.55;
  color: #94a3b8;
  text-align: center;
}

.purchase-hint a {
  color: var(--atm-primary);
  font-weight: 500;
}

.purchase-fade-enter-active,
.purchase-fade-leave-active {
  transition: opacity 0.2s ease;
}

.purchase-fade-enter-from,
.purchase-fade-leave-to {
  opacity: 0;
}
</style>
