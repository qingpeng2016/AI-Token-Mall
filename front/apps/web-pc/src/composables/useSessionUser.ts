import type { UserProfile } from '@ai-token-mall/shared'
import { clearAuthToken, getAuthToken, hasAuthToken } from '@/utils/auth-cookie'

const STORAGE_KEY = 'atm_user'
export const PENDING_BUY_KEY = 'atm_pending_buy'

export function getSessionUser(): UserProfile | null {
  if (!hasAuthToken()) return null
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    return JSON.parse(raw) as UserProfile
  } catch {
    return null
  }
}

export function isLoggedIn(): boolean {
  return hasAuthToken()
}

export function setSessionUser(profile: UserProfile): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(profile))
}

/** 退出登录：清空 Cookie 中的 atm_token，并移除本地用户信息 */
export function clearSessionUser(): void {
  clearAuthToken()
  localStorage.removeItem(STORAGE_KEY)
}

export function userAccountLabel(user: UserProfile): string {
  if (user.email) return user.email
  if (user.phone) return user.phone
  if (user.nickname) return user.nickname
  return `用户 #${user.id}`
}

/** 供调试：当前 cookie token */
export function sessionToken(): string | null {
  return getAuthToken()
}
