export type NavPlanItem = {
  label: string
  price: string
  href: string
  featured?: boolean
}

export type NavMegaColumn = {
  title: string
  items: NavPlanItem[]
}

/** 套餐购买 — 大面板（对齐 ProPlus 结构，价格取自当前 mock） */
export const navMegaMenu: NavMegaColumn[] = [
  {
    title: 'ChatGPT',
    items: [
      { label: 'ChatGPT Plus', price: '¥178/月', href: '#catalog', featured: true },
      { label: 'GPT Go', price: '¥89/月', href: '#catalog' },
      { label: 'Pro 5X', price: '¥899/月', href: '#catalog' },
      { label: 'Pro 20X', price: '¥1799/月', href: '#catalog' },
      { label: 'Codex', price: '¥599/月', href: '#catalog' },
      { label: 'Images 2.5', price: '¥998/月', href: '#catalog' },
    ],
  },
  {
    title: 'Claude',
    items: [
      { label: 'Claude Pro', price: '¥219/月', href: '#catalog' },
      { label: 'Claude Max 5X', price: '¥999/月', href: '#catalog' },
      { label: 'Claude Max 20X', price: '¥1899/月', href: '#catalog' },
      { label: 'Claude Code', price: '¥599/月', href: '#catalog' },
    ],
  },
  {
    title: 'Grok',
    items: [
      { label: 'SuperGrok', price: '¥790/季', href: '#catalog' },
      { label: 'SuperGrok Heavy', price: '¥790/季', href: '#catalog' },
      { label: 'Grok Imagine', price: '¥590/季', href: '#catalog' },
    ],
  },
  {
    title: 'Gemini',
    items: [
      { label: 'Google AI Pro', price: '¥189/月', href: '#catalog' },
      { label: 'Google AI Ultra', price: '¥1800/月', href: '#catalog' },
      { label: 'Nano Banana / Veo', price: '¥199/月', href: '#catalog' },
    ],
  },
  {
    title: 'Cursor',
    items: [
      { label: 'Cursor Pro', price: '¥198/月', href: '#catalog' },
      { label: 'Cursor Pro+', price: '¥568/月', href: '#catalog' },
      { label: 'Cursor Ultra', price: '¥1280/月', href: '#catalog' },
      { label: 'Cloud Agents', price: '¥998/月', href: '#catalog' },
    ],
  },
]

export type NavDropItem = {
  label: string
  href: string
  price?: string
  featured?: boolean
}

export type NavMenuDropdown = {
  id: string
  label: string
  hotSale?: boolean
  items: NavDropItem[]
}

export type NavMenuPlainLink = {
  id: string
  label: string
  plainLink: true
  to: string
}

export type NavMenuEntry = NavMenuDropdown | NavMenuPlainLink

export function isNavDropdown(entry: NavMenuEntry): entry is NavMenuDropdown {
  return !('plainLink' in entry)
}

export const navBrandDropdowns: NavMenuEntry[] = [
  {
    id: 'chatgpt',
    label: 'ChatGPT',
    items: navMegaMenu[0].items.map((i) => ({
      label: i.label,
      href: i.href,
      price: i.price,
      featured: i.featured,
    })),
  },
  {
    id: 'claude',
    label: 'Claude',
    items: navMegaMenu[1].items.map((i) => ({
      label: i.label,
      href: i.href,
      price: i.price,
    })),
  },
  {
    id: 'cursor',
    label: 'Cursor',
    items: navMegaMenu[4].items.map((i) => ({
      label: i.label,
      href: i.href,
      price: i.price,
    })),
  },
  {
    id: 'tools',
    label: '工具',
    items: [
      { label: '工具中心', href: '#' },
      { label: '价格中心', href: '#catalog' },
      { label: '使用说明', href: '#' },
    ],
  },
  {
    id: 'enterprise',
    label: '企业采购',
    plainLink: true,
    to: '/#faq',
  },
]
