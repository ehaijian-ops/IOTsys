import { defineStore } from 'pinia'
import { ref } from 'vue'
import http from '@/api/index'

export interface UserInfo {
  id: number
  username: string
  nickname: string
  role: string
  email?: string
  phone?: string
  avatar?: string
  enabled: boolean
  last_login_at?: string
  created_at: string
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
  user: UserInfo
}

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('access_token') || '')
  const user = ref<UserInfo | null>(null)

  const roleLabels: Record<string, string> = {
    super_admin: '超级管理员',
    admin: '管理员',
    operator: '运维人员',
    viewer: '查看者',
  }

  function setToken(t: string) {
    token.value = t
    localStorage.setItem('access_token', t)
  }

  function clearToken() {
    token.value = ''
    localStorage.removeItem('access_token')
  }

  function isLoggedIn(): boolean {
    return !!token.value
  }

  function getRoleLabel(role: string): string {
    return roleLabels[role] || role
  }

  async function login(username: string, password: string): Promise<LoginResponse> {
    const resp = await http.post('/auth/login', { username, password }) as any
    const data = resp.data as LoginResponse
    setToken(data.access_token)
    user.value = data.user
    return data
  }

  async function fetchUserInfo(): Promise<UserInfo> {
    const resp = await http.get('/auth/userinfo') as any
    user.value = resp.data as UserInfo
    return user.value!
  }

  function logout() {
    clearToken()
    user.value = null
  }

  return {
    token,
    user,
    roleLabels,
    setToken,
    clearToken,
    isLoggedIn,
    getRoleLabel,
    login,
    fetchUserInfo,
    logout,
  }
})
