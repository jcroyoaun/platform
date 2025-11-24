import apiClient from './client'

export interface User {
  id: number
  email: string
  email_verified: boolean
  api_key?: string
}

export interface LoginRequest {
  email: string
  password: string
}

export interface SignupRequest {
  email: string
  password: string
}

export const authAPI = {
  async login(data: LoginRequest) {
    return apiClient.post('/api/auth/login', data)
  },

  async signup(data: SignupRequest) {
    return apiClient.post('/api/auth/signup', data)
  },

  async logout() {
    return apiClient.post('/api/auth/logout')
  },

  async getCurrentUser(): Promise<User | null> {
    try {
      const response = await apiClient.get('/api/auth/me')
      return response.data
    } catch (error) {
      return null
    }
  },

  async forgottenPassword(email: string) {
    return apiClient.post('/api/auth/forgotten-password', { email })
  },

  async resetPassword(token: string, newPassword: string) {
    return apiClient.post(`/api/auth/password-reset/${token}`, { new_password: newPassword })
  },

  async verifyEmail(token: string) {
    return apiClient.get(`/api/auth/verify-email/${token}`)
  },

  async resendVerificationEmail() {
    return apiClient.post('/api/auth/resend-verification')
  },

  async generateAPIKey() {
    return apiClient.post('/api/auth/api-key')
  },
}
