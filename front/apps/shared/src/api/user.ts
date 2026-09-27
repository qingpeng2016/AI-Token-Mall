import type { ApiEnvelope, LoginRequest, RegisterRequest, UserProfile } from '../types/user'
import { createHttpClient, type HttpClientOptions } from './http'

export function createUserApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    register(body: RegisterRequest) {
      return http.post<ApiEnvelope<UserProfile>>('/api/v1/users/register', body)
    },
    login(body: LoginRequest) {
      return http.post<ApiEnvelope<UserProfile>>('/api/v1/users/login', body)
    },
  }
}
