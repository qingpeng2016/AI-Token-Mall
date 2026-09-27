import type {
  ApiEnvelope,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  UserProfile,
} from '../types/user'
import { createHttpClient, type HttpClientOptions } from './http'

function unwrap<T>(envelope: ApiEnvelope<T>): T {
  return envelope.data
}

export function createUserApi(options: HttpClientOptions) {
  const http = createHttpClient(options)

  return {
    async register(body: RegisterRequest): Promise<UserProfile> {
      const res = await http.post<ApiEnvelope<UserProfile>>('/api/v1/users/register', body)
      return unwrap(res)
    },
    async login(body: LoginRequest): Promise<LoginResponse> {
      const res = await http.post<ApiEnvelope<LoginResponse>>('/api/v1/users/login', body)
      return unwrap(res)
    },
  }
}
