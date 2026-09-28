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

export type NavBrandMenu = {
  id: string
  label: string
  hot_sale?: boolean
  items: NavPlanItem[]
}

export type NavMenuResponse = {
  mega_menu: NavMegaColumn[]
  brand_menus: NavBrandMenu[]
}
