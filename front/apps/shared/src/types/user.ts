/** 与后端 domain/rest 对齐，随 API 演进 */

export interface UserProfile {
  id: number
  email?: string | null
  phone?: string | null
  nickname?: string | null
  status?: string
  created_at?: string
}

export interface RegisterRequest {
  email?: string
  phone?: string
  password: string
  confirm_password: string
}

export interface LoginRequest {
  email?: string
  phone?: string
  password: string
}

export interface LoginResponse {
  user: UserProfile
}

/** 与 common/dederi/gin/response.ApiResp 一致 */
export interface ApiEnvelope<T> {
  code: number
  message: string
  data: T
}
