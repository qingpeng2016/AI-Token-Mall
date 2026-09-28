import { ref, type Ref } from 'vue'

const routeNavigationLoading = ref(false)

export function useRouteNavigationLoading(): { active: Ref<boolean> } {
  return { active: routeNavigationLoading }
}

export function setRouteNavigationLoading(active: boolean): void {
  routeNavigationLoading.value = active
}
