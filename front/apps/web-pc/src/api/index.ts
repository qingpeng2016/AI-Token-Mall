import { createUserApi } from '@ai-token-mall/shared'

const baseURL = import.meta.env.VITE_API_BASE_URL ?? ''

export const userApi = createUserApi({
  baseURL,
  getToken: () => localStorage.getItem('atm_token'),
})
