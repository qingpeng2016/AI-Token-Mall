/** 登录态 token（Cookie，开发阶段 mock，不请求后端） */
export const AUTH_TOKEN_COOKIE = 'atm_token'

const MOCK_TOKEN = 'dev_local_session'
const MAX_AGE_SEC = 60 * 60 * 24 * 30

export function getAuthToken(): string | null {
  if (typeof document === 'undefined') return null
  const match = document.cookie.match(
    new RegExp(`(?:^|;\\s*)${AUTH_TOKEN_COOKIE}=([^;]*)`),
  )
  return match ? decodeURIComponent(match[1]) : null
}

export function setAuthToken(value: string = MOCK_TOKEN): void {
  document.cookie = `${AUTH_TOKEN_COOKIE}=${encodeURIComponent(value)}; Path=/; Max-Age=${MAX_AGE_SEC}; SameSite=Lax`
}

export function clearAuthToken(): void {
  document.cookie = `${AUTH_TOKEN_COOKIE}=; Path=/; Max-Age=0; SameSite=Lax`
}

export function hasAuthToken(): boolean {
  return Boolean(getAuthToken())
}
